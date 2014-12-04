package main

import (
	"crypto/tls"
	"fmt"
	"time"
)

type LrscConnection struct {
	endpoint string
	cert     *tls.Certificate
	conn     *tls.Conn
}

func CreateLrscConnection(hostname, port string, cert, key []byte) (*LrscConnection, error) {
	certificate, err := tls.X509KeyPair(cert, key)

	if err != nil {
		logger.Debug("Invalid certificate/key: " + err.Error())
		return nil, err
	}

	endpoint := fmt.Sprintf("%v:%v", hostname, port)

	lrscConnection := &LrscConnection{cert: &certificate, endpoint: endpoint}
	return lrscConnection, nil
}

func (self *LrscConnection) Connect() error {
	context := &tls.Config{InsecureSkipVerify: true}

	context.Certificates = []tls.Certificate{*self.cert}
	conn, err := tls.Dial("tcp", self.endpoint, context)

	if err != nil {
		logger.Debug("Could not initiate TCP connection: " + err.Error())
		return err
	}

	logger.Debug("connecting to LRSC endpoint...")
	self.conn = conn

	err = self.handshake()
	if err != nil {
		logger.Debug("Could not perform handshake: " + err.Error())
		return err
	}

	logger.Debug("handshake completed, connected to " + self.endpoint)
	return nil
}

func (self *LrscConnection) handshake() error {
	err := self.send("JSON_000")
	if err != nil {
		logger.Debug("handshake failed: " + err.Error())
		return err
	}

	hello := `{"msgtag":1,"eui":"FF-00-00-00-00-00-00-00","euidom":0,"major":1,"minor":0,"build":0,"name":"LRSC Client"}`
	err = self.send(hello)
	if err != nil {
		logger.Debug("handshake failed: " + err.Error())
		return err
	}

	err = self.send("\n\n")
	if err != nil {
		logger.Debug("handshake failed: " + err.Error())
		return err
	}

	// Handshake should get immediate response, so set short timeout
	self.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// All other messages could arrive any time, so reset after this function
	defer self.conn.SetReadDeadline(time.Time{})

	_, err = self.read()
	if err != nil {
		logger.Debug("Did not receive ack in handshake: " + err.Error())
		return err
	}

	return nil
}

func (self *LrscConnection) send(message string) error {
	data := []byte(message)

	_, err := self.conn.Write(data)
	if err == nil {
		logger.Debug(">>> " + message)
	}
	return err
}

func (self *LrscConnection) read() (string, error) {
	data := make([]byte, 4096)
	length, err := self.conn.Read(data)
	msg := string(data[0:length])
	logger.Debug("<<< " + msg)
	return msg, err
}
