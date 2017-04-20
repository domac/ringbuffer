package ringbuffer

import (
	"bytes"
	"testing"
)

func TestRingBufferWrite(t *testing.T) {
	rb := NewRingBuffer(16, 0)
	rb.Write([]byte("fghibbbbccccddde"))
	rb.Write([]byte("abcdefghi"))
	if !bytes.Equal([]byte("abcdefghicccddde"), rb.Dump()) {
		t.Fatal("write error")
	}
	rb.Write([]byte("123456789"))
	if !bytes.Equal([]byte("89cdefghi1234567"), rb.Dump()) {
		t.Fatal("write error")
	}
}

func TestRingBufferResize(t *testing.T) {
	rb := NewRingBuffer(16, 0)
	rb.Write([]byte("fghibbbbccccddde"))
	rb.Write([]byte("abcdefghiABCDEFGHIJKLMN")) //超过16位
	if !bytes.Equal([]byte("fghibbbbccccddde"), rb.Dump()) {
		t.Fatal("resize error")
	}
	rb.Resize(64)
	rb.Write([]byte("JabcdefghiABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	t.Log(string(rb.Dump()))
}

func TestRingBufferWriteAt(t *testing.T) {
	rb := NewRingBuffer(16, 0)
	rb.Write([]byte("fghibbbbccccddde"))
	rb.WriteAt([]byte("ABCD"), 3)
	if !bytes.Equal([]byte("fghABCDbccccddde"), rb.Dump()) {
		t.Fatal("write at error")
	}
}

func TestRingBufferReadAt(t *testing.T) {
	rb := NewRingBuffer(16, 0)
	rb.Write([]byte("fghibbbbccccddde"))
	data := make([]byte, 5)
	rb.ReadAt(data, 3)

	if string(data) != "ibbbb" {
		t.Fatal("read at error")
	}
}

func TestRingBufferAll(t *testing.T) {
	rb := NewRingBuffer(16, 0)
	rb.Write([]byte("fghibbbbccccddde"))
	rb.Write([]byte("fghibbbbc"))
	rb.Resize(16)
	off := rb.Evacuate(9, 3)
	t.Log(string(rb.Dump()))
	if off != rb.End()-3 {
		t.Log(string(rb.Dump()), rb.End())
		t.Fatalf("off got %v", off)
	}
	off = rb.Evacuate(15, 5)
	t.Log(string(rb.Dump()))
	if off != rb.End()-5 {
		t.Fatalf("off got %v", off)
	}
	rb.Resize(64)
	rb.Resize(32)
	data := make([]byte, 5)
	rb.ReadAt(data, off)
	if string(data) != "efghi" {
		t.Fatalf("read at should be efghi, got %v", string(data))
	}

	off = rb.Evacuate(0, 10)
	if off != -1 {
		t.Fatal("evacutate out of range offset should return error")
	}
}
