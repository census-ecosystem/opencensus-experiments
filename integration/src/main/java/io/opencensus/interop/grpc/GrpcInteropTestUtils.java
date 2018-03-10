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

package io.opencensus.interop.grpc;

import com.google.common.collect.ImmutableMap;
import java.util.logging.Logger;

/** Util methods and constants. */
final class GrpcInteropTestUtils {

  private static Logger logger = Logger.getLogger(GrpcInteropTestUtils.class.getName());

  static final String ENV_PORT_KEY_GO = "OPENCENSUS_GO_GRPC_INTEGRATION_TEST_SERVER_ADDR";
  static final String ENV_PORT_KEY_JAVA = "OPENCENSUS_JAVA_GRPC_INTEGRATION_TEST_SERVER_ADDR";
  static final int DEFAULT_PORT_GO = 9800;
  static final int DEFAULT_PORT_JAVA = 9801;
  static final ImmutableMap<String, Integer> SETUP_MAP =
      ImmutableMap.of(ENV_PORT_KEY_GO, DEFAULT_PORT_GO, ENV_PORT_KEY_JAVA, DEFAULT_PORT_JAVA);

  static int getPortOrDefault(String varName, int defaultPort) {
    int portNumber = defaultPort;
    String portStr = System.getenv(varName);
    if (portStr != null) {
      try {
        portNumber = Integer.parseInt(portStr);
      } catch (NumberFormatException e) {
        logger.warning(
            String.format("Port %s is invalid, use default port %d.", portStr, defaultPort));
      }
    }
    return portNumber;
  }

  private GrpcInteropTestUtils() {}
}
