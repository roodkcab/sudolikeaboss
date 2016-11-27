package onepass

import (
	"strings"
	"fmt"
	"encoding/binary"
	"bytes"
	"encoding/hex"
	"math/rand"
)

type Codec struct { }

func (Codec)generateRandomBytesArray(length int) []byte {
	randBytes := []byte{}
	for i := 0; i < length/4; i++ {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, rand.Int31())
		randBytes = append(randBytes, buf.Bytes()...)
	}
	return randBytes
}

func (Codec)toBits(data string, b bool) []byte {
	a := strings.Replace(data, "=", "", -1)
	var c = []byte{}
	var d = uint(0)
	var e = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var f = int32(0)
	if b {
		e = e[0:len(e)-2] + "-_"
	}

	for i := 0; i < len(a); i++ {
		g := int32(strings.Index(e, string(a[i])))
		if g < 0 {
			return []byte{}
		}
		if d > 26 {
			d -= 26
			part := int32(f) ^ int32(uint32(g) >> d)
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, part)
			c = append(c, buf.Bytes()...)
			f = g << (32 - d)
		} else {
			d += 6
			f ^= g << (32 - d)
		}
	}
	if (d & 56 > 0) {
		if (d & 56) == 32 {
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, f)
			c = append(c, buf.Bytes()...)
		} else {
			fmt.Println("!", d & 56)
		}
	}
	return c
}


func toInt32(data []byte) int32 {
	return int32(uint32(data[0])<<24 + uint32(data[1])<<16 + uint32(data[2])<<8 + uint32(data[3]))
}

func (Codec)fromBits(a []byte, b bool, c bool) (string) {
	var d = ""
	var e = uint32(0)
	var f = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var g = uint32(0)
	var h = len(hex.EncodeToString(a)) * 4 //sjcl.bitArray.bitLength(a);
	if (c) {
		f = f[0: len(f) - 2] + "-_"
	}

	data := []int32{}
	for i :=0; i <= len(a) - 4; i+= 4 {
		data = append(data, toInt32(a[i:i+4]))
	}

	for i := 0; len(d) * 6 < h; {
		//d += f.charAt((g ^ a[c] >>> e) >>> 26);
		bytes := uint32(0)
		if (i < len(data)) {
			bytes = uint32(data[i])
		}
		d += string(f[(uint32(g ^ (bytes >> e)) >> 26)]);
		if (e < 6) {
			g = bytes << (6 - e)
			e += 26
			i++
		} else {
			g <<= 6
			e -= 6
		}
	}
	for ; len(d) & 3 > 0 && !b; {
		d += "="
	}
	return d
}


