/**
 * Blink
 *
 * Turns on an LED on for one second,
 * then off for one second, repeatedly.
 */
#include "Arduino.h"

#include <SPI.h>

const int kSlaveSelectPin = 10;

// Set LED_BUILTIN if it is not defined by Arduino framework
// #define LED_BUILTIN 13

class SPITransaction {
 public:
  SPITransaction() {
    digitalWrite(kSlaveSelectPin, LOW);
    SPI.beginTransaction(SPISettings(50000, MSBFIRST, SPI_MODE0));
  }

  ~SPITransaction() {
    digitalWrite(kSlaveSelectPin, HIGH);
    SPI.endTransaction();
  }
};

void reset() {
  SPITransaction transaction;
  SPI.transfer(0x30);
}

byte readReg(byte addr) {
  SPITransaction transaction;

  SPI.transfer(addr | 0x80);
  byte reply = SPI.transfer(0x00);

  return reply;
}

void setup()
{
  // initialize LED digital pin as an output.
  pinMode(LED_BUILTIN, OUTPUT);

  // Initialise the Slave Select pin.
  pinMode(kSlaveSelectPin, OUTPUT);

  Serial.begin(9600);

  SPI.begin();
}

void loop()
{
  // turn the LED on (HIGH is the voltage level)
  digitalWrite(LED_BUILTIN, HIGH);

  // wait for a second
  delay(1000);

  // turn the LED off by making the voltage LOW
  digitalWrite(LED_BUILTIN, LOW);

  Serial.write("Hello world!\n");

   // wait for a second
  delay(1000);

  reset();

  Serial.println(readReg(0xf0));
  Serial.println(readReg(0xf1));
}