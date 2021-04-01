package main

/*
head指向的位置，是下一次可以写的位置
	tail指向的下一位置，是可以读的位置
	[tail+1,head-1]区间是可以读的位置
	满的时候，head==tail
	空的时候，head-1==tail
缓存里永远有一个位置是不放数据的，就是tail所在的位置

1.head在前，空的时候：head-1==tail
	0 1 2 3 4 5 6 7
	x x	x x x x x x
	    |->
	 	t h

2.head在前，有数据的时候： [tail+1,head-1]区间是可以读的数据
	0 1 2 3 4 5 6 7
	x x	x o o x x x
	    |----->
	    t     h

3.head不在前，有数据的时候： [tail+1,结尾]和[开始，head-1]区间是可以读的数据
	0 1 2 3 4 5 6 7
	o x	x o o o o o
	--> |----------
	  h t

4.head不在前，满的时候： head==tail，[tail+1,结尾]和[开始，head-1]区间是可以读的数据
	0 1 2 3 4 5 6 7
	o o	x o o o o o
	    |>
	    th
*/
type RingBuffer struct {
	buf    []uint64 //缓存将要被编码的数字序列
	bufLen int      //缓冲区长度大小
	head   int      //头的位置
	tail   int      //尾的位置
}

//NewRingBuffer create a ringbuffer with defaultSzie
func NewRingBuffer(defaultSize int) *RingBuffer {
	return &RingBuffer{
		bufLen: defaultSize,
		buf:    make([]uint64, defaultSize),
		head:   1,
		tail:   0,
	}
}

//Read read only one data.if there is one data in the buf,n is 1, or n is 0.
//Read 读取一个数据,如果获取到则n=1,否则n=0
func (e *RingBuffer) Read() (data uint64, n int) {
	data, n = e.Peek()
	//advance the tail
	if n > 0 {
		e.tail = (e.tail + 1) % e.bufLen
	}
	return
}

//ReadAll read out all datas
//ReadAll 读取所有的数据
func (e *RingBuffer) ReadAll() (data []uint64) {
	head, tail := e.head, e.tail
	e.head = 1
	e.tail = 0

	if head > tail { //head is ahead of tail
		if head > tail-1 { //case 2
			return e.buf[tail+1 : head]
		} else { // case 1
			return e.buf[0:0] //空
		}
	} else { //head isn't ahead of tail( head <= tail)
		var part1 []uint64
		var part2 []uint64
		//[tail+1:] has data
		if tail < e.bufLen-1 {
			part1 = e.buf[tail+1:]
		}
		//[:head-1] has data
		if head > 0 {
			part2 = e.buf[:head]
		}
		//add two parts together
		part1 = append(part1, part2...)
		return part1
	}
}

//Write append a data into the ringbuffer.If ringbuffer is full,return 0,or return 1.
//Write 向RingBuffer追加一个数字
func (e *RingBuffer) Write(v uint64) (n int) {
	if e.UnusedSize() == 0 { //full
		n = 1
		return
	} else {
		//append into ringbuffer
		e.buf[e.head] = v
		//advance the head
		e.head = (e.head + 1) % e.bufLen
		return
	}
}

//UnusedSize returns the unused number of current buffer
//UnusedSize 返回当前ringbuffer可用的空闲数量
func (e *RingBuffer) UnusedSize() int {
	n := (e.bufLen + e.tail - e.head) % e.bufLen
	return n
}

//UsedSize returns the number of data that have been written into the current buffer.
//UnusedSize 返回当前ringbuffer存放数据的数量
func (e *RingBuffer) UsedSize() int {
	n := (e.bufLen + e.head - e.tail - 1) % e.bufLen
	return n
}

//Peek read only one data.if there is one data in the buf,n is 1, or n is 0. It doesn't change the head or tail index
//Peek 获取一个数据,不改变其他,如果获取到则n=1,否则n=0
func (e *RingBuffer) Peek() (data uint64, n int) {
	size := e.UsedSize()
	if size > 0 {
		idx := (e.tail + 1) % e.bufLen
		data = e.buf[idx]
		n = 1
	} else {
		n = 0
	}
	return
}

//Reset reset the ringbuffer
//Reset 重置ringbuffer
func (e *RingBuffer) Reset() {
	e.head = 1
	e.tail = 0
	e.buf = e.buf[0:0]
}
