#include <Arduino.h>
#include <SPI.h>
#include <Wire.h>

#include "cc1101.h"
#include "LowPower.h"
#include "SHT31D.h"
#include "spi_transaction.h"

static SHT31D sensor;
static CC1101 radio;

void setup() {
  Serial.begin(115200);

  Wire.begin();
  SPI.begin();

  sensor.Init();
  radio.Reset();

  delay(100);
}

void Send(uint16_t temp, uint16_t humidity) {
  byte packet[] = {byte(temp >> 8), byte(temp & 0xff), byte(humidity >> 8), byte(humidity & 0xff)};
  radio.SendPacket(packet, sizeof(packet));
}

void loop() {
  Serial.write("Hello world!\n");

  SHT31D::Reading reading = sensor.ReadTemperatureAndHumidity();
  Serial.print(reading.temperature);
  Serial.println();
  Serial.print(reading.humidity);
  Serial.println();

  uint16_t temp = uint16_t(reading.temperature * 100.0f);
  uint16_t humidity = uint16_t(reading.humidity * 100.0f);

  radio.Wake();

  Send(temp, humidity);

  radio.Sleep();

  // Sleep for 32s in low power mode.
  LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
  LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
  LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
  LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
}