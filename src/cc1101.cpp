#include "cc1101.h"

#include <Arduino.h>
#include <SPI.h>

#include "cc1101_868_3.h"
#include "spi_transaction.h"

byte CC1101::ReadRegister(byte addr) {
  SPITransaction transaction(kSlaveSelectPin);

  SPI.transfer(addr | kReadSingleByte);
  byte reply = SPI.transfer(0x00);

  return reply;
}

void CC1101::WriteByte(byte address, byte data) {
  SPITransaction transaction(kSlaveSelectPin);
  SPI.transfer(address | kWriteSingleByte);
  SPI.transfer(data);
}

void CC1101::Strobe(byte command) {
  SPITransaction transaction(kSlaveSelectPin);
  SPI.transfer(command);
}

void CC1101::WriteBurst(byte address, byte* data, int length) {
  SPITransaction transaction(kSlaveSelectPin);
  SPI.transfer(address | kWriteBurst);
  for (int i = 0; i < length; ++i) {
    data[i] = SPI.transfer(data[i]);
  }
}

void CC1101::ReadBurst(byte address, byte* buffer, int length) {
  SPITransaction transaction(kSlaveSelectPin);
  SPI.transfer(address | kReadBurst);
  for (int i = 0; i < length; ++i) {
    buffer[i] = SPI.transfer(0x00);
  }
}

void CC1101::SendPacket(byte* buffer, byte length) {
  Serial.println("Sending packet");
  WakeLock lock(this);
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

void CC1101::Sleep() {
  Strobe(kSPWD);
}

void CC1101::Wake() {
  Strobe(kSIDLE);
  delay(1);  // Should only take 240us.
}

CC1101::WakeLock::WakeLock(const CC1101* cc1101)
  : cc1101_(cc1101) {
  cc1101_->Wake();
}

CC1101::WakeLock::~WakeLock() {
  cc1101_->Sleep();
}

void CC1101::Reset() {
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

  // Always Be Sleeping.
  Sleep();
}