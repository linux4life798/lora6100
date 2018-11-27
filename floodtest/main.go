package main

import (
	CRAND "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"time"

	"github.com/linux4life798/lora6100"
)

type Message struct {
	ID  uint8
	TTL uint8
	Msg [40]byte
}

func (m *Message) String() string {
	return fmt.Sprintf("ID=%v TTL=%v MSG=\"%s\"\n", m.ID, m.TTL, string(m.Msg[:]))
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
	randdelay := flag.Duration("rdelay", time.Duration(0), "Specifies the random delay before retransmission")
	datarate := flag.Uint("dr", 3, "Select the datarate [0 to 9]")
	flag.Parse()
	args := flag.Args()

	if len(*sendmsg) > len(Message{}.Msg) {
		panic("Provided message is too long")
	}

	if *datarate > 9 {
		panic("Datarate must range from 0 to 9 (inclusive)")
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

		if uint(p.RFDataRate) != *datarate {
			p.RFDataRate = byte(*datarate)
			r, err := l.SetParameters(p)
			if err != nil {
				panic(err)
			}
			fmt.Printf("SetParameters: %+v | RetStatus=%v\n", *p, r)
		}

	}

	var msg Message

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

		log.Printf("Read message: %s\n", msg.String())

		if msg.TTL > 0 {
			msg.TTL--
			var delay time.Duration
			if *randdelay != time.Duration(0) {
				r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(int64(*randdelay)))
				if err != nil {
					panic(err)
				}
				delay = time.Duration(r.Int64())
			}
			log.Printf("Sending message: %s in %v\n", msg.String(), delay)
			time.Sleep(delay)
			if _, err := msg.WriteTo(l); err != nil {
				panic(err)
			}
		}
	}
}
