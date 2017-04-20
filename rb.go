package ringbuffer

import (
	"errors"
	"fmt"
	"io"
)

var ErrOverFlow = errors.New("out of data stream range")

//环形buffer
type RingBuffer struct {
	data        []byte
	beginCursor int64 //开始指针: 数据流的开始位置
	endCursor   int64 //结束指: 数控流的结束位置
	index       int   //数据索引,范围从 0 到 len(rb.data)-1
}

//size 必须大于0
func NewRingBuffer(size int64, begin int64) (rb RingBuffer) {
	//预分配data
	rb.data = make([]byte, size)
	rb.index = 0
	rb.beginCursor = begin
	rb.endCursor = begin
	return
}

func (rb *RingBuffer) Write(buf []byte) (n int, err error) {

	if len(buf) > len(rb.data) {
		err = ErrOverFlow
		return
	}

	//buff数据入队
	for n < len(buf) {
		writtenCount := copy(rb.data[rb.index:], buf[n:]) //FIFO
		rb.endCursor += int64(writtenCount)
		n += writtenCount
		rb.index += writtenCount

		if rb.index >= len(rb.data) { //索引重塑
			rb.index -= len(rb.data)
		}
	}

	//入队结束后, 要移动相关的游标
	if int(rb.endCursor-rb.beginCursor) > len(rb.data) {
		rb.beginCursor = rb.endCursor - int64(len(rb.data))
	}

	return
}

func (rb *RingBuffer) WriteAt(buf []byte, offset int64) (n int, err error) {

	//越界
	if offset+int64(len(buf)) > rb.endCursor || offset < rb.beginCursor {
		err = ErrOverFlow
		return
	}

	var writeOffset int //从哪里开始写

	if rb.endCursor-rb.beginCursor < int64(len(rb.data)) {
		writeOffset = int(offset - rb.beginCursor)
	} else {
		writeOffset = rb.index + int(offset-rb.beginCursor)
	}

	if writeOffset > len(rb.data) {
		writeOffset -= len(rb.data)
	}

	//就是写到哪里位置
	writeEndCursor := writeOffset + int(rb.endCursor-offset)
	if writeEndCursor <= len(rb.data) {
		n = copy(rb.data[writeOffset:writeEndCursor], buf)
	} else {
		n = copy(rb.data[writeOffset:], buf)
		if n < len(buf) { //数据没进全的处理
			n += copy(rb.data[:writeEndCursor-len(rb.data)], buf[n:])
		}
	}
	return
}

func (rb *RingBuffer) ReadAt(buf []byte, offset int64) (n int, err error) {

	if offset > rb.endCursor || offset < rb.beginCursor {
		err = ErrOverFlow
		return
	}

	var readOffset int //从哪里开始读

	if rb.endCursor-rb.beginCursor < int64(len(rb.data)) {
		readOffset = int(offset - rb.beginCursor)
	} else {
		readOffset = rb.index + int(offset-rb.beginCursor)
	}

	if readOffset >= len(rb.data) {
		readOffset -= len(rb.data)
	}

	readEndCursor := readOffset + int(rb.endCursor-offset)

	if readEndCursor <= len(rb.data) {
		n = copy(buf, rb.data[readOffset:readEndCursor])
	} else {
		n = copy(buf, rb.data[readOffset:])
		if n < len(buf) {
			n += copy(buf[n:], rb.data[:readEndCursor-len(rb.data)])
		}
	}
	if n < len(buf) {
		err = io.EOF
	}

	return
}

//容量调整
func (rb *RingBuffer) Resize(newSize int) {
	if newSize == len(rb.data) {
		return
	}
	newData := make([]byte, newSize)
	var offset int

	//旧队列是满的情况
	if rb.endCursor-rb.beginCursor == int64(len(rb.data)) {
		offset = rb.index
	}
	if int(rb.endCursor-rb.beginCursor) > newSize {
		discard := int(rb.endCursor-rb.beginCursor) - newSize
		offset = (offset + discard) % len(rb.data)
		rb.beginCursor = rb.endCursor - int64(newSize)
	}
	//首次进队
	n := copy(newData, rb.data[offset:])
	if n < newSize { //数据没进全的处理
		copy(newData[n:], rb.data[:offset])
	}
	//reflesh
	rb.data = newData
	rb.index = 0
}

func (rb *RingBuffer) Skip(length int64) {
	rb.endCursor += length
	rb.index += int(length)
	for rb.index >= len(rb.data) {
		rb.index -= len(rb.data)
	}
	if int(rb.endCursor-rb.beginCursor) > len(rb.data) {
		rb.beginCursor = rb.endCursor - int64(len(rb.data))
	}
}

func (rb *RingBuffer) Dump() []byte {
	dump := make([]byte, len(rb.data))
	copy(dump, rb.data)
	return dump
}

func (rb *RingBuffer) String() string {
	return fmt.Sprintf("[size:%v, start:%v, end:%v, index:%v]", len(rb.data), rb.beginCursor, rb.endCursor, rb.index)
}

func (rb *RingBuffer) Size() int64 {
	return int64(len(rb.data))
}

func (rb *RingBuffer) Begin() int64 {
	return rb.beginCursor
}

func (rb *RingBuffer) End() int64 {
	return rb.endCursor
}

func (rb *RingBuffer) Evacuate(offset int64, length int) (newOff int64) {
	if offset+int64(length) > rb.endCursor || offset < rb.beginCursor {
		return -1
	}
	var readOff int
	if rb.endCursor-rb.beginCursor < int64(len(rb.data)) {
		readOff = int(offset - rb.beginCursor)
	} else {
		readOff = rb.index + int(offset-rb.beginCursor)
	}
	if readOff >= len(rb.data) {
		readOff -= len(rb.data)
	}

	if readOff == rb.index {
		// no copy evacuate
		rb.index += length
		if rb.index >= len(rb.data) {
			rb.index -= len(rb.data)
		}
	} else if readOff < rb.index {
		var n = copy(rb.data[rb.index:], rb.data[readOff:readOff+length])
		rb.index += n
		if rb.index == len(rb.data) {
			rb.index = copy(rb.data, rb.data[readOff+n:readOff+length])
		}
	} else {
		var readEnd = readOff + length
		var n int
		if readEnd <= len(rb.data) {
			n = copy(rb.data[rb.index:], rb.data[readOff:readEnd])
			rb.index += n
			if rb.index == len(rb.data) {
				rb.index = copy(rb.data, rb.data[readOff+n:readEnd])
			}
		} else {
			n = copy(rb.data[rb.index:], rb.data[readOff:])
			rb.index += n
			var tail = length - n
			n = copy(rb.data[rb.index:], rb.data[:tail])
			rb.index += n
			if rb.index == len(rb.data) {
				rb.index = copy(rb.data, rb.data[n:tail])
			}
		}
	}
	newOff = rb.endCursor
	rb.endCursor += int64(length)
	if rb.beginCursor < rb.endCursor-int64(len(rb.data)) {
		rb.beginCursor = rb.endCursor - int64(len(rb.data))
	}
	return
}
