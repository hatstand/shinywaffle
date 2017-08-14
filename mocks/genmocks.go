//go:generate mockgen -package mocks -destination MockI2CBus.go github.com/kidoman/embd I2CBus
//go:generate mockgen -package mocks -destination MockSPIBus.go github.com/kidoman/embd SPIBus
package mocks
