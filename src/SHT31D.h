#ifndef SHT31D_H
#define SHT31D_H

#include <Arduino.h>

class SHT31D {
 public:
  void Init();

  struct Reading {
    float temperature;
    float humidity;

    Reading(float temperature, float humidity);
  };

  Reading ReadTemperatureAndHumidity();

 private:
  static float ConvertTemperature(uint16_t raw_temp);
  static float ConvertHumidity(uint16_t raw_humidity);
  static bool WriteCommand(uint16_t command);

  static const uint8_t kSensorAddress = 0x44;
  static const uint16_t kSoftReset = 0x30a2;
  static const uint16_t kReadSensor = 0x2400;
};

#endif  // SHT31D_H