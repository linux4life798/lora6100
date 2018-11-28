package main

import (
	CRAND "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/linux4life798/lora6100"
)

const (
	SleepBetweenTX = time.Millisecond * 50
	MsgTTL         = 10
)

type Message struct {
	ID  uint8
	TTL uint8
	Msg [40]byte
}

func (m *Message) String() string {
	return fmt.Sprintf("ID=%v TTL=%v MSG=\"%s\"", m.ID, m.TTL, string(m.Msg[:]))
}

func (m *Message) WriteTo(out io.Writer) (int64, error) {
	n := int64(binary.Size(m))
	if n > 62 {
		panic("Packet is too large")
	}
	err := binary.Write(out, binary.BigEndian, m)
	return n, err
}

func (m *Message) ReadFrom(in io.Reader) (int64, error) {
	n := int64(binary.Size(m))
	if n > 62 {
		panic("Packet is too large")
	}
	err := binary.Read(in, binary.BigEndian, m)
	return n, err
}

func (m *Message) RandID() {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100))
	if err != nil {
		panic(err)
	}
	m.ID = uint8(r.Int64())
}

func randomDelay(maxDelay time.Duration) time.Duration {
	if maxDelay == 0 {
		return 0
	}

	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(int64(maxDelay)))
	if err != nil {
		panic(err)
	}
	return time.Duration(r.Int64())
}

func main() {
	info := flag.Bool("info", false, "Show hw version and params on startup (must have RTS connected to SET)")
	sendmsg := flag.String("msg", "", "The message to send. Must be 4 chars max.")
	randdelay := flag.Duration("rdelay", time.Duration(0), "Specifies the random delay before retransmission")
	datarate := flag.Uint("dr", 3, "Select the datarate [0 to 9]")
	inputmsg := flag.Bool("imsg", false, "Enables the console message send feature")
	baud := flag.Int("baud", 9600, "The serial baud rate for data transfer")

	flag.Parse()
	args := flag.Args()

	if len(*sendmsg) > len(Message{}.Msg) {
		panic("Provided message is too long")
	}

	if *datarate > 9 {
		panic("Datarate must range from 0 to 9 (inclusive)")
	}

	portName := "/dev/ttyUSB0"

	if _, err := os.Stat("/dev/ttyAMA0"); err == nil {
		portName = "/dev/ttyAMA0"
	}

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

		var newp = *p

		newp.RFDataRate = byte(*datarate)

		newp.SerialBaud.FromSpeed(*baud)
		if newp.SerialBaud == lora6100.SerialBaudRateUnknown {
			panic("Invalid baud rate")
		}

		if newp != *p {
			r, err := l.SetParameters(&newp)
			if err != nil {
				panic(err)
			}
			fmt.Printf("SetParameters: %+v | RetStatus=%v\n", newp, r)
		}

		if err := l.ChangeBaudRate(newp.SerialBaud); err != nil {
			panic(err)
		}

	}

	inbound := make(chan Message, 1000)
	outbound := make(chan Message) // launches go routine for all sends

	go func() {
		log.Println("Listening for inbound messages")

		for {
			var msg Message
			if _, err := msg.ReadFrom(l); err != nil {
				panic(err)
			}
			inbound <- msg
		}

	}()

	go func() {
		log.Println("Started outbound thread")

		for msg := range outbound {
			log.Printf("TX: %s | Firing!", msg.String())
			if _, err := msg.WriteTo(l); err != nil {
				panic(err)
			}
			time.Sleep(SleepBetweenTX)
		}
	}()

	send := func(msg Message, delay time.Duration) {
		log.Printf("Sending message: %s in %v\n", msg.String(), delay)
		go func() {
			time.Sleep(delay)
			outbound <- msg
		}()
	}

	if *inputmsg {
		go func() {
			log.Println("Launching console message scanner")
			for {
				var line string
				fmt.Scanln(&line)
				log.Printf("Console msg: %s", line)
				var msg Message
				msg.TTL = MsgTTL
				msg.RandID()
				copy(msg.Msg[:], line) // will copy at most len(msg.Msg)
				send(msg, randomDelay(*randdelay))
			}
		}()
	}

	if len(*sendmsg) > 0 {
		var msg Message
		msg.TTL = MsgTTL
		msg.RandID()
		copy(msg.Msg[:], []byte(*sendmsg))
		send(msg, 0)
	}

	for msg := range inbound {
		log.Printf("RX: %s\n", msg.String())

		if msg.TTL > 0 {
			msg.TTL--
			send(msg, randomDelay(*randdelay))
		}

	}
}
