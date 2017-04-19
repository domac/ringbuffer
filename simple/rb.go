package simple

import (
	"errors"
)

type RingBuffer struct {
	data         []byte
	size         int64
	writeCursor  int64
	writtenCount int64
}

//size 必须大于0
func NewBuffer(size int64) (*RingBuffer, error) {
	if size <= 0 {
		return nil, errors.New("Size must be positive")
	}

	b := &RingBuffer{
		size: size,
		data: make([]byte, size)}

	return b, nil
}

//读出buffer所有数据
func (b *RingBuffer) ReadAll() []byte {

	switch {
	case b.writtenCount >= b.size && b.writeCursor == 0:
		return b.data
	case b.writtenCount > b.size:
		out := make([]byte, b.size)
		copy(out, b.data[b.writeCursor:])
		copy(out[b.size-b.writeCursor:], b.data[:b.writeCursor])
		return out
	default:
		return b.data[:b.writeCursor]
	}

	return nil
}

//写入buf到ringbuffer内部
//如果需要会覆盖旧数据(fifo)
func (b *RingBuffer) Write(buf []byte) (int, error) {

	n := len(buf)
	b.writtenCount += int64(n)

	//如果buf的大小超过容量限制,根据fifo原则
	//我们只关注最近最新的部分数据
	if int64(n) > b.size {
		buf = buf[int64(n)-b.size:]
	}

	remain := b.size - b.writeCursor
	copy(b.data[b.writeCursor:], buf)
	if int64(len(buf)) > remain {
		copy(b.data, buf[remain:])
	}

	b.writeCursor = ((b.writeCursor + int64(len(buf))) % b.size)
	return n, nil
}

func (b *RingBuffer) Size() int64 {
	return b.size
}

func (b *RingBuffer) TotalWrittenCount() int64 {
	return b.writtenCount
}

func (b *RingBuffer) Reset() {
	b.writeCursor = 0
	b.writtenCount = 0
}

func (b *RingBuffer) String() string {
	return string(b.ReadAll())
}
