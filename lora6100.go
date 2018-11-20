// Package lora6100 serves as a driver for interfacing with the Nice-RF
// Lora6100 AES module over UART.
//
// Craig Hesling
// Nov 19, 2018
package lora6100

import (
	"bytes"

	"github.com/tarm/serial"
)

type LoRa6100 struct {
	c *serial.Config
	s *serial.Port
}

func NewLoRa6100(port string) *LoRa6100 {
	m := new(LoRa6100)
	m.c = &serial.Config{
		Name: port,
		Baud: 9600,
	}
	return m
}

func (m *LoRa6100) Open() error {
	s, err := serial.OpenPort(m.c)
	if err != nil {
		m.s = nil
		return err
	}
	m.s = s

	// TODO:
	// We should try clearing the serial channel just incase other bytes had
	// been written before we opened the serial interface.
	// We could try to write a blank \r\n and expect a any line terminated
	// properly.

	return nil
}

func (m *LoRa6100) IsOpen() bool {
	return m.s != nil
}

// getLine reads until it sees \r\n
// If no error was returned, the buffer will contain the line without \r\n
func (m *LoRa6100) getLine() (*bytes.Buffer, error) {
	var buf = new(bytes.Buffer)
	var b = make([]byte, 1)
	var crSeen, lfSeen bool

	for !(crSeen && lfSeen) {
		// read one byte
		if _, err := m.s.Read(b); err != nil {
			return buf, err
		}

		switch b[0] {
		case '\r':
			// '\n' couldn't have been seen before this point
			crSeen = true
		case '\n':
			// '\r' should have already been seen
			if !crSeen {
				buf.WriteByte(b[0]) // if corrupt, include offending byte
				return buf, ErrMalformedLine
			}
			lfSeen = true
		default:
			buf.WriteByte(b[0]) // if corrupt, include offending byte
			if crSeen || lfSeen {
				return buf, ErrMalformedLine
			}
		}

	}

	return buf, nil
}

func (m *LoRa6100) writeLineEnding() error {
	_, err := m.s.Write([]byte(LineEnding))
	return err
}

func (m *LoRa6100) GetVersion() (string, error) {
	if !m.IsOpen() {
		return "", ErrUnopened
	}

	if err := CmdReadVersion.WriteTo(m.s); err != nil {
		return "", err
	}
	if err := m.writeLineEnding(); err != nil {
		return "", err
	}
	resp, err := m.getLine()
	if err != nil {
		return "", err
	}

	return resp.String(), nil
}

func (m *LoRa6100) GetParameters() (*Parameters, error) {
	if !m.IsOpen() {
		return nil, ErrUnopened
	}

	var p = new(Parameters)

	if err := CmdReadParameters.WriteTo(m.s); err != nil {
		return nil, err
	}
	if err := m.writeLineEnding(); err != nil {
		return nil, err
	}
	resp, err := m.getLine()
	if err != nil {
		return nil, err
	}

	err = p.ReadFrom(resp)

	return p, err
}
func (m *LoRa6100) ResetParameters() (RetStatus, error) {
	if !m.IsOpen() {
		return RetStatusError, ErrUnopened
	}

	if err := CmdResetDefault.WriteTo(m.s); err != nil {
		return RetStatusError, err
	}
	if err := m.writeLineEnding(); err != nil {
		return RetStatusError, err
	}
	resp, err := m.getLine()
	if err != nil {
		return RetStatusError, err
	}
	var ret RetStatus
	if err := ret.ReadFrom(resp); err != nil {
		return RetStatusError, err
	}

	return ret, nil
}

func (m *LoRa6100) SetParameters(p *Parameters) (RetStatus, error) {
	if !m.IsOpen() {
		return RetStatusError, ErrUnopened
	}

	if err := CmdSetParameters.WriteTo(m.s); err != nil {
		return RetStatusError, err
	}
	if err := p.WriteTo(m.s); err != nil {
		return RetStatusError, err
	}
	if err := m.writeLineEnding(); err != nil {
		return RetStatusError, err
	}
	resp, err := m.getLine()
	if err != nil {
		return RetStatusError, err
	}
	var ret RetStatus
	if err := ret.ReadFrom(resp); err != nil {
		return RetStatusError, err
	}

	return ret, nil
}
