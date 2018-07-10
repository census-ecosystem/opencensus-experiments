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
//
// This program shows how to send requests of registration and sending data to the raspberry Pi based on the protocols.

#include <ArduinoJson.h>

#define OK 200
#define FAIL 404
#define BUFFER_SIZE 256
#define JSON_BUFFER_SIZE 200

void setup() {
  // Initialize Serial port
  Serial.begin(9600);
  while (!Serial) continue;
}

/*
 * Arduino would keep sending record requests to the Pi.
 */
void loop() {
   request(sendData);
}

/*
 * The template for sending requests to the raspberry Pi.
 * It would keep sending the commands until receive OK.
 */
void request(void (*func)()) {
  int code;
  do {
    JsonObject* response;
    do {
      // TODO: Currently we hard-code the argument in the code. Would handle the problem that how to solve
      // multiple arguments.
      func();
      char buffer[BUFFER_SIZE];
      readLine(buffer, BUFFER_SIZE);
      response = parseResponse(buffer);
    } while (response == NULL);

    code = (*response)["Code"];
    const char *info = (*response)["Info"];

    //Serial.print("Receive Response: Code ");
    //Serial.print(code);
    //Serial.print(" info: ");
    //Serial.println(info);
    if (code != OK){
      // If it receives a negative response, it would delay for one second
      delay(1000);
    }
  } while (code != OK);
}

/*
 * Implement the readLine in arduino. It would block until receive whitespace character.
 */
void readLine(char * buffer, int maxLength)
{
  uint8_t idx = 0;
  char c;
  do
  {
    while (Serial.available() == 0) ; // wait for a char this causes the blocking
    // TODO: In reality, it should set a timeout since both ends might wait for each other.
    c = Serial.read();
    //Serial.print(c);
    buffer[idx++] = c;
  } while (c != '\n' && c != '\r' && idx < maxLength - 1);
  // To sure it would not overflow. The last character has to be a \0
  buffer[idx] = 0;
}


JsonObject* parseResponse(char *json) {
  // Memory pool for JSON object tree.
  //
  // Inside the brackets, 200 is the size of the pool in bytes.
  // Don't forget to change this value to match your JSON document.
  // Use arduinojson.org/assistant to compute the capacity.
  StaticJsonBuffer<200> jsonBuffer;

  // StaticJsonBuffer allocates memory on the stack, it can be
  // replaced by DynamicJsonBuffer which allocates in the heap.
  //
  // DynamicJsonBuffer  jsonBuffer(200);
  JsonObject& root = jsonBuffer.parseObject(json);

  // Test if parsing succeeds.
  if (!root.success()) {
    //Serial.println("parseObject() failed");
    return NULL;
  } else{
    return &root;
  }
}

void sendData() {
  // Memory pool for JSON object tree.
  //
  // Inside the brackets, 200 is the size of the pool in bytes.
  // Don't forget to change this value to match your JSON document.
  // Use arduinojson.org/assistant to compute the capacity.
  StaticJsonBuffer<JSON_BUFFER_SIZE> jsonBuffer;

  // StaticJsonBuffer allocates memory on the stack, it can be
  // replaced by DynamicJsonBuffer which allocates in the heap.
  //
  // DynamicJsonBuffer  jsonBuffer(200);

  // Create the root of the object tree.
  //
  // It's a reference to the JsonObject, the actual bytes are inside the
  // JsonBuffer with all the other nodes of the object tree.
  // Memory is freed when jsonBuffer goes out of scope.
  JsonObject& root = jsonBuffer.createObject();
  JsonObject& measure = root.createNestedObject("Measure");
  measure["Name"] = "my.org/measure/Measure_Test";
  measure["Measurement"] = "9";

  JsonObject& tagPairs = root.createNestedObject("Tag");
  tagPairs["DeviceId"] = "Arduino-1";
  tagPairs["SampleDate"] = "2018-07-02";

  root.printTo(Serial);

  Serial.println();

  Serial.flush();
}