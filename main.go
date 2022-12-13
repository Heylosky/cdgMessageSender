package main

import (
	"encoding/json"
	"github.com/cdgProcessor/messageSender/logger"
	"github.com/cdgProcessor/messageSender/messageQ"
	"github.com/cdgProcessor/messageSender/models"
	messagebird "github.com/messagebird/go-rest-api/v9"
	"github.com/messagebird/go-rest-api/v9/sms"
	"go.uber.org/zap"
	"os"
)

type smsToMBConf struct {
	MBKey string `json:"mb_key"`
}

func main() {
	logger.InitLogger("./msgSender.log")
	zap.L().Info("Processor sms to MB starting...")

	MBConfig := readFile("./MBconfig.json")

	in2mbChan := make(chan []byte)

	go messageQ.MQRead(in2mbChan, "inboundSMS", "smsToMB", "consumeToMB")

	MBSender(in2mbChan, MBConfig.MBKey)
}

func readFile(filePath string) *smsToMBConf {
	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Panic("No file found.")
	}
	defer file.Close()
	var conf smsToMBConf
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		zap.L().Panic("Read config failed.")
	}
	return &conf
}

func MBSender(c1 <-chan []byte, accessKey string) {
	client := messagebird.New(accessKey)

	params := &sms.Params{Reference: "MyReference"}

	var msg models.SMS

	for ctx := range c1 {
		json.Unmarshal(ctx, &msg)
		message, err := sms.Create(
			client,
			msg.Originator,
			[]string{msg.Recipients},
			msg.Payload,
			params,
		)
		if err != nil {
			//retry
			message, err = sms.Create(
				client,
				msg.Originator,
				[]string{msg.Recipients},
				msg.Payload,
				params,
			)
			if err != nil {
				logger.SMSLogOnError(msg.Originator, msg.Payload, msg.Recipients)
			} else {
				logger.SMSLogger(message)
			}
		} else {
			logger.SMSLogger(message)
		}
	}
}
