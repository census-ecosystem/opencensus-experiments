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

package testdata

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"
)

// LoadTestSuites returns the test suites parsed from the text files under this directory.
func LoadTestSuites() []*interop.TestRequest {
	var reqs []*interop.TestRequest
	for _, file := range getAllTxtFiles() {
		reqs = append(reqs, parseReq(file))
	}
	return reqs
}

func getAllTxtFiles() []string {
	var files []string

	root := "."
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".txt" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalln("Error getting file path:", err)
	}
	return files
}

func parseReq(filepath string) *interop.TestRequest {
	req := &interop.TestRequest{}

	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}

	if err := proto.Unmarshal(f, req); err != nil {
		log.Fatalln("Failed to parse TestRequest:", err)
	}

	return req
}
