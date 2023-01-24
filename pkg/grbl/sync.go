package grbl

import (
	"bufio"
	"bytes"
	"go.bug.st/serial"
	"strings"
)

// https://github.com/gnea/grbl/blob/master/doc/script/stream.py

func NewSyncConn(portName string, baudRate int) (*SyncConn, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	conn := &SyncConn{
		port:           port,
		responseBuffer: make(chan []byte),
		pushMsg:        make(chan []byte, 100),
	}

	return conn, nil
}

type SyncConn struct {
	port           serial.Port
	responseBuffer chan []byte
	pushMsg        chan []byte
	hold           bool
}

func (c *SyncConn) init() error {
	// tell grbl to wake up
	_, err := c.Write([]byte(WakeUp))
	return err
}

func (c *SyncConn) Start() error {
	if err := c.init(); err != nil {
		return err
	}
	// block until the reader exits (assuming it will when port is closed)
	return <-c.read()
}

func (c *SyncConn) Stop() {
	//c.port.Close()
}

func (c *SyncConn) Pushed() <-chan []byte {
	return c.pushMsg
}

func (c *SyncConn) Write(cmd []byte) ([]byte, error) {
	if c.hold && !IsStartResume(cmd) {
		return nil, nil
	}
	if _, err := c.port.Write(cmd); err != nil {
		return nil, err
	}
	if IsStartResume(cmd) {
		c.hold = false
	}
	if IsFeedHold(cmd) || c.hold {
		c.hold = true
		return nil, nil
	}
	// some commands don't return ok/error, so we can't do anything more here
	if !ExpectConfirmation(cmd) {
		return nil, nil
	}
	return <-c.responseBuffer, nil
}

// Basically we need to peak at all the messages being published and if they're push messages take them out of the request/response flow.
func (c *SyncConn) read() chan error {
	errExit := make(chan error)

	// there is no way to stop the scanner when it's blocked at Scan
	// so hopefully closing the underlying reader will kill it and if not we'll
	// just need to return.
	go func() {
		scanner := bufio.NewScanner(c.port)
		payload := bytes.Buffer{}
		for scanner.Scan() {
			// if a push message is received, publish it to a separate channel and try again
			// they can apparently appear at random :/
			// https://github.com/gnea/grbl/wiki/Grbl-v1.1-Interface#grbl-interface-basics
			if IsPushMsg(scanner.Bytes()) {
				select {
				case c.pushMsg <- scanner.Bytes():
				default:
					// channel is full - The client must not be reading them, so I guess we can discard.
				}
			} else {
				if scanner.Text() == "ok" || strings.HasPrefix(scanner.Text(), "error:") {
					payload.Write(scanner.Bytes())
					c.responseBuffer <- payload.Bytes()
					payload.Reset()
				} else {
					payload.Write(scanner.Bytes())
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errExit <- err
			return
		}
	}()

	return errExit
}
