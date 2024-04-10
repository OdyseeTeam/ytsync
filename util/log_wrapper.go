package util

import (
	"fmt"
	"time"

	"github.com/lbryio/ytsync/v5/configs"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/util"

	log "github.com/sirupsen/logrus"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// SendErrorToSlack Sends an error message to the default channel and to the process log.
func SendErrorToSlack(format string, a ...interface{}) {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	err := sendToGraylog(message)
	if err != nil {
		log.Errorln(err)
	}
	log.Errorln(message)
	log.SetLevel(log.InfoLevel) //I don't want to change the underlying lib so this will do...
	err = util.SendToSlack(":sos: ```" + message + "```")
	log.SetLevel(log.DebugLevel)
	if err != nil {
		log.Errorln(err)
	}
}

// SendInfoToSlack Sends an info message to the default channel and to the process log.
func SendInfoToSlack(format string, a ...interface{}) {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	err := sendToGraylog(message)
	if err != nil {
		log.Errorln(err)
	}
	log.Infoln(message)
	log.SetLevel(log.InfoLevel) //I don't want to change the underlying lib so this will do...
	err = util.SendToSlack(":information_source: " + message)
	log.SetLevel(log.DebugLevel)
	if err != nil {
		log.Errorln(err)
	}
}

var gelfClient *gelf.TCPWriter

func sendToGraylog(message string) error {
	if gelfClient == nil && configs.Configuration.LoggingEndpoint != "" {
		var err error
		gelfClient, err = gelf.NewTCPWriter(configs.Configuration.LoggingEndpoint)
		if err != nil {
			return errors.Err(err)
		}
	}
	if gelfClient != nil {
		err := gelfClient.WriteMessage(&gelf.Message{
			Host:     configs.Configuration.GetHostname(),
			Short:    message,
			TimeUnix: float64(time.Now().Unix()),
		})
		if err != nil {
			return errors.Err(err)
		}
	}
	return nil
}
