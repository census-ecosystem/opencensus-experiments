// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
//
// This program shows how to send requests of registration and sending data to the raspberry Pi based on the protocols.

#include <ArduinoJson.h>

#define OK 200
#define FAIL 404
void setup() {
  // Initialize Serial port
  Serial.begin(9600);
  while (!Serial) continue;
}

/*
 * It would first send request of registration to the raspberry Pi.
 * Once receive OK, it would keep sending collected data to the Pi.
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
      char buffer[256];
      readLine(buffer, 256);
      response = parseResponse(buffer);
    }
    while (response == NULL);

    code = (*response)["Code"];
    const char *info = (*response)["Info"];

    //Serial.print("Receive Response: Code ");
    //Serial.print(code);
    //Serial.print(" info: ");
    //Serial.println(info);
  }
  while (code != OK);
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
  }
  while (c != '\n' && c != '\r' && idx < maxLength - 1);
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
  }
  else
    return &root;
}

void sendData() {
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

  // Create the root of the object tree.
  //
  // It's a reference to the JsonObject, the actual bytes are inside the
  // JsonBuffer with all the other nodes of the object tree.
  // Memory is freed when jsonBuffer goes out of scope.
  JsonObject& root = jsonBuffer.createObject();
  JsonObject& measure = root.createNestedObject("Measure");
  measure["Name"] = "my.org/measure/Measure_Test";
  measure["MeasureType"] = "int64";
  measure["MeasureValue"] = "9";

  JsonArray& tagValues = root.createNestedArray("TagValues");
  tagValues.add("Arduino-1");
  tagValues.add("2018-07-02");

  root.printTo(Serial);

  Serial.println();

  Serial.flush();
}

