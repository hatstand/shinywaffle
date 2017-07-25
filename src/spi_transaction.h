#ifndef SPI_TRANSACTION_H
#define SPI_TRANSACTION_H

class SPITransaction {
 public:
  SPITransaction(int slave_select_pin);
  ~SPITransaction();

 private:
  const int slave_select_pin_;
};

#endif  // SPI_TRANSACTION_H