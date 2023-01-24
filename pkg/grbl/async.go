package grbl

import (
	"bufio"
	"context"
	"fmt"
	"go.bug.st/serial"
)

func NewAsyncConn(portName string, baudRate int) (*AsyncConn, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return &AsyncConn{
		port:                    port,
		highPriorityWriteBuffer: make(chan []byte, 100),
		writeBuffer:             make(chan []byte, 100),
		readBuffer:              make(chan []byte, 100),
		errors:                  make(chan error, 100),
	}, nil
}

// AsyncConn is just a fully async rw connection. This makes it hard to track
// which specific commands have been executed.
type AsyncConn struct {
	port                    serial.Port
	writeBuffer             chan []byte
	highPriorityWriteBuffer chan []byte
	readBuffer              chan []byte
	errors                  chan error
}

func (c *AsyncConn) Start(ctx context.Context) error {
	defer func() {
		c.port.Close()
		close(c.writeBuffer)
		close(c.highPriorityWriteBuffer)
		close(c.readBuffer)
		close(c.errors)
	}()

	go func() {
		if err := c.startSerialReader(ctx); err != nil {
			c.errors <- err
		}
	}()

	return c.startSerialWriter(ctx)
}

func (c *AsyncConn) startSerialReader(ctx context.Context) error {
	responseScanner := bufio.NewScanner(c.port)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := responseScanner.Err(); err != nil {
				return fmt.Errorf("response scanner failed: %w", err)
			}
			// I'm hoping this will be unblocked by c.port.Close()
			if responseScanner.Scan() {
				c.readBuffer <- []byte(responseScanner.Text())
			} else {
				return responseScanner.Err()
			}
		}
	}
}

func (c *AsyncConn) startSerialWriter(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// allow some messages to jump the queue e.g. e-stop or status
		select {
		case msgBytes := <-c.highPriorityWriteBuffer:
			_, err := c.port.Write(msgBytes)
			if err != nil {
				c.errors <- err
			}
		default:
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msgBytes := <-c.highPriorityWriteBuffer:
				_, err := c.port.Write(msgBytes)
				if err != nil {
					c.errors <- err
				}
			case msgBytes := <-c.writeBuffer:
				_, err := c.port.Write(msgBytes)
				if err != nil {
					c.errors <- err
				}
			}
		}
	}
}

func (c *AsyncConn) Errors() chan error {
	return c.errors
}

func (c *AsyncConn) Read() chan []byte {
	return c.readBuffer
}

func (c *AsyncConn) Write(msg []byte) {
	c.writeBuffer <- msg
}

func (c *AsyncConn) PriorityWrite(msg []byte) {
	c.highPriorityWriteBuffer <- msg
}
