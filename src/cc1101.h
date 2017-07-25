#ifndef CC1101_H
#define CC1101_H

#include <Arduino.h>

class CC1101 {
 public:
  void SendPacket(byte* buffer, byte length);
  void Sleep();
  void Wake();
  void Reset();

 private:
  byte ReadRegister(byte addr);
  void WriteByte(byte address, byte data);
  void Strobe(byte command);
  void WriteBurst(byte address, byte* data, int length);
  void ReadBurst(byte address, byte* buffer, int length);

  static const int kSlaveSelectPin = 10;
  static const int kGDO0Pin = 5;
  static const int kGDO2Pin = 6;

  static const byte kWriteSingleByte = 0x00;
  static const byte kWriteBurst = 0x40;
  static const byte kReadBurst = 0xc0;
  static const byte kReadSingleByte = 0x80;

  static const byte kSRES = 0x30;  // Reset
  static const byte kSTX = 0x35;   // Transmit mode
  static const byte kSFTX = 0x3b;  // Flush TX FIFO buffer
  static const byte kSIDLE = 0x36; // Idle mode
  static const byte kSPWD = 0x39;  // Sleep mode

  static const byte kIOCFG0Config = 0x07;
  static const byte kIOCFG1Config = 0x2e;
  static const byte kIOCFG2Config = 0x06;

  static const byte kIOCFG0 = 0x02;
  static const byte kIOCFG1 = 0x01;
  static const byte kIOCFG2 = 0x00;

  static const byte kTXFIFO = 0x3f;

  static const byte kFIFOTHR = 0x03;
  static const byte kSYNC1 = 0x04;
  static const byte kSYNC0 = 0x05;

  static const byte kPKTLEN = 0x06;
  static const byte kPKTCTRL1 = 0x07;
  static const byte kPKTCTRL0 = 0x08;

  static const byte kADDR = 0x09;

  static const byte kCHANNR = 0x0a;
  static const byte kFSCTRL1 = 0x0b;
  static const byte kFSCTRL0 = 0x0c;

  static const byte kFREQ2 = 0x0d;
  static const byte kFREQ1 = 0x0e;
  static const byte kFREQ0 = 0x0f;

  static const byte kMDMCFG4 = 0x10;
  static const byte kMDMCFG3 = 0x11;
  static const byte kMDMCFG2 = 0x12;
  static const byte kMDMCFG1 = 0x13;
  static const byte kMDMCFG0 = 0x14;

  static const byte kDEVIATN = 0x15;

  static const byte kMCSM2 = 0x16;
  static const byte kMCSM1 = 0x17;
  static const byte kMCSM0 = 0x18;

  static const byte kFOCCFG = 0x19;
  static const byte kBSCFG = 0x1a;

  static const byte kAGCCTRL2 = 0x1b;
  static const byte kAGCCTRL1 = 0x1c;
  static const byte kAGCCTRL0 = 0x1d;

  static const byte kFREND1 = 0x21;
  static const byte kFREND0 = 0x22;

  static const byte kFSCAL3 = 0x23;
  static const byte kFSCAL2 = 0x24;
  static const byte kFSCAL1 = 0x25;
  static const byte kFSCAL0 = 0x26;

  static const byte kFSTEST = 0x29;
  static const byte kTEST2 = 0x2c;
  static const byte kTEST1 = 0x2d;
  static const byte kTEST0 = 0x2e;
};

#endif  // CC1101_H