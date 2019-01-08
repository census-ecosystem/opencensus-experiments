/**
 * Copyright 2019, OpenCensus Authors
 *
 * Licensed under the Apache License, Version 2.0 (the 'License');
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

exports.constants = {
  HTTP_URL_ENDPOINT: '/test/request',
  PROTOBUF_HEADER: {'Content-Type': 'application/x-protobuf'},
  POST_METHOD: 'POST',
  DEFAULT_HOST: 'localhost'
};

function toBuffer (protobufObj) {
  const bytes = protobufObj.serializeBinary();
  const buffer = Buffer.from(bytes);
  return buffer;
}

function fromBuffer (buffer, protobufObj) {
  const bytes = new Uint8Array(Buffer.concat(buffer));
  const obj = protobufObj.deserializeBinary(bytes);
  return obj;
}

exports.toBuffer = toBuffer;
exports.fromBuffer = fromBuffer;
