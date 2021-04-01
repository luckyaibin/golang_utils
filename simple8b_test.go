package main

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestEncode(t *testing.T) {
	values := []uint64{0, 1143, 320, 311, 170, 301, 170, 140, 140, 281}
	v, n, e := Encode(values)
	fmt.Println("编码后:", v, n, e)
	dst := [240]uint64{}
	p := (*[240]uint64)(unsafe.Pointer(&dst[0]))
	m := Decode(v, p)
	fmt.Println("解码后:", m, dst)

	encodedList, err := EncodeList(values)
	fmt.Println("编码后:", encodedList, err)

	decodedList := DecodeList(encodedList)
	fmt.Println("解码后:", decodedList)
}

func TestEncoder(t *testing.T) {

	for N := 1; N < 100000000; N += 1 {
		//先编码
		enc := NewEncoder()
		encodeList := make([]uint64, 0)
		for i := 0; i < N; i++ {
			encodeList = append(encodeList, uint64(i))
			err := enc.Write(uint64(i))
			if err != nil {
				t.FailNow()
			}
		}
		bytes, err := enc.Bytes()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		//再解码
		dec := NewDecoder(bytes)
		decodedList := make([]uint64, 0)
		for dec.Next() {
			v := dec.Read()
			decodedList = append(decodedList, v)
		}
		if len(decodedList) != len(encodeList) {
			t.Failed()
			return
		}
		if N > 100000 {
			fmt.Println(N)
		}
		for i := 0; i < len(decodedList); i++ {
			if decodedList[i] != encodeList[i] {
				fmt.Println("不相等")
				t.FailNow()
				return
			}
		}
	}
}
