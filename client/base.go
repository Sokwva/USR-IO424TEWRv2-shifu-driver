package client

import (
	"time"

	"github.com/simonvetter/modbus"
)

func commonClient(clientAddr string, unitId uint, timeoutSec uint) (*modbus.ModbusClient, error) {
	var client *modbus.ModbusClient
	var err error

	client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:     clientAddr,
		Timeout: time.Duration(timeoutSec) * time.Second,
	})
	client.SetEncoding(modbus.LITTLE_ENDIAN, modbus.LOW_WORD_FIRST)

	client.SetUnitId(uint8(unitId))

	if err != nil {
		return nil, err
	}

	err = client.Open()
	if err != nil {
		return nil, err
	}

	return client, nil
}
