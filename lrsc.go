package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"io"
	"io/ioutil"
)

type lrscConnection struct {
	conn   io.ReadWriteCloser
	reader *bufio.Reader
	dialer dialer
	reporter.StatusReporter
	inbound chan lrscMessage
	err     chan error
}

type lrscMessage struct {
	Deveui, Pdu string
}

func (self *lrscMessage) toJson() string {
	json, err := json.Marshal(self)
	if err != nil {
		logger.Error("lrscMessage JSON marshaling failed: %v", err.Error())
	}
	return string(json)
}

type dialer interface {
	dial() (io.ReadWriteCloser, error)
	endpoint() string
}

type tlsDialer struct {
	raddr      string
	sslContext *tls.Config
	reporter   *reporter.StatusReporter
}

type dialerConfig struct {
	host, port, cert, key string
}

func (self *tlsDialer) dial() (io.ReadWriteCloser, error) {
	return tls.Dial("tcp", self.raddr, self.sslContext)
}

func (self *tlsDialer) endpoint() string {
	return self.raddr
}

func createTlsDialer(config dialerConfig, reporter *reporter.StatusReporter) (dialer, error) {
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
	return &tlsDialer{raddr: endpoint, sslContext: context, reporter: reporter}, nil
}

func (self *lrscConnection) Connect() error {
	err := self.establish()
	return err
}

func (self *lrscConnection) Loop() {
	for {
		line, err := self.readLine()
		if err != nil {
			self.err <- err
			break
		} else {
			message, err := parseLrscMessage(string(line))
			if err != nil {
				self.err <- err
				break
			}

			self.inbound <- message
		}
	}
}

func (self *lrscConnection) Error() <-chan error {
	return self.err
}

func (self *lrscConnection) close() {
	if self.conn != nil {
		self.conn.Close()
	}
}

func (self *lrscConnection) establish() error {
	self.close()

	logger.Debug("Attempting TCP connection")
	conn, err := self.dialer.dial()

	if err != nil {
		logger.Error("Could not establish TCP connection: %v", err)
		self.Report("CONNECTION", err.Error())
		return err
	}
	logger.Info("Connected successfully")

	self.conn = conn
	self.reader = bufio.NewReader(self.conn)

	err = self.handshake()
	if err != nil {
		logger.Error("Could not perform handshake: " + err.Error())
		return err
	}

	self.Report("CONNECTION", "OK")
	return nil
}

func (self *lrscConnection) handshake() error {
	err := self.send("JSON_000")
	if err != nil {
		logger.Debug("handshake failed: " + err.Error())
		return err
	}

	hello := `{"msgtag":1,"eui":"FF-00-00-00-00-00-00-00","euidom":0,"major":1,"minor":0,"build":0,"name":"LRSC Client"}`
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
		logger.Info("handshake completed, connected to " + self.dialer.endpoint())
	} else {
		logger.Error("Failed to validate handshake response")
	}

	return nil
}

func (self *lrscConnection) send(message string) error {
	data := []byte(message + "\n\n")

	_, err := self.conn.Write(data)
	if err == nil {
		logger.Debug(">>> " + message)
	}
	return err
}

func (self *lrscConnection) readLine() (string, error) {
	for {
		data, _, err := self.reader.ReadLine()
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

func parseLrscMessage(data string) (lrscMessage, error) {
	var message lrscMessage
	err := json.Unmarshal([]byte(data), &message)
	//logger.Debug("Parsed Message: %v", message)
	return message, err
}

func validateHandshake(handshake string) bool {
	return true
}
