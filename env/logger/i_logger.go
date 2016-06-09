package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"regexp"

	"github.com/ottemo/foundation/env"
)

// REGEXP_CREDIT_CARD_CVC Regexp to find cvc group
const REGEXP_CREDIT_CARD_CVC = regexp.MustCompile(`"cvc":"\d{3}"`)

// REGEXP_CREDIT_CARD_NUMBER Regexp to find cc number group
const REGEXP_CREDIT_CARD_NUMBER = regexp.MustCompile(`("cc":.+"number":")(\d+)(\d{4})(".+})`)

// Log is a general case logging function
func (it *DefaultLogger) Log(storage string, prefix string, msg string) {
	message := time.Now().Format(time.RFC3339) + " [" + prefix + "]: " + msg + "\n"

	logFile, err := os.OpenFile(baseDirectory+storage, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(message)
		return
	}
	defer logFile.Close()

	logFile.Write([]byte(cleanCreditCardNumber(message)))
}

// LogError makes error log
func (it *DefaultLogger) LogError(err error) {
	if err != nil {
		if ottemoErr, ok := err.(env.InterfaceOttemoError); ok {
			if ottemoErr.ErrorLevel() <= errorLogLevel && !ottemoErr.IsLogged() {
				it.Log(defaultErrorsFile, env.ConstLogPrefixError, ottemoErr.ErrorFull())
				ottemoErr.MarkLogged()
			}
		} else {
			it.Log(defaultErrorsFile, env.ConstLogPrefixError, err.Error())
		}
	}
}

// LogEvent Saves log details out to a file for logstash consumption
func (it *DefaultLogger) LogEvent(fields env.LogFields, eventName string) {
	f, err := os.OpenFile(baseDirectory+"events.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	msg, err := logstashFormatter(fields, eventName)
	if err != nil {
		fmt.Println(err)
	}
	f.Write([]byte(cleanCreditCardNumber(convertByteArrToString(msg))))
}


// return cleaned from credit card number given log msg, so we don't log customer's credit cards
func cleanCreditCardNumber(msg string) string {
	res := REGEXP_CREDIT_CARD_CVC.ReplaceAllString(msg, `"cvc":"123"`)

	return REGEXP_CREDIT_CARD_NUMBER.ReplaceAllString(res, `$1$3$4`)
}

func convertByteArrToString(byteArr []byte) string {
	return string(byteArr[:])
}

func logstashFormatter(fields env.LogFields, eventName string) ([]byte, error) {
	// Attach the message
	fields["message"] = eventName

	// Logstash required fields
	fields["@version"] = 1
	fields["@timestamp"] = time.Now().Format(time.RFC3339)

	// default to info level
	_, ok := fields["level"]
	if !ok {
		fields["level"] = "INFO"
	}

	serialized, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	return append(serialized, '\n'), nil
}
