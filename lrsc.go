package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"time"
)

type LrscConnection struct {
	endpoint   string
	conn       *tls.Conn
	sslContext *tls.Config
	scanner    *bufio.Scanner
}

func CreateLrscConnection(hostname, port string, cert, key []byte) (*LrscConnection, error) {
	certificate, err := tls.X509KeyPair(cert, key)

	if err != nil {
		logger.Err("Invalid certificate/key: " + err.Error())
		return nil, err
	}

	endpoint := fmt.Sprintf("%v:%v", hostname, port)

	context := &tls.Config{InsecureSkipVerify: true}
	context.Certificates = []tls.Certificate{certificate}

	lrscConnection := &LrscConnection{endpoint: endpoint, sslContext: context}

	return lrscConnection, nil
}

func (self *LrscConnection) StartListening(buffer chan string) {
	go func() {
		self.connect()
		for {
			message, err := self.readLine()
			if err != nil {
				logger.Err("read failed: " + err.Error())
				self.connect()
				continue
			}

			if len(message) == 0 {
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
			break
		}
	}
}

func (self *LrscConnection) establish() error {
	if self.conn != nil {
		self.conn.Close()
	}
	conn, err := tls.Dial("tcp", self.endpoint, self.sslContext)
	if err != nil {
		logger.Err("Could not establish TCP connection")
		return err
	}
	logger.Info("Connected successfully")

	self.conn = conn
	self.scanner = bufio.NewScanner(self.conn)

	err = self.handshake()
	if err != nil {
		logger.Err("Could not perform handshake: " + err.Error())
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

	hello := `{"msgtag":1,"eui":"FF-00-00-00-00-00-00-00","euidom":0,"major":1,"minor":0,"build":0,"name":"LRSC Client"}`
	err = self.send(hello)
	if err != nil {
		logger.Err("handshake failed: " + err.Error())
		return err
	}

	err = self.send("\n\n")
	if err != nil {
		logger.Err("handshake failed: " + err.Error())
		return err
	}

	// Handshake should get immediate response, so set short timeout
	self.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// All other messages could arrive any time, so reset after this function
	defer self.conn.SetReadDeadline(time.Time{})

	_, err = self.readLine()
	if err != nil {
		logger.Err("Did not receive ack in handshake: " + err.Error())
		return err
	}

	logger.Info("handshake completed, connected to " + self.endpoint)

	return nil
}

func (self *LrscConnection) Listen() (chan string, error) {
	messages := make(chan string)
	return messages, nil
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
	status := self.scanner.Scan()
	if !status {
		logger.Warning("read from socket failed: " + self.scanner.Err().Error())
		return "", errors.New("failed to read message")
	}

	message := self.scanner.Text()

	logger.Debug("<<< " + message)
	return message, nil
}
