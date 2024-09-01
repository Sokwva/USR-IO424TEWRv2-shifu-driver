package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/simonvetter/modbus"
)

func ReadCoils(target string, unitId uint, addr uint16, quantity uint16) ([]bool, error) {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return []bool{}, err
	}
	defer client.Close()
	values, err := client.ReadCoils(addr, quantity)
	if err != nil {
		return []bool{}, err
	}
	return values, nil
}

func ReadHoldRegs(target string, unitId uint, addr uint16, quantity uint16) ([]byte, error) {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return []byte{}, err
	}
	defer client.Close()
	values, err := client.ReadBytes(addr, quantity, modbus.HOLDING_REGISTER)
	if err != nil {
		return []byte{}, err
	}
	return values, nil
}

func ReadInput(target string, unitId uint, addr uint16, quantity uint16) ([]bool, error) {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return []bool{}, err
	}
	defer client.Close()
	values, err := client.ReadDiscreteInputs(addr, quantity)
	if err != nil {
		return []bool{}, err
	}
	return values, nil
}

func ReadOutCoils(target string, unitId uint) ([]bool, error) {
	return ReadCoils(target, unitId, 0, 4)
}

func ReadInDiscrete(target string, unitId uint) ([]bool, error) {
	return ReadInput(target, unitId, 32, 4)
}

func ReadOutHoldRegs(target string, unitId uint) ([]byte, error) {
	return ReadHoldRegs(target, unitId, 0, 4)
}

func ReadInHoldRegs(target string, unitId uint) ([]byte, error) {
	return ReadHoldRegs(target, unitId, 32, 4)
}

func WriteOutCoil(target string, unitId uint, coilNum uint16, state bool) error {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.WriteCoil(coilNum, state)
	if err != nil {
		return err
	}
	return nil
}

func WriteOutCoils(target string, unitId uint, coilNum []uint16, state []bool) []error {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return []error{
			err,
		}
	}
	defer client.Close()
	errs := []error{}
	for i, v := range coilNum {
		err = client.WriteCoil(uint16(v), state[i])
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func ReadPT100(target string, unitId uint, regType modbus.RegType) (float32, error) {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	values, err := client.ReadBytes(uint16(80), 2, regType)
	if err != nil {
		return 0, err
	}
	slices.Reverse(values)
	num, err := strconv.ParseUint(hex.EncodeToString(values), 16, 16)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(fmt.Sprintf("%.2f", (float32(num)-1e4)/1e2), 32)
	if err != nil {
		return 0, err
	}
	return float32(value), nil
}

func ReadAnalogIn(target string, unitId uint, dataType string, regType modbus.RegType) (uint64, error) {
	client, err := commonClient(target, unitId, 10)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	var values []byte
	if dataType == "mV" {
		values, err = client.ReadBytes(uint16(88), 4, regType)
	}
	if dataType == "uA" {
		values, err = client.ReadBytes(uint16(60), 4, regType)
	}
	if err != nil {
		return 0, err
	}
	slices.Reverse(values)
	if dataType == "mA" && len(values) == 0 {
		return 0, errors.New("no analog input")
	}
	num, err := strconv.ParseUint(hex.EncodeToString(values), 16, 16)
	if err != nil {
		return 0, err
	}
	return num, nil
}
