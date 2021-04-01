package main

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type Encoder struct {
	buf   *RingBuffer //缓存将要被编码的数字序列
	bytes []byte      //缓存buf被编码后的字节数组
}

func NewEncoder() *Encoder {
	return &Encoder{
		buf:   NewRingBuffer(512), //为什么取512?因为240个1可能在任何位置出现,我们保证所有240个1出现在任何位置都能在一次编码中获取到,所以至少是480，并且满足2^n.所以是2^9 = 512
		bytes: make([]byte, 0, 120),
	}
}

//重设Encoder
func (e *Encoder) Reset() {
	e.buf.Reset()
	e.bytes = make([]byte, 0, 120) //清空
}

//向Encoder追加一个数字
func (e *Encoder) Write(v uint64) (err error) {
	if e.buf.UnusedSize() == 0 { //已满，需要编码输出一下当前缓冲里的数组
		err = e.flush()
		if err != nil {
			return
		}
	}
	//添加到ringbuffer中
	e.buf.Write(v)
	return
}

//返回将所有的数字进行编码后的字节数组
func (e *Encoder) Bytes() ([]byte, error) {
	for e.buf.UsedSize() > 0 {
		if err := e.flush(); err != nil {
			return nil, err
		}
	}
	return e.bytes, nil
}

//flush 把数字从ringbuffer编码输出到bytes里
func (e *Encoder) flush() (err error) {
	bufferd := e.buf.ReadAll()
	var encodedList []uint64
	encodedList, err = EncodeList(bufferd)
	if err != nil {
		return
	}
	//编码后的uint64数组，存入字节数组
	for _, encoded := range encodedList {
		b := make([]byte, 8)
		//fmt.Println("编码后数值：", encoded)
		binary.BigEndian.PutUint64(b, encoded)
		e.bytes = append(e.bytes, b...)
	}
	return
}

// Decoder converts a compressed byte slice to a stream of unsigned 64bit integers.
// Decoder 解码器,把一系列编码过的字节序列解码成一系列uint64的
type Decoder struct {
	bytes    []byte
	buf      []uint64 //解码之后的数据放在这个数组
	bufIndex int
	bufLen   int
}

// NewDecoder returns a Decoder from a byte slice
func NewDecoder(b []byte) *Decoder {
	return &Decoder{
		bytes: b,
		buf:   make([]uint64, 240),
	}
}

// Next 如果接下来还有值可以读则返回true.
// 连续调用会把当前元素指针向下移动
func (d *Decoder) Next() bool {
	d.bufIndex += 1
	if d.bufIndex >= d.bufLen {
		d.decode8bytes()
	}
	var n bool
	if len(d.bytes) >= 8 || (d.bufIndex >= 0 && d.bufIndex < d.bufLen) {
		n = true
	}
	return n
}

// Read returns the current value.  Successive calls to Read return the same value.
//Read 返回解码的当前的值
func (d *Decoder) Read() uint64 {
	v := d.buf[d.bufIndex]
	return v
}

//decode8bytes 每次解码8bytes，也就是一个uint64的block。
func (d *Decoder) decode8bytes() {
	if len(d.bytes) < 8 {
		return
	}
	encoded := binary.BigEndian.Uint64(d.bytes[:8])
	d.bytes = d.bytes[8:]

	n := Decode(encoded, (*[240]uint64)(unsafe.Pointer(&d.buf[0])))
	d.bufLen = n
	d.bufIndex = 0
}

//Simple-8b方案的编码模式：
//selector value  0   1   2   3   4   5   6   7   8   9   10  11  12  13  14  15
//integers coded  240 120 60  30  20  15  12  10  8   7   6   5   4   3   2   1
//bits per integer0   0   1   2   3   4   5   6   7   8   10  12  15  20  30  60

//对src数组进行编码.尽可能多的填满到[1个]uint64里
//返回编码后的uint64,编码了多少个数字,和是否出错
func Encode(src []uint64) (encoded uint64, n int, err error) {
	for sel := 0; sel < 16; sel++ {
		if canPackN(src, sel) { //贪婪的依次进行尝试,直到某个selector可以填满一个uint64
			encoded, n = packN(src, sel)
			return
		}
	}
	if len(src) > 0 {
		encoded = 0
		n = 0
		err = fmt.Errorf("value out of boundary:%v", src)
		return
	}
	return
}

