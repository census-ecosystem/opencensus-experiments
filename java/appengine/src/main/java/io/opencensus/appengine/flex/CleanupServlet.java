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

import com.google.cloud.storage.BucketInfo;
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

// [START example]
@SuppressWarnings("serial")
// [START annotations]
@WebServlet(name = "cleanup", urlPatterns = {"/cleanup"})
// [END annotations]
public class CleanupServlet extends HttpServlet {
  private static final String BUCKET_NAME = "ocj-gae-bucket-name";

  @Override
  public void doGet(HttpServletRequest req, HttpServletResponse resp)
      throws IOException, ServletException {
    if (OpenCensusTraceUtil.PROJECT_ID.isEmpty()) {
      resp.setContentType("text/plain");
      resp.getWriter().println("Cleanup - empty project id.");
      return;
    }
    try (
        Scope statsScope = LatencyStatsRecorder.create(CLIENT_METHOD, "Cleanup:Get:Begin");
        Scope traceScope = OpenCensusTraceUtil.createSpanBuilder("Cleanup:Get").startScopedSpan()) {
      OpenCensusTraceUtil.current().addAnnotation("Cleanup:Get:Begin");
      Storage storage = StorageOptions.getDefaultInstance().toBuilder().setProjectId(
          OpenCensusTraceUtil.PROJECT_ID).build().getService();
      try (Scope deleteScope = LatencyStatsRecorder.create(CLIENT_METHOD, "Cleanup:Get::Delete")) {
        OpenCensusTraceUtil.current().addAnnotation("Cleanup:Get:Delete");
        if (storage.get(BUCKET_NAME) != null) {
          storage.get(BUCKET_NAME).delete();
        } else {
          OpenCensusTraceUtil.current().addAnnotation("Cleanup:Get:Delete:Null");
        }
      }
      OpenCensusTraceUtil.current().addAnnotation("Cleanup:Get:Response");
      resp.setContentType("text/plain");
      resp.getWriter().println("Done cleanup.");
    }
  }
}
// [END example]
