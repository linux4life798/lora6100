// Code generated by "stringer -type=SerialDataBits"; DO NOT EDIT.

package lora6100

import "strconv"

const _SerialDataBits_name = "SerialDataBits7BitsSerialDataBits8Bits"

var _SerialDataBits_index = [...]uint8{0, 19, 38}

func (i SerialDataBits) String() string {
	i -= 1
	if i >= SerialDataBits(len(_SerialDataBits_index)-1) {
		return "SerialDataBits(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _SerialDataBits_name[_SerialDataBits_index[i]:_SerialDataBits_index[i+1]]
}
