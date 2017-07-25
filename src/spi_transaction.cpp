#include "spi_transaction.h"

#include <Arduino.h>
#include <SPI.h>

SPITransaction::SPITransaction(int slave_select_pin)
    : slave_select_pin_(slave_select_pin) {
  digitalWrite(slave_select_pin_, LOW);
  SPI.beginTransaction(SPISettings(50000, MSBFIRST, SPI_MODE0));
}

SPITransaction::~SPITransaction() {
  digitalWrite(slave_select_pin_, HIGH);
  SPI.endTransaction();
}