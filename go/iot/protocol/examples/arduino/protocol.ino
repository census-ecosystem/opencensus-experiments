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
//
// Hardware Connections (Breakoutboard to Arduino):
// -VCC = 3.3V
// -GND = GND
// -SDA = A4 (use inline 330 ohm resistor if your board is 5V)
// -SCL = A5 (use inline 330 ohm resistor if your board is 5V)

#include <ArduinoJson.h>

#define OK 200
#define FAIL 404
#define MEASUREUNREG 501
#define TAGUNREG 502
#define BUFFER_SIZE 256
#define JSON_BUFFER_SIZE 200
// The maximum backoff waiting time is 32 seconds
#define MAX_BACKOFF_TIME 32000
#include <Wire.h>
#include "SparkFunHTU21D.h"
#include "DHT.h"

//Create an instance of the object
//DHT dht;
HTU21D myHumidity;

void setup() {
  // Initialize Serial port
  Serial.begin(9600);
  myHumidity.begin();
  while (!Serial) continue;
  pinMode(LED_BUILTIN, OUTPUT);
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
  // For Arduino UNO, it only supports 4-byte integer.
  unsigned int failcounts = 0;
  unsigned int delaytime;
  // The outer loop is to keep sending requests to the slave until receiving a positive response
  do {
    JsonObject* response;
    // The inner loop is to receive the response from the slave end after sending the request
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
    switch (code){
      case FAIL:
        delaytime = (1 << failcounts) + random(1000);
        if (delaytime > MAX_BACKOFF_TIME){
          delaytime = MAX_BACKOFF_TIME;
          // When there are too many fail attempts, turn on the light to notify
          // Meanwhile stop increasing the fail attempts counter
          digitalWrite(LED_BUILTIN, HIGH);
        }
        Serial.println(delaytime);
        delay(delaytime);
        // In case of the overflow
        if (failcounts < 15)
          failcounts = failcounts + 1;
        break;
      case MEASUREUNREG:
        exit(-1);
        break;
      case OK:
        failcounts = 0;
        digitalWrite(LED_BUILTIN, LOW);
        break;
      case TAGUNREG:
        failcounts = 0;
        digitalWrite(LED_BUILTIN, LOW);
        break;
      default:
        exit(-1);
        //TODO
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
  root["Name"] = "opencensus.io/measure/Temperature";
  float temp = myHumidity.readTemperature();
  root["Measurement"] = String(temp);

  JsonObject& tagPairs = root.createNestedObject("Tag");
  tagPairs["ArduinoId"] = "Arduino-1";
  tagPairs["Date"] = "2018-07-02";

  root.printTo(Serial);

  Serial.println();

  Serial.flush();
}

