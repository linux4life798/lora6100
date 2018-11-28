// Package lora6100 serves as a driver for interfacing with the Nice-RF
// Lora6100 AES module over UART.
//
// Craig Hesling
// Nov 19, 2018
package lora6100

import (
	"bytes"
	"time"

	"go.bug.st/serial.v1"
)

const (
	SettingsModeInDelay  = time.Millisecond * 20 // minimum seems to be 6ms
	SettingsModeOutDelay = time.Millisecond * 100
	DefaultBaudRate      = 9600
)

type LoRa6100 struct {
	serial.Port
	portName   string
	isopen     bool
	insettings bool // if we are in settings mode
}

func NewLoRa6100(portName string) *LoRa6100 {
	m := new(LoRa6100)
	m.portName = portName
	return m
}

func (m *LoRa6100) Open() error {
	p, err := serial.Open(m.portName, &serial.Mode{
		BaudRate: DefaultBaudRate,
	})
	if err != nil {
		return err
	}
	m.Port = p
	m.isopen = true

	// TODO:
	// We should try clearing the serial channel just incase other bytes had
	// been written before we opened the serial interface.
	// We could try to write a blank \r\n and expect a any line terminated
	// properly.
	if err := m.ResetInputBuffer(); err != nil {
		// allow port to remain open
		return err
	}
	if err := m.ResetOutputBuffer(); err != nil {
		// allow port to remain open
		return err
	}

	m.insettings = true
	if err := m.SettingsModeDisable(); err != nil {
		return err
	}

	return nil
}

func (m *LoRa6100) Close() error {
	m.isopen = false
	return m.Port.Close()
}

func (m *LoRa6100) ChangeBaudRate(baud SerialBaudRate) error {
	return m.SetMode(&serial.Mode{
		BaudRate: baud.GetSpeed(),
	})
}

func (m *LoRa6100) IsOpen() bool {
	return m.isopen
}

func (m *LoRa6100) SettingsModeEnable() error {
	if !m.insettings {
		m.insettings = true
		defer time.Sleep(SettingsModeInDelay)
		return m.SetRTS(true)
	}
	return nil
}

func (m *LoRa6100) SettingsModeDisable() error {
	if m.insettings {
		time.Sleep(SettingsModeInDelay)
		defer time.Sleep(SettingsModeOutDelay)
		m.insettings = false
		return m.SetRTS(false)
	}
	return nil
}

// GetLine reads until it sees \r\n
// If no error was returned, the buffer will contain the line without \r\n
func (m *LoRa6100) GetLine() (*bytes.Buffer, error) {
	var buf = new(bytes.Buffer)
	var b = make([]byte, 1)
	var crSeen, lfSeen bool

	for !(crSeen && lfSeen) {
		// read one byte
		if _, err := m.Read(b); err != nil {
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

func (m *LoRa6100) WriteLineEnding() error {
	_, err := m.Write([]byte(LineEnding))
	return err
}

func (m *LoRa6100) GetVersion() (string, error) {
	if !m.IsOpen() {
		return "", ErrUnopened
	}

	if err := m.SettingsModeEnable(); err != nil {
		return "", err
	}

	if _, err := CmdReadVersion.WriteTo(m); err != nil {
		return "", err
	}
	if err := m.WriteLineEnding(); err != nil {
		return "", err
	}
	resp, err := m.GetLine()
	if err != nil {
		return "", err
	}

	if err := m.SettingsModeDisable(); err != nil {
		return "", err
	}

	return resp.String(), nil
}

func (m *LoRa6100) GetParameters() (*Parameters, error) {
	if !m.IsOpen() {
		return nil, ErrUnopened
	}

	if err := m.SettingsModeEnable(); err != nil {
		return nil, err
	}

	var p = new(Parameters)

	if _, err := CmdReadParameters.WriteTo(m); err != nil {
		return nil, err
	}
	if err := m.WriteLineEnding(); err != nil {
		return nil, err
	}
	resp, err := m.GetLine()
	if err != nil {
		return nil, err
	}

	if err := m.SettingsModeDisable(); err != nil {
		return nil, err
	}

	_, err = p.ReadFrom(resp)

	return p, err
}
func (m *LoRa6100) ResetParameters() (RetStatus, error) {
	if !m.IsOpen() {
		return RetStatusError, ErrUnopened
	}

	if err := m.SettingsModeEnable(); err != nil {
		return RetStatusError, err
	}

	if _, err := CmdResetDefault.WriteTo(m); err != nil {
		return RetStatusError, err
	}
	if err := m.WriteLineEnding(); err != nil {
		return RetStatusError, err
	}
	resp, err := m.GetLine()
	if err != nil {
		return RetStatusError, err
	}
	var ret RetStatus
	if _, err := ret.ReadFrom(resp); err != nil {
		return RetStatusError, err
	}

	if err := m.SettingsModeDisable(); err != nil {
		return ret, err
	}

	return ret, nil
}

func (m *LoRa6100) SetParameters(p *Parameters) (RetStatus, error) {
	if !m.IsOpen() {
		return RetStatusError, ErrUnopened
	}

	if err := m.SettingsModeEnable(); err != nil {
		return RetStatusError, err
	}

	if _, err := CmdSetParameters.WriteTo(m); err != nil {
		return RetStatusError, err
	}
	if _, err := p.WriteTo(m); err != nil {
		return RetStatusError, err
	}
	if err := m.WriteLineEnding(); err != nil {
		return RetStatusError, err
	}
	resp, err := m.GetLine()
	if err != nil {
		return RetStatusError, err
	}
	var ret RetStatus
	if _, err := ret.ReadFrom(resp); err != nil {
		return RetStatusError, err
	}

	if err := m.SettingsModeDisable(); err != nil {
		return ret, err
	}

	return ret, nil
}
