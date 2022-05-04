package ringbuf

import (
	"runtime"
	"sync/atomic"

	"golang.org/x/sys/cpu"
)

type Opt func(rb *ringBuf)

/*
	这里 会将  capacity 向上 取最接近 的 2^n 作为 最终的 cap
	eg :
		capacity = 3
		finCap = 4 // 因为 距离3 最接近的 2^n 为 2^2=4

		capacity = 4
		finCap = 4 // 因为 距离3 最接近的 2^n 为 2^2=4
*/

type ringBuf struct {
	_ cpu.CacheLinePad

	data       []*rbItem
	cap        uint32
	capModMask uint32
	head       uint32
	tail       uint32
}

func WithCap(capacity uint32) Opt {

	return func(rb *ringBuf) {
		// 表示 capacity 当前就是一个 2^n
		finCap := roundUpToPower2(capacity)
		rb.cap = finCap
		rb.capModMask = rb.cap - 1
	}
}

func NewRingBuf(opt ...Opt) RingBuf {

	rb := &ringBuf{}

	for i := 0; i < len(opt); i++ {
		opt[i](rb)
	}

	rb.data = make([]*rbItem, rb.cap)

	for i := 0; i < int(rb.cap); i++ {

		rb.data[i] = &rbItem{}
		rb.data[i].Init()

	}

	return rb

}

func (rb *ringBuf) Enqueue(item any) error {

	for {

		head := atomic.LoadUint32(&rb.head)
		tail := atomic.LoadUint32(&rb.tail)
		newTail := (tail + 1) & rb.capModMask

		if newTail == head {
			return ErrQueueFull
		}

		if head == tail && head == MaxUint32 {
			return ErrQueueNotReady
		}

		ele := rb.data[tail]

		atomic.CompareAndSwapUint32(&rb.tail, tail, newTail)
		// 将 状态 从 初始状态 变为 已经占用
	retry:

		err := ele.enqueue(item)

		if err == ErrItemNeedReTry {
			goto retry
		}

		if err == ErrItemGosched {
			runtime.Gosched()
			continue
		}

		return err
	}

}

func (rb *ringBuf) Dequeue() (any, error) {

	for {

		tail := atomic.LoadUint32(&rb.tail)
		head := atomic.LoadUint32(&rb.head)

		if head == tail {

			if head == MaxUint32 {
				return nil, ErrQueueNotReady
			}
			return nil, ErrQueueEmpty
		}

		ele := rb.data[head]

		newHead := (head + 1) & rb.capModMask

		atomic.CompareAndSwapUint32(&rb.head, head, newHead)

	retry:
		item, err := ele.dequeue()

		if err == ErrItemNeedReTry {
			goto retry
		}

		if err == ErrItemGosched {
			runtime.Gosched()
			continue
		}
		return item, err
	}

}

func (rb *ringBuf) Reset() {

	atomic.StoreUint32(&rb.head, MaxUint32)
	atomic.StoreUint32(&rb.tail, MaxUint32)

	for i := 0; i < len(rb.data); i++ {
		rb.data[i].Reset()
	}

	atomic.StoreUint32(&rb.head, 0)
	atomic.StoreUint32(&rb.tail, 0)

}

func (rb *ringBuf) Len() uint32 {

	head := atomic.LoadUint32(&rb.head)
	tail := atomic.LoadUint32(&rb.tail)

	if tail >= head {
		return tail - head
	}

	// head > tail 表示

	return rb.cap - head + tail
}

func (rb *ringBuf) Cap() uint32 {
	return rb.cap
}

func roundUpToPower2(v uint32) uint32 {

	v--

	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16

	v++

	return v
}
