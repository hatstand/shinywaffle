#include <Arduino.h>
#include <SPI.h>
#include <Wire.h>

#include "LowPower.h"
#include "SHT31D.h"
#include "spi_transaction.h"

#include "cc1101_868_3.h"

class CC1101 {
 private:
  const int kSlaveSelectPin = 10;
  const int kGDO0Pin = 5;
  const int kGDO2Pin = 6;

  const byte kWriteSingleByte = 0x00;
  const byte kWriteBurst = 0x40;
  const byte kReadBurst = 0xc0;
  const byte kReadSingleByte = 0x80;

  const byte kSRES = 0x30;  // Reset
  const byte kSTX = 0x35;   // Transmit mode
  const byte kSFTX = 0x3b;  // Flush TX FIFO buffer
  const byte kSIDLE = 0x36; // Idle mode
  const byte kSPWD = 0x39;  // Sleep mode

  const byte kIOCFG0Config = 0x07;
  const byte kIOCFG1Config = 0x2e;
  const byte kIOCFG2Config = 0x06;

  const byte kIOCFG0 = 0x02;
  const byte kIOCFG1 = 0x01;
  const byte kIOCFG2 = 0x00;

  const byte kTXFIFO = 0x3f;

  const byte kFIFOTHR = 0x03;
  const byte kSYNC1 = 0x04;
  const byte kSYNC0 = 0x05;

  const byte kPKTLEN = 0x06;
  const byte kPKTCTRL1 = 0x07;
  const byte kPKTCTRL0 = 0x08;

  const byte kADDR = 0x09;

  const byte kCHANNR = 0x0a;
  const byte kFSCTRL1 = 0x0b;
  const byte kFSCTRL0 = 0x0c;

  const byte kFREQ2 = 0x0d;
  const byte kFREQ1 = 0x0e;
  const byte kFREQ0 = 0x0f;

  const byte kMDMCFG4 = 0x10;
  const byte kMDMCFG3 = 0x11;
  const byte kMDMCFG2 = 0x12;
  const byte kMDMCFG1 = 0x13;
  const byte kMDMCFG0 = 0x14;

  const byte kDEVIATN = 0x15;

  const byte kMCSM2 = 0x16;
  const byte kMCSM1 = 0x17;
  const byte kMCSM0 = 0x18;

  const byte kFOCCFG = 0x19;
  const byte kBSCFG = 0x1a;

  const byte kAGCCTRL2 = 0x1b;
  const byte kAGCCTRL1 = 0x1c;
  const byte kAGCCTRL0 = 0x1d;

  const byte kFREND1 = 0x21;
  const byte kFREND0 = 0x22;

  const byte kFSCAL3 = 0x23;
  const byte kFSCAL2 = 0x24;
  const byte kFSCAL1 = 0x25;
  const byte kFSCAL0 = 0x26;

  const byte kFSTEST = 0x29;
  const byte kTEST2 = 0x2c;
  const byte kTEST1 = 0x2d;
  const byte kTEST0 = 0x2e;

  byte ReadRegister(byte addr) {
    SPITransaction transaction(kSlaveSelectPin);

    SPI.transfer(addr | kReadSingleByte);
    byte reply = SPI.transfer(0x00);

    return reply;
  }

  void WriteByte(byte address, byte data) {
    SPITransaction transaction(kSlaveSelectPin);
    SPI.transfer(address | kWriteSingleByte);
    SPI.transfer(data);
  }

  void Strobe(byte command) {
    SPITransaction transaction(kSlaveSelectPin);
    SPI.transfer(command);
  }

  void WriteBurst(byte address, byte* data, int length) {
    SPITransaction transaction(kSlaveSelectPin);
    SPI.transfer(address | kWriteBurst);
    for (int i = 0; i < length; ++i) {
      data[i] = SPI.transfer(data[i]);
    }
  }

  void ReadBurst(byte address, byte* buffer, int length) {
    SPITransaction transaction(kSlaveSelectPin);
    SPI.transfer(address | kReadBurst);
    for (int i = 0; i < length; ++i) {
      buffer[i] = SPI.transfer(0x00);
    }
  }

 public:
  void SendPacket(byte* buffer, byte length) {
    Serial.println("Sending packet");
    WriteByte(kTXFIFO, length);
    WriteBurst(kTXFIFO, buffer, length);
    // Send the packet.
    Strobe(kSTX);

    // Wait for sync word to transmit.
    while(!digitalRead(kGDO2Pin)) {}
    // Wait for packet to be sent.
    while(digitalRead(kGDO2Pin)) {}

    // Return to idle state.
    Strobe(kSIDLE);
    // Flush the TXFIFO buffer.
    Strobe(kSFTX);
    Serial.println("Sent packet");
  }

  void Sleep() {
    Strobe(kSPWD);
  }

  void Wake() {
    Strobe(kSIDLE);
    delay(1);  // Should only take 240us.
  }

  void Reset() {
    pinMode(kGDO0Pin, INPUT);
    pinMode(kGDO2Pin, INPUT);
    pinMode(kSlaveSelectPin, OUTPUT);

    Strobe(kSRES);
    delay(100);

    WriteByte(kFSCTRL1, FSCTRL1);
    WriteByte(kFSCTRL0, FSCTRL0);

    WriteByte(kFREQ2, FREQ2);
    WriteByte(kFREQ1, FREQ1);
    WriteByte(kFREQ0, FREQ0);

    WriteByte(kMDMCFG4, MDMCFG4);
    WriteByte(kMDMCFG3, MDMCFG3);
    WriteByte(kMDMCFG2, MDMCFG2);
    WriteByte(kMDMCFG1, MDMCFG1);
    WriteByte(kMDMCFG0, MDMCFG0);

    WriteByte(kCHANNR, CHANNR);
    WriteByte(kDEVIATN, DEVIATN);
    WriteByte(kFREND1, FREND1);
    WriteByte(kFREND0, FREND0);
    WriteByte(kMCSM2, MCSM2);
    WriteByte(kMCSM1, MCSM1);
    WriteByte(kMCSM0, MCSM0);
    WriteByte(kFOCCFG, FOCCFG);
    WriteByte(kBSCFG, BSCFG);

    WriteByte(kAGCCTRL2, AGCCTRL2);
    WriteByte(kAGCCTRL1, AGCCTRL1);
    WriteByte(kAGCCTRL0, AGCCTRL0);

    WriteByte(kFSCAL3, FSCAL3);
    WriteByte(kFSCAL2, FSCAL2);
    WriteByte(kFSCAL1, FSCAL1);
    WriteByte(kFSCAL0, FSCAL0);

    WriteByte(kFSTEST, FSTEST);
    WriteByte(kTEST2, TEST2);
    WriteByte(kTEST1, TEST1);
    WriteByte(kTEST0, TEST0);

    WriteByte(kIOCFG0, kIOCFG0Config);
    WriteByte(kIOCFG1, kIOCFG1Config);
    WriteByte(kIOCFG2, kIOCFG2Config);

    WriteByte(kPKTCTRL1, PKTCTRL1);
    WriteByte(kPKTCTRL0, PKTCTRL0);

    WriteByte(kADDR, ADDR);

    WriteByte(kPKTLEN, PKTLEN);

    WriteByte(kSYNC1, 0x42);
    WriteByte(kSYNC0, 0x42);

    Serial.print(ReadRegister(0xf0));
    Serial.print(ReadRegister(0xf1));
  }
};

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