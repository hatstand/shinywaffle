#include "SHT31D.h"

#include <Arduino.h>
#include <Wire.h>

void SHT31D::Init() {
  WriteCommand(kSoftReset);
  delay(10);
}

SHT31D::Reading::Reading(float temperature, float humidity)
  : temperature(temperature),
    humidity(humidity) {}

float SHT31D::ConvertTemperature(uint16_t raw_temp) {
  return -45 + 175 * (float(raw_temp) / 0xffff);
}

float SHT31D::ConvertHumidity(uint16_t raw_humidity) {
  return 100 * (float(raw_humidity) / 0xffff);
}

bool SHT31D::WriteCommand(uint16_t command) {
  Wire.beginTransmission(kSensorAddress);
  Wire.write(command >> 8);
  Wire.write(command & 0xff);
  return Wire.endTransmission() == 0;
}

SHT31D::Reading SHT31D::ReadTemperatureAndHumidity() {
  WriteCommand(kReadSensor);
  delay(500);
  Wire.requestFrom(kSensorAddress, uint8_t(6));
  if (Wire.available() == 6) {
    uint16_t raw_temp = (Wire.read() << 8) | (Wire.read() & 0xff);
    uint8_t temp_crc = Wire.read();
    uint16_t raw_humidity = (Wire.read() << 8) | (Wire.read() & 0xff);
    uint8_t humidity_crc = Wire.read();
    return Reading(ConvertTemperature(raw_temp), ConvertHumidity(raw_humidity));
  }
  Serial.println("Failed to read sensor");
  return SHT31D::Reading(0.0f, 0.0f);
}