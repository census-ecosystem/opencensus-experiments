/* Copyright 2018 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.opencensus.appengine.flex;

import static io.opencensus.appengine.flex.OpenCensusStatsUtil.CLIENT_METHOD;
import static java.nio.charset.StandardCharsets.UTF_8;

import com.google.cloud.storage.Blob;
import com.google.cloud.storage.BlobId;
import com.google.cloud.storage.BlobInfo;
import com.google.cloud.storage.Storage;
import com.google.cloud.storage.StorageOptions;

import io.opencensus.appengine.flex.OpenCensusStatsUtil.LatencyStatsRecorder;
import io.opencensus.appengine.flex.OpenCensusTraceUtil;
import io.opencensus.common.Scope;

import java.io.IOException;

import javax.servlet.ServletException;
import javax.servlet.annotation.WebServlet;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@SuppressWarnings("serial")
@WebServlet(name = "work", urlPatterns = {"/work"})
public class WorkServlet extends HttpServlet {
  private static final String BUCKET_NAME = "oc-appengine-flex-bucket-name";
  private static final String BLOB_NAME = "oc-appengine-flex-blob-name";
  private static final String CONTENT_STRING = "oc appengine flex blob";

  @Override
  public void doGet(HttpServletRequest req, HttpServletResponse resp)
      throws ServletException, IOException {
    if (OpenCensusTraceUtil.PROJECT_ID.isEmpty()) {
      resp.setContentType("text/plain");
      resp.getWriter().println("Work - empty project id.");
      return;
    }
    try (
        Scope statsScope = LatencyStatsRecorder.create(CLIENT_METHOD, "Work:Get");
        Scope traceScope = OpenCensusTraceUtil.createSpanBuilder("Work:Get").startScopedSpan()) {
      OpenCensusTraceUtil.current().addAnnotation("Work:Get:Begin");
      Storage storage = StorageOptions.getDefaultInstance().toBuilder().setProjectId(
          OpenCensusTraceUtil.PROJECT_ID).build().getService();
      BlobId blobId = BlobId.of(BUCKET_NAME, BLOB_NAME);
      try (Scope createScope = LatencyStatsRecorder.create(CLIENT_METHOD, "Work:Get:Create")) {
        OpenCensusTraceUtil.current().addAnnotation("Work:Get:Create");
        // Upload a blob.
        BlobInfo blobInfo = BlobInfo.newBuilder(blobId).setContentType("text/plain").build();
        Blob blob = storage.create(blobInfo, CONTENT_STRING.getBytes(UTF_8));
      }
      try (Scope readScope = LatencyStatsRecorder.create(CLIENT_METHOD,"Work:Get:Read")) {
        // Read a blob.
        OpenCensusTraceUtil.current().addAnnotation("Work:Get:Read");
        byte[] content = storage.readAllBytes(blobId);
        String contentString = new String(content, UTF_8);
        if (!contentString.equals(CONTENT_STRING)) {
          throw new RuntimeException("Invalid read after upload.");
        }
      }
      try (Scope deleteScope = LatencyStatsRecorder.create(CLIENT_METHOD, "Work:Get:Delete")) {
        // Delete a blob.
        OpenCensusTraceUtil.current().addAnnotation("Work:Get:Delete");
        storage.delete(blobId);
      }
      resp.setContentType("text/plain");
      resp.getWriter().println("Done work.");
    }
  }
}
// [END example]
