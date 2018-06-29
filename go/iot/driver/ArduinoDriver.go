package driver

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/spi"
	"time"
)

// ArduinoDriver is a driver for the communication between the Raspberry Pi and Arduino.
type ArduinoDriver struct {
	name       string
	connector  spi.Connector
	connection spi.Connection
	spi.Config
	gobot.Commander
}

// NewArduinoDriver creates a new Gobot Driver for ArduinoDriver SPI communication
//
// Params:
//      a *Adaptor - the Adaptor to use with this Driver
//
// Optional params:
//      spi.WithBus(int):    	bus to use with this driver
//     	spi.WithChip(int):    	chip to use with this driver
//      spi.WithMode(int):    	mode to use with this driver
//      spi.WithBits(int):    	number of bits to use with this driver
//      spi.WithSpeed(int64):   speed in Hz to use with this driver
//
func NewArduinoDriver(a spi.Connector, options ...func(spi.Config)) *ArduinoDriver {
	d := &ArduinoDriver{
		name:      gobot.DefaultName("ArduinoDriver"),
		connector: a,
		Config:    spi.NewConfig(),
	}
	for _, option := range options {
		option(d)
	}
	return d
}

// Name returns the name of the device.
func (d *ArduinoDriver) Name() string { return d.name }

// SetName sets the name of the device.
func (d *ArduinoDriver) SetName(n string) { d.name = n }

// Connection returns the Connection of the device.
func (d *ArduinoDriver) Connection() gobot.Connection { return d.connection.(gobot.Connection) }

// Start initializes the driver.
func (d *ArduinoDriver) Start() (err error) {
	bus := d.GetBusOrDefault(d.connector.GetSpiDefaultBus())
	chip := d.GetChipOrDefault(d.connector.GetSpiDefaultChip())
	mode := d.GetModeOrDefault(d.connector.GetSpiDefaultMode())
	bits := d.GetBitsOrDefault(d.connector.GetSpiDefaultBits())
	maxSpeed := d.GetSpeedOrDefault(d.connector.GetSpiDefaultMaxSpeed())

	d.connection, err = d.connector.GetSpiConnection(bus, chip, mode, bits, maxSpeed)
	if err != nil {
		return err
	}
	return nil
}

// Halt stops the driver.
func (d *ArduinoDriver) Halt() (err error) {
	d.connection.Close()
	return
}

// Read reads the current analog data for the desired channel.
func (d *ArduinoDriver) Read(count int) (result []byte, err error) {

	rx := make([]byte, count)
	tx := make([]byte, count)

	err = d.connection.Tx(tx, rx)
	return rx, err
}

// Read reads the current analog data for the desired channel.
func (d *ArduinoDriver) Write(tx []byte) (err error) {

	rx := make([]byte, len(tx))
	err = d.connection.Tx(tx, rx)
	fmt.Println(rx, err)
	return err
}

// AnalogRead returns value from analog reading of specified pin
func (d *ArduinoDriver) AnalogRead(pin string) (value int, err error) {
	var tmp int = len(pin)
	result, err := d.Read(tmp)
	return len(result), err
}

func (d *ArduinoDriver) TransferAndWait(what byte) (result byte, err error) {
	rx := make([]byte, 1)
	tx := make([]byte, 1)
	tx[0] = what
	err = d.connection.Tx(tx, rx)
	time.Sleep(20 * time.Microsecond)
	return rx[0], err
}