//EncodeList 把src数组的数组全部进行编码
//返回编码后的数组
//如果出错则err!=nil
func EncodeList(src []uint64) (encodedList []uint64, err error) {
	i := 0
	encodedList = make([]uint64, 0, 240)
	for {
		if i >= len(src) {
			break
		}
		remaining := src[i:]
		var encoded uint64
		var n int

		encoded, n, err = Encode(remaining)
		if err != nil {
			return
		}
		//返回的一个uint64添加到切片,然后根据返回的编码了多少个数字n,跳过n,继续处理接下来的
		encodedList = append(encodedList, encoded)
		i += n
	}
	return
}

//Decode 解码encoded到dst
//返回解码出来的数量
func Decode(encoded uint64, dst *[240]uint64) (n int) {
	sel := encoded >> 60
	selector := selectors[sel]
	if sel == 0 || sel == 1 { //特殊处理,填满240或120个1
		n = selector.n
		for i := 0; i < n; i++ {
			dst[i] = 1
		}
		return
	}
	//fmt.Printf("原始数据:%b\n", encoded)
	max := selector.max
	for i := 0; i < selector.n; i++ {
		v := encoded >> ((selector.n - 1 - i) * selector.bits) //移位后取出对应数据
		v &= max                                               //max is actually the mask,reset the high tag bits to zeros
		dst[i] = v
	}
	n = selector.n
	return
}

func DecodeList(encodedList []uint64) (decodedList []uint64) {
	dstslice := make([]uint64, 240)
	dst := (*[240]uint64)(unsafe.Pointer(&dstslice[0]))
	for _, encoded := range encodedList {
		n := Decode(encoded, dst)
		decodedList = append(decodedList, dstslice[0:n]...)
	}
	return
}

type packing struct {
	n, bits int
	max     uint64
}

var selectors [16]packing = [16]packing{
	{240, 0, 1},                       //特殊:240个1(or 0?)
	{120, 0, 1},                       //特殊:120个1(or 0?)
	{60, 1, 1},                        //60个1或0
	{30, 2, uint64(1<<uint64(2) - 1)}, //30个 0,1,2或3
	{20, 3, uint64(1<<uint64(3) - 1)},
	{15, 4, uint64(1<<uint64(4) - 1)},
	{12, 5, uint64(1<<uint64(5) - 1)},
	{10, 6, uint64(1<<uint64(6) - 1)},
	{8, 7, uint64(1<<uint64(7) - 1)},
	{7, 8, uint64(1<<uint64(8) - 1)},
	{6, 10, uint64(1<<uint64(10) - 1)},
	{5, 12, uint64(1<<uint64(12) - 1)},
	{4, 15, uint64(1<<uint64(15) - 1)},
	{3, 20, uint64(1<<uint64(20) - 1)},
	{2, 30, uint64(1<<uint64(30) - 1)},
	{1, 60, uint64(1<<uint64(60) - 1)},
}

//针对src数组,按照selector对应的n个uint64,以bits个比特的宽度编码到1个uint64里
func canPackN(src []uint64, sel int) bool {
	selector := selectors[sel]
	n, bits := selector.n, selector.bits

	end := len(src)
	if end < n { //src不够长,没有n个数字
		return false
	} else {
		end = n
	}
	//bits=0表示接下来是240或120个1
	if bits == 0 {
		for i := 0; i < end; i++ {
			v := src[i]
			if v != 1 {
				return false
			}
		}
		return true
	}
	//判断接下来的这些数字是否超过了bits能表示的最大值
	max := uint64(1<<uint64(bits) - 1)
	for i := 0; i < end; i++ {
		if src[i] > max {
			return false
		}
	}
	return true
}

//针对src数组,按照selector对应的n个uint64,以bits个比特的宽度编码到1个uint64里
//返回编码后的uint64
func packN(src []uint64, sel int) (encoded uint64, n int) {
	selector := selectors[sel]
	n, bits := selector.n, selector.bits
	var tag uint64 = uint64(sel) << 60
	if sel == 0 || sel == 1 { //特殊处理,只返回tag即可
		encoded = tag
		return
	}
	encoded = tag
	//针对接下来的n个数字进行编码. n0 n1 n2 ... 由高到低的顺序依次存储在encoded,并且是右对齐在各自的bit范围内
	for i := 0; i < n; i++ {
		v := src[i]
		encoded |= v << ((n - 1 - i) * bits)
	}
	return
}

//返回解码 encoded 需要的数量大小
func unpackNeedN(encoded uint64) (n int, err error) {
	sel := encoded >> 60
	//we can check below two special cases only
	if sel == 0 && encoded != 0 {
		err = fmt.Errorf("selector 0 error %b", encoded)
		return
	}
	if sel == 1 && encoded != 0x1000000000000000 {
		err = fmt.Errorf("selector 1 error %b", encoded)
		return
	}
	n = selectors[sel].n
	return
}
