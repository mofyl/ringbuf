package ringbuf

type RingBuf interface {
	Enqueue(item any) error
	Dequeue() (any, error)
	Len() uint32
	Cap() uint32
	Reset()
}
