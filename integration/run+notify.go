// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/keighl/postmark"
	"github.com/orijtech/infra"
)

var infraClient *infra.Client
var emailClient *postmark.Client
var emails []string
var project, bucket string
var appEmail string
var postmarkAccountToken, postmarkServerToken, emailStr string

func init() {
	flag.StringVar(&project, "gcp-project", "opencensus-integration-tests", "the GCP project that'll be used to save information about builds")
	flag.StringVar(&bucket, "gcs-bucket", "opencensus-integration-tests", "the GCS bucket that failure results will be saved to")
	flag.StringVar(&emailStr, "notify-emails", "emmanuel@orijtech.com", "the emails of folks to notify on failures e.g. foo@example.org,bar@foo.com")
	flag.StringVar(&appEmail, "app-email", "emmanuel@orijtech.com", "the postmark email account to use to send notifications")
	flag.StringVar(&postmarkAccountToken, "postmark-account-token", "", "the postmark account token")
	flag.StringVar(&postmarkServerToken, "postmark-server-token", "", "the postmark server token")
	flag.Parse()

	var err error
	infraClient, err = infra.NewDefaultClient()
	if err != nil {
		log.Fatalf("Creating GCP infraClient %v", err)
	}
	emailClient = postmark.NewClient(postmarkServerToken, postmarkAccountToken)
	emails = strings.Split(emailStr, ",")
}

type grouping struct {
	cmd   string
	steps [][]string
}

var preliminaryCommands = []*grouping{
	{
		cmd: "go",
		steps: [][]string{
			// {"get", "-u", "go.opencensus.io/..."},
		},
	},
}

func runCommand(maxTimeSeconds int64, sCmd string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxTimeSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, sCmd, args...)
	return cmd.CombinedOutput()
}

func main() {
	// 1. Run the preliminary installations
	for _, pCmd := range preliminaryCommands {
		for _, args := range pCmd.steps {
			output, err := runCommand(100, pCmd.cmd, args...)
			if err != nil {
				// Upload the results to GCS then mail everyone the link to the test failures.
				intro := fmt.Sprintf("While running `%s %s` encountered", pCmd.cmd, strings.Join(args, " "))
				if e := uploadToGCSThenMailWithLink(intro, output, err, emails...); e != nil {
					msg := "While trying to notify concerened parties, got error %v\n\nOriginal error: %v\nOutput: %s\n"
					log.Fatalf(msg, e, err, output)
				}
				log.Fatalf("Encountered error %v so proceeding to notify concerned parties by email", err)
			}
		}
	}

	// 2. Run `make test`
	output, err := runCommand(600, "make", "test")
	if err != nil {
		// Mail these results to the team
		intro := "While running `make test` encountered"
		if e := uploadToGCSThenMailWithLink(intro, output, err, emails...); e != nil {
			msg := "While trying to notify concerened parties after `make test`, got error %v\n\nOriginal error: %v\nOutput: %s\n"
			log.Fatalf(msg, e, err, output)
		}
		log.Fatalf("Failed to run `make test` error: %v\noutput: %s", err, output)
	}
}

func uploadToGCSThenMailWithLink(intro string, output []byte, err error, emails ...string) error {
	// 1. Ensure that the bucket exists firstly on GCS
	bh := &infra.BucketCheck{Project: project, Bucket: bucket}
	if _, err := infraClient.EnsureBucketExists(bh); err != nil {
		return err
	}

	uploadBody := fmt.Sprintf("%s\nError %v\n\nOutput: %s", intro, err, output)

	now := time.Now().Round(time.Second)
	mHash := md5.New()
	fmt.Fprintf(mHash, "%s", now)
	// 2. Upload the content to the GCS bucket
	params := &infra.UploadParams{
		Public: true,
		Reader: func() io.Reader { return strings.NewReader(uploadBody) },
		Bucket: bucket,
		Name:   fmt.Sprintf("test-failures/%d/%d/%d/%x.txt", now.Year(), now.Month(), now.Day(), mHash.Sum(nil)),
	}
	obj, err := infraClient.UploadWithParams(params)
	if err != nil {
		return err
	}
	theURL := infra.ObjectURL(obj)

	htmlBuf := new(bytes.Buffer)
	info := map[string]string{"URL": theURL, "Time": fmt.Sprintf("%s", now.UTC())}
	if err := emailTemplate.Execute(htmlBuf, info); err != nil {
		return err
	}
	email := postmark.Email{
		From:     appEmail,
		To:       strings.Join(emails, ","),
		Subject:  "OpenCensus integration test failures",
		HtmlBody: htmlBuf.String(),
	}
	_, err = emailClient.SendEmail(email)
	return err
}

var emailTemplate = template.Must(template.New("notification").Parse(`
Test failures for OpenCensus integration tests at {{.Time}}.<br />Please visit the URL {{.URL}}
`))
