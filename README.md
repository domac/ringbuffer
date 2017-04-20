# ringbuffer

基于Go实现的Ring Buffer

使用方式:

```go

func main() {
    rb := NewRingBuffer(32, 0)
    rb.Write([]byte("ABCDEFGH"))
    println(string(rb.Dump()))
}

```
