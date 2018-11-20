package main

import (
	"fmt"

	"github.com/linux4life798/lora6100"
)

func main() {
	l := lora6100.NewLoRa6100("/dev/ttyUSB0")
	if err := l.Open(); err != nil {
		panic(err)
	}
	ver, err := l.GetVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println("Version:", ver)

	p, err := l.GetParameters()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parameters: %+v\n", *p)
}
