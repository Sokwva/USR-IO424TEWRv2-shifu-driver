package mqtt

import (
	"encoding/json"
	"log"
	clientBussi "sokwva/shaft/usr-io424tewrv2-shifu-driver/client"
	"sokwva/shaft/usr-io424tewrv2-shifu-driver/utils"
	"strconv"
	"strings"
	"time"

	mqttDrv "github.com/eclipse/paho.mqtt.golang"
	"github.com/simonvetter/modbus"
)

var (
	// 设备 tcp://192.168.1.20:502
	target string = ""
	// MQTT 中间件
	addr        string = ""
	user        string = ""
	password    string = ""
	topicPath   string = ""
	name        string = ""
	client      mqttDrv.Client
	healthCheck bool
	enviroment  string

	modbusUnit uint = 1

	checkChan chan bool = make(chan bool)
)

func SetUp(targetExt string, healthChkExt bool, envExt string, addrExt string, userExt string, passExt string, pathExt string, nameExt string, unitId uint) {
	target = targetExt
	addr = addrExt
	user = userExt
	password = passExt
	topicPath = pathExt
	name = nameExt
	healthCheck = healthChkExt
	modbusUnit = unitId
}

func Serve() {
	log.Println("start mqtt driver")
	optios := mqttDrv.NewClientOptions()
	optios.AddBroker("tcp://" + addr)
	optios.SetClientID(target)
	optios.SetUsername(user)
	optios.SetPassword(password)

	client = mqttDrv.NewClient(optios)
	defer client.Disconnect(3)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	Pub("[deviceLive] device " + name + " is ready")
	Sub()

	if healthCheck {
		go HealthCheck()
	}
	go LoopToReportStatus()

	ret := <-checkChan
	if !ret {
		log.Fatal("device is not healthy")
		return
	}
}

func Pub(payload string) bool {
	token := client.Publish(topicPath+name, 0, false, payload)
	return token.Wait()
}

func Sub() bool {
	topicAll := topicPath + name + "/cmd"
	token := client.Subscribe(topicAll, 0, MqttMsgEventHandler)
	return token.Wait()
}

func MqttMsgEventHandler(ctx mqttDrv.Client, msg mqttDrv.Message) {
	if len(msg.Payload()) == 0 {
		return
	}
	//command params
	cmdList := strings.Split(string(msg.Payload()), " ")

	switch cmdList[0] {
	case "get":
		ReportStates()
	case "close":
		if len(cmdList) == 1 {
			errs := clientBussi.WriteOutCoils(target, modbusUnit, []uint16{0, 1, 2, 3}, []bool{false, false, false, false})
			for i, err := range errs {
				if err != nil {
					log.Printf("MqttMsgEventHandler > closeAll: close %d faild.", i)
				}
			}
		}
		if len(cmdList) == 2 {
			btnNum, err := strconv.Atoi(cmdList[1])
			if err != nil {
				log.Printf("MqttMsgEventHandler > close: %s is not a valid btnNum.", cmdList[1])
				return
			}
			clientBussi.WriteOutCoil(target, modbusUnit, uint16(btnNum), false)
		}
		ReportStates()
	case "open":
		if len(cmdList) == 1 {
			errs := clientBussi.WriteOutCoils(target, modbusUnit, []uint16{0, 1, 2, 3}, []bool{true, true, true, true})
			for i, err := range errs {
				if err != nil {
					log.Printf("MqttMsgEventHandler > openAll: close %d faild.", i)
				}
			}
		}
		if len(cmdList) == 2 {
			btnNum, err := strconv.Atoi(cmdList[1])
			if err != nil {
				log.Printf("MqttMsgEventHandler > open: %s is not a valid btnNum.", cmdList[1])
				return
			}
			clientBussi.WriteOutCoil(target, modbusUnit, uint16(btnNum), true)
		}
		ReportStates()
	}
}

type LoopReport struct {
	GetInputSuccess  bool
	GetOutputSuccess bool
	GetPT100Success  bool

	GetVotageSuccess bool

	InputState  []bool
	OutputState []bool
	PT100       float32
	Analog      struct {
		Millivolt  uint64
		Millampere uint64
	}
}

func ReportStates() {
	result := LoopReport{}

	inputState, err := clientBussi.ReadInDiscrete(target, modbusUnit)
	if err != nil {
		log.Printf("ReportState > InState: %s\n", err.Error())
		result.GetInputSuccess = false
		inputState = []bool{false, false, false, false}
	} else {
		result.GetInputSuccess = true
	}
	result.InputState = inputState

	outputState, err := clientBussi.ReadOutCoils(target, modbusUnit)
	if err != nil {
		log.Printf("ReportState > OutState: %s\n", err.Error())
		result.GetOutputSuccess = false
		outputState = []bool{false, false, false, false}
	} else {
		result.GetOutputSuccess = true
	}
	result.OutputState = outputState

	PT100Value, err := clientBussi.ReadPT100(target, modbusUnit, modbus.INPUT_REGISTER)
	if err != nil {
		log.Printf("ReportState > PT100Value: %s\n", err.Error())
		result.GetPT100Success = false
		PT100Value = 0
	} else {
		result.GetPT100Success = true
	}
	result.PT100 = PT100Value

	AnalogmV, err := clientBussi.ReadAnalogIn(target, modbusUnit, "mV", modbus.INPUT_REGISTER)
	if err != nil {
		log.Printf("ReportState > AnalogmV: %s\n", err.Error())
		result.GetVotageSuccess = false
		AnalogmV = 0
	} else {
		result.GetVotageSuccess = true
	}
	result.Analog.Millivolt = AnalogmV

	resultStr, err := json.Marshal(result)
	if err != nil {
		log.Println("ReportStatus: Marshal json result faild.")
		return
	}
	Pub(string(resultStr))
}

func LoopToReportStatus() {
	reporterTimer := time.NewTicker(time.Second * 30)
	for {
		log.Print("Reporting states...")
		<-reporterTimer.C
		ReportStates()
	}
}

func HealthCheck() {
	timer := time.NewTicker(time.Second * 10)
	for {
		log.Print("Checking healthy...")
		<-timer.C
		if utils.ProbeTCP(strings.Replace(strings.Replace(target, "tcp://", "", -1), "udp://", "", -1)) {
			continue
		} else {
			if enviroment == "host" {
				log.Print("warn: device is not healthy.")
				Pub("[healthCheck] device " + target + " is not healthy")
			}
			if enviroment == "container" {
				checkChan <- false
			}
		}
	}
}
