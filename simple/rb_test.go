package simple

import (
	"bytes"
	"testing"
)

func TestRingBufferShortWrite(t *testing.T) {
	buf, err := NewBuffer(1024)
	if err != nil {
		t.Fatal(err.Error())
	}

	demo := []byte("hello world")

	n, err := buf.Write(demo)
	if err != nil {
		t.Fatal(err.Error())
	}

	if n != len(demo) {
		t.Fatalf("error len %v", n)
	}

	if !bytes.Equal(buf.ReadAll(), demo) {
		t.Fatal("not equal")
	}
}

func TestRingBufferFullWrite(t *testing.T) {
	demo := []byte("hello world")
	buf, err := NewBuffer(int64(len(demo)))
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err := buf.Write(demo)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if n != len(demo) {
		t.Fatalf("error: %v", n)
	}

	if !bytes.Equal(buf.ReadAll(), demo) {
		t.Fatalf("error: %v", buf.ReadAll())
	}
}

func TestBufferLongWrite(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(6)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("error: %v", n)
	}

	expect := []byte(" world")
	if !bytes.Equal(buf.ReadAll(), expect) {
		t.Fatalf("error: %s", buf.ReadAll())
	}
}

func TestBufferHugeWrite(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(3)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("error: %v", n)
	}

	expect := []byte("rld")
	if !bytes.Equal(buf.ReadAll(), expect) {
		t.Fatalf("error: %s", buf.ReadAll())
	}
}

func TestBufferManySmall(t *testing.T) {
	inp := []byte("hello world")

	buf, err := NewBuffer(3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, b := range inp {
		n, err := buf.Write([]byte{b})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if n != 1 {
			t.Fatalf("error: %v", n)
		}
	}

	expect := []byte("rld")
	if !bytes.Equal(buf.ReadAll(), expect) {
		t.Fatalf("error: %v", buf.ReadAll())
	}
}

func TestBufferMultiPart(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}
	total := 0

	buf, err := NewBuffer(16)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, b := range inputs {
		total += len(b)
		n, err := buf.Write(b)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if n != len(b) {
			t.Fatalf("error: %v", n)
		}
	}

	if int64(total) != buf.TotalWrittenCount() {
		t.Fatalf("bad total")
	}

	expect := []byte("t\nmy cool input\n")
	if !bytes.Equal(buf.ReadAll(), expect) {
		t.Fatalf("error: %v", buf.ReadAll())
	}
}

func TestBufferReset(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}

	buf, err := NewBuffer(4)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for _, b := range inputs {
		n, err := buf.Write(b)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if n != len(b) {
			t.Fatalf("error: %v", n)
		}
	}

	buf.Reset()

	input := []byte("hello")
	n, err := buf.Write(input)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if n != len(input) {
		t.Fatalf("error: %v", n)
	}

	expect := []byte("ello")
	if !bytes.Equal(buf.ReadAll(), expect) {
		t.Fatalf("error: %v", string(buf.ReadAll()))
	}
}
