package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

type LrscConnection struct {
	conn    io.ReadWriteCloser
	scanner *bufio.Reader
	dialer  Dialer
	StatusReporter
}

type lrscMessage struct {
	Deveui string
	Pdu    string
}

func (self *lrscMessage) toJson() string {
	json, err := json.Marshal(self)
	if err != nil {
		logger.Error("lrscMessage JSON marshaling failed: %v", err.Error())
	}
	return string(json)
}

type Dialer interface {
	Dial() (io.ReadWriteCloser, error)
	Endpoint() string
}

type TlsDialer struct {
	endpoint   string
	sslContext *tls.Config
	reporter   *StatusReporter
}

type dialerConfig struct {
	host, port, cert, key string
}

func (self *TlsDialer) Dial() (io.ReadWriteCloser, error) {
	return tls.Dial("tcp", self.endpoint, self.sslContext)
}

func (self *TlsDialer) Endpoint() string {
	return self.endpoint
}

func CreateTlsDialer(config dialerConfig, reporter *StatusReporter) (Dialer, error) {
	cert, err := ioutil.ReadFile(config.cert)
	if err != nil {
		return nil, fmt.Errorf("Could not read client certificate: %v", err)
	}
	key, err := ioutil.ReadFile(config.key)
	if err != nil {
		return nil, fmt.Errorf("Could not read client key: %v", err)
	}

	context := &tls.Config{InsecureSkipVerify: true}
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	context.Certificates = []tls.Certificate{certificate}
	endpoint := fmt.Sprintf("%v:%v", config.host, config.port)
	return &TlsDialer{endpoint: endpoint, sslContext: context, reporter: reporter}, nil
}

func (self *LrscConnection) StartListening(buffer chan lrscMessage) {
	go func() {
		self.connect()
		for {
			data, err := self.readLine()
			if err != nil {
				self.Report("CONNECTION", "Connection lost")
				logger.Error("read failed (%v)", err)
				self.connect()
				continue
			}

			message, err := parseLrscJson(data)
			if err != nil {
				logger.Error("Invalid message JSON received from LRSC (%v)\nMessage data: (%v)", err, data)
				continue
			}

			buffer <- message
		}
	}()
}

func (self *LrscConnection) Close() {
	if self.conn != nil {
		self.conn.Close()
	}

}

func (self *LrscConnection) connect() {
	for timeout := time.Second; ; timeout *= 2 {
		if timeout > time.Minute*5 {
			timeout = time.Minute * 5
		}
		err := self.establish()
		if err != nil {
			logger.Info(fmt.Sprintf("Connecting in %v seconds", timeout))
			time.Sleep(timeout)
		} else {
			self.Report("CONNECTION", "OK")
			break
		}
	}
}

func (self *LrscConnection) establish() error {
	if self.conn != nil {
		self.conn.Close()
	}

	logger.Debug("Attempting TCP connection")
	conn, err := self.dialer.Dial()

	if err != nil {
		logger.Error("Could not establish TCP connection")
		return err
	}
	logger.Info("Connected successfully")

	self.conn = conn
	self.scanner = bufio.NewReader(self.conn)
	//self.scanner = bufio.NewScanner(self.conn)

	err = self.handshake()
	if err != nil {
		logger.Error("Could not perform handshake: " + err.Error())
		return err
	}

	return nil
}

func (self *LrscConnection) handshake() error {
	err := self.send("JSON_000")
	if err != nil {
		logger.Debug("handshake failed: " + err.Error())
		return err
	}

	hello := `{"msgtag":1,"eui":"FF-00-00-00-00-00-00-00","euidom":0,"major":1,"minor":0,"build":0,"name":"LRSC Client"}` + "\n\n"
	err = self.send(hello)
	if err != nil {
		logger.Error("handshake failed: " + err.Error())
		return err
	}

	handshake, err := self.readLine()
	if err != nil {
		logger.Error("Did not receive ack in handshake: " + err.Error())
		return err
	}

	if validateHandshake(handshake) {
		logger.Info("handshake completed, connected to " + self.dialer.Endpoint())
	} else {
		logger.Error("Failed to validate handshake response")
	}

	return nil
}

func validateHandshake(handshake string) bool {
	return true
}

func parseLrscJson(data string) (lrscMessage, error) {
	var message lrscMessage
	err := json.Unmarshal([]byte(data), &message)
	logger.Info("Parsed Message: %v", message)
	return message, err
}

func (self *LrscConnection) send(message string) error {
	data := []byte(message)

	_, err := self.conn.Write(data)
	if err == nil {
		logger.Debug(">>> " + message)
	}
	return err
}

func (self *LrscConnection) readLine() (string, error) {
	for {
		data, _, err := self.scanner.ReadLine()
		if err != nil {
			return "", errors.New("failed to read message")
		}

		if len(data) == 0 {
			continue
		}

		message := string(data)
		logger.Debug("<<< " + message)
		return message, nil
	}
}
