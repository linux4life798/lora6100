// Code generated by "stringer -type=SerialBaudRate"; DO NOT EDIT.

package lora6100

import "strconv"

const _SerialBaudRate_name = "SerialBaudRate1200SerialBaudRate2400SerialBaudRate4800SerialBaudRate9600SerialBaudRate14400SerialBaudRate19200SerialBaudRate38400SerialBaudRate57600SerialBaudRate76800SerialBaudRate115200SerialBaudRateUnknown"

var _SerialBaudRate_index = [...]uint8{0, 18, 36, 54, 72, 91, 110, 129, 148, 167, 187, 208}

func (i SerialBaudRate) String() string {
	if i >= SerialBaudRate(len(_SerialBaudRate_index)-1) {
		return "SerialBaudRate(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SerialBaudRate_name[_SerialBaudRate_index[i]:_SerialBaudRate_index[i+1]]
}
