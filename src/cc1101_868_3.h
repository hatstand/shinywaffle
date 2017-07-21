#ifndef CC1101_CONFIG_H
#define CC1101_CONFIG_H

#include <Arduino.h>

// Rf settings for CC1101
const byte     IOCFG2 = 0x06;        // GDO2 Output Pin Configuration
const byte     IOCFG1 = 0x2e;        // GDO1 Output Pin Configuration
const byte     IOCFG0 = 0x07;        // GDO0 Output Pin Configuration
const byte     FIFOTHR = 0x47;       // RX FIFO and TX FIFO Thresholds
const byte     SYNC1 = 0xd3;         // Sync Word, High Byte
const byte     SYNC0 = 0x91;         // Sync Word, Low Byte
const byte     PKTLEN = 0x3d;        // Packet Length
const byte     PKTCTRL1 = 0x04;      // Packet Automation Control
const byte     PKTCTRL0 = 0x05;      // Packet Automation Control
const byte     ADDR = 0x00;          // Device Address (Broadcast)
const byte     CHANNR = 0x00;        // Channel Number
const byte     FSCTRL1 = 0x06;       // Frequency Synthesizer Control
const byte     FSCTRL0 = 0x00;       // Frequency Synthesizer Control
const byte     FREQ2 = 0x21;         // Frequency Control Word, High Byte
const byte     FREQ1 = 0x65;         // Frequency Control Word, Middle Byte
const byte     FREQ0 = 0x44;         // Frequency Control Word, Low Byte
const byte     MDMCFG4 = 0xf5;       // Modem Configuration
const byte     MDMCFG3 = 0x83;       // Modem Configuration
const byte     MDMCFG2 = 0x03;       // Modem Configuration
const byte     MDMCFG1 = 0x22;       // Modem Configuration
const byte     MDMCFG0 = 0xf8;       // Modem Configuration
const byte     DEVIATN = 0x34;       // Modem Deviation Setting
const byte     MCSM2 = 0x07;         // Main Radio Control State Machine Configuration
const byte     MCSM1 = 0x30;         // Main Radio Control State Machine Configuration
const byte     MCSM0 = 0x18;         // Main Radio Control State Machine Configuration
const byte     FOCCFG = 0x16;        // Frequency Offset Compensation Configuration
const byte     BSCFG = 0x6c;         // Bit Synchronization Configuration
const byte     AGCCTRL2 = 0x03;      // AGC Control
const byte     AGCCTRL1 = 0x40;      // AGC Control
const byte     AGCCTRL0 = 0x91;      // AGC Control
const byte     WOREVT1 = 0x87;       // High Byte Event0 Timeout
const byte     WOREVT0 = 0x6b;       // Low Byte Event0 Timeout
const byte     WORCTRL = 0xfb;       // Wake On Radio Control
const byte     FREND1 = 0x56;        // Front End RX Configuration
const byte     FREND0 = 0x10;        // Front End TX Configuration
const byte     FSCAL3 = 0xe9;        // Frequency Synthesizer Calibration
const byte     FSCAL2 = 0x2a;        // Frequency Synthesizer Calibration
const byte     FSCAL1 = 0x00;        // Frequency Synthesizer Calibration
const byte     FSCAL0 = 0x1f;        // Frequency Synthesizer Calibration
const byte     RCCTRL1 = 0x41;       // RC Oscillator Configuration
const byte     RCCTRL0 = 0x00;       // RC Oscillator Configuration
const byte     FSTEST = 0x59;        // Frequency Synthesizer Calibration Control
const byte     PTEST = 0x7f;         // Production Test
const byte     AGCTEST = 0x3f;       // AGC Test
const byte     TEST2 = 0x81;         // Various Test Settings
const byte     TEST1 = 0x35;         // Various Test Settings
const byte     TEST0 = 0x09;         // Various Test Settings
const byte     PARTNUM = 0x00;       // Chip ID
const byte     VERSION = 0x04;       // Chip ID
const byte     FREQEST = 0x00;       // Frequency Offset Estimate from Demodulator
const byte     LQI = 0x00;           // Demodulator Estimate for Link Quality
const byte     RSSI = 0x00;          // Received;Signal Strength Indication
const byte     MARCSTATE = 0x00;     // Main Radio Control State Machine State
const byte     WORTIME1 = 0x00;      // High Byte of WOR Time
const byte     WORTIME0 = 0x00;      // Low Byte of WOR Time
const byte     PKTSTATUS = 0x00;     // Current GDOx Status and Packet Status
const byte     VCO_VC_DAC = 0x00;    // Current Setting from PLL Calibration Module
const byte     TXBYTES = 0x00;       // Underflow and Number of Bytes
const byte     RXBYTES = 0x00;       // Overflow and Number of Bytes
const byte     RCCTRL1_STATUS = 0x00;// Last RC Oscillator Calibration Result
const byte     RCCTRL0_STATUS = 0x00;// Last RC Oscillator Calibration Result

#endif  // CC1101_CONFIG_H
