/*
 * Copyright 2018, OpenCensus Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.opencensus.interop.http;

import com.google.common.collect.ImmutableList;
import com.google.common.collect.ImmutableMap;
import java.util.List;
import java.util.logging.Logger;

/** Util methods and constants. */
final class HttpInteropTestUtils {

  private static Logger logger = Logger.getLogger(HttpInteropTestUtils.class.getName());

  static final String ENV_PORT_KEY_GO = "OPENCENSUS_GO_HTTP_INTEGRATION_TEST_SERVER_ADDR";
  static final String ENV_PORT_KEY_JAVA = "OPENCENSUS_JAVA_HTTP_INTEGRATION_TEST_SERVER_ADDR";
  static final int DEFAULT_PORT_GO = 9900;
  static final int DEFAULT_PORT_JAVA = 9901;
  static final ImmutableMap<String, Setup> SETUP_MAP =
      ImmutableMap.of(
          "Go",
          new Setup(ENV_PORT_KEY_GO, DEFAULT_PORT_GO, ImmutableList.of("b3", "google")),
          "Java",
          new Setup(ENV_PORT_KEY_JAVA, DEFAULT_PORT_JAVA, ImmutableList.of("b3", "google")));

  static class Setup {
    String serverPortKey;
    int defaultServerPort;
    List<String> propagators;

    public Setup(String serverPortKey, int defaultServerPort, List<String> propagators) {
      this.defaultServerPort = defaultServerPort;
      this.serverPortKey = serverPortKey;
      this.propagators = propagators;
    }
  }

  private HttpInteropTestUtils() {}
}
