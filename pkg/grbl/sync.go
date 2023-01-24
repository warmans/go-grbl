package grbl

import (
	"bufio"
	"bytes"
	"fmt"
	"go.bug.st/serial"
	"strings"
)

// https://github.com/gnea/grbl/blob/master/doc/script/stream.py

func NewSyncConn(portName string, baudRate int, debug bool) (*SyncConn, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	conn := &SyncConn{
		port:  port,
		debug: debug,
	}
	return conn, conn.Init()
}

type SyncConn struct {
	port  serial.Port
	debug bool
}

func (c *SyncConn) Init() error {
	// tell grbl to wake up
	resp, err := c.Write([]byte("\r\n\r\n"))
	if err != nil {
		return err
	}
	fmt.Println(string(resp))
	return nil
}

func (c *SyncConn) Write(msg []byte) ([]byte, error) {
	if c.debug {
		var prefix = "CMD: "
		if IsRealtimeCommand(msg) {
			prefix = "REALTIME: "
		}
		fmt.Println(prefix + string(msg))
	}
	if _, err := c.port.Write(msg); err != nil {
		return nil, err
	}
	if IsFeedHold(msg) {
		return []byte{}, nil
	}
	scanner := bufio.NewScanner(c.port)
	buff := bytes.Buffer{}
	for {
		// todo: timeout
		for scanner.Scan() {
			text := scanner.Text()
			if strings.HasPrefix(text, "ok") {
				if buff.Len() == 0 {
					return []byte(text), nil
				}
				return buff.Bytes(), nil
			}
			if strings.HasPrefix(text, "error") {
				return nil, fmt.Errorf("error: %s", buff.String())
			}
			// welcome text and realtime don't come with an ok/error (allegedly...)
			if strings.HasPrefix(text, "Grbl ") || IsRealtimeCommand(msg) {
				buff.WriteString(text)
				return buff.Bytes(), nil
			}
			buff.WriteString(text)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
}
