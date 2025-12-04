package utils

import (
	"encoding/binary"
	"encoding/json"
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
	var out []int
	_ = json.Unmarshal(b, &out)
	return out
}
