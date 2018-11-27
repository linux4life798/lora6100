package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/linux4life798/lora6100"
)

type Message struct {
	ID  uint8
	TTL uint8
	Msg [4]byte
}

func (m *Message) WriteTo(out io.Writer) (int64, error) {
	n := int64(binary.Size(m))
	err := binary.Write(out, binary.BigEndian, m)
	return n, err
}

func (m *Message) ReadFrom(in io.Reader) (int64, error) {
	n := int64(binary.Size(m))
	err := binary.Read(in, binary.BigEndian, m)
	return n, err
}

func main() {
	info := flag.Bool("info", false, "Show hw version and params on startup (must have RTS connected to SET)")
	sendmsg := flag.String("msg", "", "The message to send. Must be 4 chars max.")
	flag.Parse()
	args := flag.Args()

	if len(*sendmsg) > len(Message{}.Msg) {
		panic("Provided message is too long")
	}

	portName := "/dev/ttyUSB0"
	if len(args) > 0 {
		portName = args[0]
	}
	log.Printf("Opening device %s\n", portName)
	l := lora6100.NewLoRa6100(portName)
	if err := l.Open(); err != nil {
		panic(err)
	}
	defer l.Close()

	if *info {
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

	var msg Message

	l.SettingsModeEnable()
	if len(*sendmsg) > 0 {
		log.Println("Sending first message")
		msg.ID = 45
		msg.TTL = 10
		copy(msg.Msg[:], []byte(*sendmsg))
		if _, err := msg.WriteTo(l); err != nil {
			panic(err)
		}
	}

	for {
		log.Println("Listening for messages")
		if _, err := msg.ReadFrom(l); err != nil {
			panic(err)
		}

		log.Printf("Read message: %+v\n", msg)

		if msg.TTL > 0 {
			msg.TTL--
			log.Printf("Sending message: %+v\n", msg)
			if _, err := msg.WriteTo(l); err != nil {
				panic(err)
			}
		}
	}
}