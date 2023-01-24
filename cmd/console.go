package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/warmans/go-grbl/pkg/grbl"
	"os"
)

func main() {

	mode := flag.String("mode", "sync", "switch between sync or async connections")
	flag.Parse()

	switch *mode {
	case "sync":
		sync()
	case "async":
		async()
	}
}

func sync() {

	conn, err := grbl.NewSyncConn("/dev/ttyUSB0", 115200, true)
	if err != nil {
		panic(err)
	}
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter command: ")
		text, _ := reader.ReadString('\n')

		resp, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("ERR: ", err.Error())
			continue
		}
		fmt.Println("RESP: ", string(resp))

	}
}

func async() {

	conn, err := grbl.NewAsyncConn("/dev/ttyUSB0", 115200)
	if err != nil {
		panic(err)
	}

	go func() {
		for err := range conn.Errors() {
			fmt.Println("ERR ", err.Error())
		}
	}()

	go func() {
		for v := range conn.Read() {
			fmt.Println("READ ", string(v))
		}
	}()

	//go func() {
	//	for {
	//		conn.Write([]byte("?"))
	//		time.Sleep(time.Second)
	//	}
	//}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter command: ")
			text, _ := reader.ReadString('\n')
			conn.PriorityWrite([]byte(text))
		}
	}()

	if err := conn.Start(context.Background()); err != nil {
		panic(err)
	}
}
