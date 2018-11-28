//go:generate stringer -type=SerialBaudRate
//go:generate stringer -type=SerialDataBits
//
package lora6100

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrUnopened        = errors.New("The device has not been opened for communicaiton")
	ErrMalformedLine   = errors.New("Read a corrupt line")
	ErrBadReturnStatus = errors.New("Received a bad return status")
)

var (
	CmdPrefix  = []byte{0xAA, 0xFA}
	LineEnding = "\r\n"
)

type Cmd byte

func (c Cmd) WriteTo(out io.Writer) (int64, error) {
	var buf bytes.Buffer
	buf.Write(CmdPrefix)
	buf.WriteByte(byte(c))
	n, err := out.Write(buf.Bytes())
	return int64(n), err
}

const (
	CmdReadVersion    Cmd = 0xAA
	CmdReadParameters Cmd = 0x01
	CmdResetDefault   Cmd = 0x02
	CmdSetParameters  Cmd = 0x03
)

type SerialBaudRate byte

const (
	SerialBaudRate1200 SerialBaudRate = iota
	SerialBaudRate2400
	SerialBaudRate4800
	SerialBaudRate9600
	SerialBaudRate14400
	SerialBaudRate19200
	SerialBaudRate38400
	SerialBaudRate57600
	SerialBaudRate76800
	SerialBaudRate115200
	SerialBaudRateUnknown
)

var speedToBaud = map[int]SerialBaudRate{
	1200:   SerialBaudRate1200,
	2400:   SerialBaudRate2400,
	4800:   SerialBaudRate4800,
	9600:   SerialBaudRate9600,
	14400:  SerialBaudRate14400,
	19200:  SerialBaudRate19200,
	38400:  SerialBaudRate38400,
	57600:  SerialBaudRate57600,
	76800:  SerialBaudRate76800,
	115200: SerialBaudRate115200,
}

func (b *SerialBaudRate) FromSpeed(speed int) {
	s, ok := speedToBaud[speed]
	if !ok {
		*b = SerialBaudRateUnknown
	} else {
		*b = s
	}
}

func (b SerialBaudRate) GetSpeed() int {
	for speed, baud := range speedToBaud {
		if baud == b {
			return speed
		}
	}
	return -1
}

type SerialDataBits byte

const (
	SerialDataBits7Bits SerialDataBits = iota
	SerialDataBits8Bits
)

// 31 bytes total
type Parameters struct {
	RFChannel      byte
	RFFreq         byte
	RFDataRate     byte           // 0-9
	TXPower        byte           // 0-7
	SerialBaud     SerialBaudRate // 0-9
	SerialDataBits SerialDataBits // 1 (7bit) or 2(8bits)
	SerialStopBits byte           // 1 (1bit) or 2 (2bits)
	SerialParity   byte           // 1 (none), 2 (odd), 3 (even)
	NetID          uint32         // 0x00000000 - 0xFFFFFFFF
	NodeID         uint16         // 0x0000 - 0xFFFF
	AESKeySetting  byte           // 0 (default-does not use defined AES key) or 1 (user-defined AES key)
	AESKey         [16]byte       // AES 128 key
}

func (p *Parameters) ReadFrom(in io.Reader) (int64, error) {
	// if len(data) != 31 {
	// 	return fmt.Errorf("Payload is incorrect size. Must be 31 bytes.")
	// }
	// p.RFChannel = data[0]
	// p.RFFreq = data[1]
	// p.RFDataRate = data[2]
	// p.TXPower = data[3]
	// p.SerialBaud = data[4]
	// p.SerialDataBits = data[5]
	// p.SerialStopBits = data[6]
	// p.SerialParity = data[7]
	// // p.NetID  = data[8]
	// // p.NodeID =  = data[9]
	// p.AESKeySetting = data[10]
	// // p.AESKey = data[11]
	n := int64(binary.Size(p))
	return n, binary.Read(in, binary.BigEndian, p)
}

func (p *Parameters) WriteTo(out io.Writer) (int64, error) {
	n := int64(binary.Size(p))
	return n, binary.Write(out, binary.BigEndian, p)
}

type RetStatus string

const (
	RetStatusOk    RetStatus = "OK"
	RetStatusError RetStatus = "ERROR"
)

func (r *RetStatus) ReadFrom(in io.Reader) (int64, error) {
	var buf bytes.Buffer
	n, err := buf.ReadFrom(in)
	if err != nil {
		return n, err
	}
	switch buf.String() {
	case string(RetStatusOk):
		*r = RetStatusOk
	case string(RetStatusError):
		*r = RetStatusError
	default:
		*r = RetStatusError
		return n, ErrBadReturnStatus
	}

	return n, nil
}
