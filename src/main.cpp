#include "Arduino.h"

#include <SPI.h>
#include <Wire.h>

using namespace std;

const int kSlaveSelectPin = 10;

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

class SHT31D {
  static const uint8_t kSensorAddress = 0x44;
  static const uint16_t kSoftReset = 0x30a2;
  static const uint16_t kReadSensor = 0x2400;

 public:
  void Init() {
    WriteCommand(kSoftReset);
    delay(10);
  }

  struct Reading {
    float temperature;
    float humidity;

    Reading(float temperature, float humidity)
        : temperature(temperature),
          humidity(humidity) {}
  };

 private:
  static float ConvertTemperature(uint16_t raw_temp) {
    return -45 + 175 * (float(raw_temp) / 0xffff);
  }

  static float ConvertHumidity(uint16_t raw_humidity) {
    return 100 * (float(raw_humidity) / 0xffff);
  }

  static bool WriteCommand(uint16_t command) {
    Wire.beginTransmission(kSensorAddress);
    Wire.write(command >> 8);
    Wire.write(command & 0xff);
    return Wire.endTransmission() == 0;
  }

 public:
  Reading ReadTemperatureAndHumidity() {
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
    return Reading(0.0f, 0.0f);
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

static SHT31D sensor;

void setup() {
  // initialize LED digital pin as an output.
  pinMode(LED_BUILTIN, OUTPUT);

  // Initialise the Slave Select pin.
  pinMode(kSlaveSelectPin, OUTPUT);

  Serial.begin(9600);

  SPI.begin();
  Wire.begin();

  sensor.Init();

  delay(100);
}

void loop() {
  // turn the LED on (HIGH is the voltage level)
  digitalWrite(LED_BUILTIN, HIGH);

  // wait for a second
  delay(1000);

  // turn the LED off by making the voltage LOW
  digitalWrite(LED_BUILTIN, LOW);

  Serial.write("Hello world!\n");

   // wait for a second
  delay(1000);

  SHT31D::Reading reading = sensor.ReadTemperatureAndHumidity();
  Serial.print(reading.temperature);
  Serial.println();
  Serial.print(reading.humidity);
  Serial.println();
}