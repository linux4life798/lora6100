package main

import (
	"fmt"
	"log"
	"os"
	"time"

	serial "go.bug.st/serial.v1"
)

func main() {

	portName := "/dev/ttyUSB0"

	if len(os.Args) > 1 {
		portName = os.Args[1]
	}

	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("setting: true")
	if err := port.SetRTS(true); err != nil {
		panic(err)
	}
	fmt.Println("Waiting")
	time.Sleep(time.Millisecond * 2000)
	fmt.Println("setting: false")
	if err := port.SetRTS(false); err != nil {
		panic(err)
	}
	fmt.Println("Waiting")
	time.Sleep(time.Millisecond * 2000)

	fmt.Println("setting: true")
	if err := port.SetRTS(true); err != nil {
		panic(err)
	}
	fmt.Println("Waiting")
}
