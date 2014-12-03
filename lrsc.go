package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"time"
)

type LrscConnection struct {
	endpoint string
	cert     *tls.Certificate
	conn     *tls.Conn
}

func CreateLrscConnection(hostname, port string, cert, key []byte) *LrscConnection {
	certificate, err := tls.X509KeyPair(cert, key)

	if err != nil {
		logger.Panic(err)
	}

	endpoint := fmt.Sprintf("%v:%v", hostname, port)

	lrscConnection := &LrscConnection{cert: &certificate, endpoint: endpoint}
	return lrscConnection
}

func (self *LrscConnection) Connect() {
	context := &tls.Config{InsecureSkipVerify: true}

	context.Certificates = []tls.Certificate{*self.cert}
	conn, err := tls.Dial("tcp", self.endpoint, context)

	if err != nil {
		logger.Panic(err)
	}

	logger.Printf("connecting to LRSC endpoint...")
	self.conn = conn

	self.handshake()

	logger.Printf("handshake completed, connected to %v", self.endpoint)
}

func (self *LrscConnection) handshake() {
	hello := `{"msgtag":1,"eui":"FF-00-00-00-00-00-00-00","euidom":0,"major":1,"minor":0,"build":0,"name":"LRSC Client"}`
	err1 := self.send("JSON_000")
	err2 := self.send(hello)
	err3 := self.send("\n\n")

	if err1 != nil || err2 != nil || err3 != nil {
		logger.Panic("Handshake failed")
	}

	_, err := self.read()
	if err != nil {
		logger.Panic(fmt.Printf("Did not receive ack in handshake (%v)", err))
	}
}

func (self *LrscConnection) send(message string) error {
	data := []byte(message)

	_, err := self.conn.Write(data)
	if err == nil {
		logger.Printf(">>> %v", message)
	}
	return err
}

func (self *LrscConnection) read() (string, error) {
	data := make([]byte, 4096)
	self.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	length, err := self.conn.Read(data)
	msg := string(data[0:length])
	logger.Printf("<<< %v", msg)
	return msg, err
}

func readCertificate(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Panic(err)
	}

	return data
}
