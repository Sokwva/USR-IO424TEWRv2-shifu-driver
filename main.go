package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sokwva/shaft/usr-io424tewrv2-shifu-driver/client"
	"sokwva/shaft/usr-io424tewrv2-shifu-driver/server/mqtt"
	"strconv"

	"github.com/simonvetter/modbus"
	"github.com/urfave/cli/v2"
)

var (
	target      string = ""
	healthCheck bool   = false
	enviroment  string = "container"
	serverType  string = "http"

	mqttAddr            string = ""
	mqttUser            string = ""
	mqttPassword        string = ""
	mqttTitleName       string = ""
	mqttParentTopicPath string = ""

	modbusUnit uint = 1
)

func main() {
	cliApp := &cli.App{
		Name:  "USR-IO424TEWRv2-shifu-driver",
		Usage: "USR-IO424TEWRv2-shifu-driver [options]",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "start server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "target",
						Value:       "tcp://192.168.1.20:502",
						Usage:       "ip address and port of target device",
						Destination: &target,
					},
					&cli.UintFlag{
						Name:        "unit",
						Value:       1,
						Usage:       "remote device modbus address",
						Destination: &modbusUnit,
					},
					&cli.BoolFlag{
						Name:        "check",
						Value:       true,
						Usage:       "loop to check device health",
						Destination: &healthCheck,
					},
					&cli.StringFlag{
						Name:        "env",
						Value:       "container",
						Usage:       "unhealthy reaction",
						Destination: &enviroment,
					},
					&cli.StringFlag{
						Name:        "svr",
						Value:       "http",
						Usage:       "application layer server type: http or mqtt",
						Destination: &serverType,
					},
					&cli.StringFlag{
						Name:        "mqttAddr",
						Value:       "",
						Usage:       "mqtt broker ip address and port",
						Destination: &mqttAddr,
					},
					&cli.StringFlag{
						Name:        "mqttUser",
						Value:       "",
						Usage:       "mqtt broker user name",
						Destination: &mqttUser,
					},
					&cli.StringFlag{
						Name:        "mqttPassword",
						Value:       "",
						Usage:       "mqtt broker user password",
						Destination: &mqttPassword,
					},
					&cli.StringFlag{
						Name:        "mqttName",
						Value:       "unkowDevice",
						Usage:       "mqtt topic name without parent topic path",
						Destination: &mqttTitleName,
					},
					&cli.StringFlag{
						Name:        "mqttPath",
						Value:       "",
						Usage:       "mqtt topic parent topic path",
						Destination: &mqttParentTopicPath,
					},
				},
				Action: func(ctx *cli.Context) error {
					if serverType == "http" {
						fmt.Println("imple me")
						return nil
					}
					if serverType == "mqtt" {
						if mqttAddr == "" {
							log.Fatal("missing param: mqttAddr")
							return errors.New("missing param: --mqttAddr")
						}
						mqtt.SetUp(target, healthCheck, enviroment, mqttAddr, mqttUser, mqttPassword, mqttParentTopicPath, mqttTitleName, modbusUnit)
						mqtt.Serve()
						return nil
					}
					return nil
				},
			},
			{
				Name:  "test",
				Usage: "test device",
				Subcommands: []*cli.Command{
					{
						Name:  "get",
						Usage: "get device Output coild's state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadOutCoils(target, modbusUnit)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getIn",
						Usage: "get device Input coild's state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadInDiscrete(target, modbusUnit)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getHold",
						Usage: "get device Output coild's hold register state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadOutHoldRegs(target, modbusUnit)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getInHold",
						Usage: "get device Input coild's hold register state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadInHoldRegs(target, modbusUnit)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getPT100",
						Usage: "get device pt100's state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadPT100(target, modbusUnit, modbus.INPUT_REGISTER)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getPT100Hold",
						Usage: "get device pt100's hold register state",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadPT100(target, modbusUnit, modbus.HOLDING_REGISTER)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getmV",
						Usage: "get device Analog of mV",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadAnalogIn(target, modbusUnit, "mV", modbus.INPUT_REGISTER)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "getmA",
						Usage: "get device Analog of mA",
						Action: func(ctx *cli.Context) error {
							values, err := client.ReadAnalogIn(target, modbusUnit, "mA", modbus.INPUT_REGISTER)
							if err != nil {
								return err
							}
							fmt.Println(values)
							return nil
						},
					},
					{
						Name:  "close",
						Usage: "close device coild's",
						Action: func(ctx *cli.Context) error {
							if ctx.Args().Len() != 1 {
								return errors.New("invalid param quantity")
							}
							btnId, err := strconv.Atoi(ctx.Args().Get(0))
							if err != nil {
								return err
							}
							if btnId < 1 || btnId > 4 {
								return errors.New("invalid button id range")
							}
							err = client.WriteOutCoil(target, modbusUnit, uint16(btnId)-1, false)
							return err
						},
					},
					{
						Name:  "closeAll",
						Usage: "close device all coild's",
						Action: func(ctx *cli.Context) error {
							errs := []error{}
							for i := range 4 {
								err := client.WriteOutCoil(target, modbusUnit, uint16(i), false)
								if err != nil {
									fmt.Println("err occured while process coil: "+strconv.Itoa(i), err.Error())
									errs = append(errs, err)
								}
							}
							if len(errs) != 0 {
								return errs[0]
							}
							return nil
						},
					},
					{
						Name:  "open",
						Usage: "open device coild's",
						Action: func(ctx *cli.Context) error {
							if ctx.Args().Len() != 1 {
								return errors.New("invalid param quantity")
							}
							btnId, err := strconv.Atoi(ctx.Args().Get(0))
							if err != nil {
								return err
							}
							if btnId < 1 || btnId > 4 {
								return errors.New("invalid button id range")
							}
							err = client.WriteOutCoil(target, modbusUnit, uint16(btnId)-1, true)
							return err
						},
					},
					{
						Name:  "openAll",
						Usage: "open device all coild's",
						Action: func(ctx *cli.Context) error {
							errs := []error{}
							for i := range 4 {
								err := client.WriteOutCoil(target, modbusUnit, uint16(i), true)
								if err != nil {
									fmt.Println("err occured while process coil: "+strconv.Itoa(i), err.Error())
									errs = append(errs, err)
								}
							}
							if len(errs) != 0 {
								return errs[0]
							}
							return nil
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target",
				Value:       "tcp://192.168.1.20:502",
				Usage:       "ip address and port of target device",
				Destination: &target,
			},
			&cli.UintFlag{
				Name:        "unit",
				Value:       1,
				Usage:       "remote device modbus address",
				Destination: &modbusUnit,
			},
		},
		Action: func(ctx *cli.Context) error {
			if serverType == "test" {
				fmt.Println("here")
			}
			return nil
		},
	}
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}
