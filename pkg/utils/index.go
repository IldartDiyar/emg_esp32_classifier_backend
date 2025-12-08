package utils

import (
	"encoding/binary"
	"strconv"
	"strings"
)

func IntSliceToBytea(samples []int) []byte {
	if len(samples) == 0 {
		return nil
	}

	buf := make([]byte, len(samples)*2)

	for i, v := range samples {
		if v < -32768 {
			v = -32768
		} else if v > 32767 {
			v = 32767
		}
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(int16(v)))
	}

	return buf
}

func ByteaToIntSlice(b []byte) []int {
	out := make([]int, len(b))
	for i := range b {
		out[i] = int(b[i])
	}
	return out
}

func IntSliceToString(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	strs := make([]string, len(nums))
	for i, v := range nums {
		strs[i] = strconv.Itoa(v)
	}
	return strings.Join(strs, ",")
}

func DecodeRawBytes(b []byte) []int {
	if len(b)%2 != 0 {
		return []int{}
	}

	out := make([]int, len(b)/2)

	for i := 0; i < len(b); i += 2 {
		val := binary.LittleEndian.Uint16(b[i : i+2])
		out[i/2] = int(val)
	}

	return out
}
