package ringbuf

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestRingBufList_OneByOne(t *testing.T) {

	rb := NewRingBufList()

	eleN := 500

	for i := 0; i < eleN; i++ {
		rb.Enqueue(i)
	}

	for i := 0; i < eleN; i++ {

		e, err := rb.Dequeue()

		if err != nil {
			panic(fmt.Sprintf("dequeue err %s", err.Error()))
		}

		if e != i {
			panic(fmt.Sprintf("dequeue err , expect value is %d , get value is %v", i, e))
		}

	}

	t.Log("done")

}

func TestRingBufList_EnDe(t *testing.T) {
	writeWg := &sync.WaitGroup{}
	readWg := &sync.WaitGroup{}

	writeCtx, cancel := context.WithCancel(context.Background())

	gNum := runtime.NumCPU()
	runtime.GOMAXPROCS(gNum)

	preGWriteEleNum := 100

	rb := NewRingBufList()

	for i := 0; i < gNum; i++ {
		writeWg.Add(1)

		go func(idx int) {

			writeNum := idx + preGWriteEleNum

			for j := idx; j < writeNum; j++ {

				err := rb.Enqueue(j)

				if err != nil {
					panic(fmt.Sprintf("Enqueue : err is %s", err.Error()))
				}

			}

			t.Logf("Enqueue: g idx is %d , enqueue done \n", idx)

			writeWg.Done()

		}(i)

	}

	for i := 0; i < gNum; i++ {

		readWg.Add(1)

		go func(idx int) {

			for {

				item, err := rb.Dequeue()

				if err != nil && err != ErrQueueEmpty {
					panic(fmt.Sprintf("Dequeue: g idx is %d , err is %s", idx, err.Error()))
				}

				if err == ErrQueueEmpty {

					select {
					case _, ok := <-writeCtx.Done():
						if !ok {
							t.Logf("Dequeue: g idx is %d ,dequeue done \n", idx)
							readWg.Done()
							return
						}
					default:
						time.Sleep(1 * time.Millisecond)
						continue
					}

				}

				if item == nil {
					panic(fmt.Sprintf("Dequeue: item is nil g idx is %d", idx))
				}

				t.Logf("Dequeue: g idx is %d , item is %v \n", idx, item)

			}

		}(i)

	}

	writeWg.Wait()
	cancel()
	t.Logf("write done")

	readWg.Wait()
	t.Logf("read done")
}

func BenchmarkEnDe(b *testing.B) {

	b.ResetTimer()

	ctxWrite, writeCancel := context.WithCancel(context.Background())

	ctxRead, readCancel := context.WithCancel(context.Background())

	gNum := runtime.NumCPU()
	runtime.GOMAXPROCS(gNum)

	rb := NewRingBufList()

	enqueueMultiGorutine(b, rb, ctxWrite, writeCancel, gNum, b.N)

	dequeueMultiGorutine(b, rb, ctxWrite, readCancel, gNum)

	<-ctxWrite.Done()
	<-ctxRead.Done()

}

func enqueueMultiGorutine(b *testing.B,
	rb RingBuf,
	writeCtx context.Context,
	cancel context.CancelFunc,
	gNum int,
	maxN int) {

	wg := &sync.WaitGroup{}

	for i := 0; i < gNum; i++ {
		wg.Add(1)

		go func(idx int) {

			writeNum := idx + maxN

			for j := idx; j < writeNum; j++ {

				err := rb.Enqueue(j)

				if err != nil {
					panic(fmt.Sprintf("Enqueue : err is %s", err.Error()))
				}

			}

			// b.Logf("Enqueue: g idx is %d , enqueue done \n", idx)

			wg.Done()

		}(i)

	}

	wg.Wait()
	cancel()
	b.Logf("write done")

}

func dequeueMultiGorutine(b *testing.B,
	rb RingBuf,
	writeCtx context.Context,
	cancel context.CancelFunc,
	gNum int) {

	wg := &sync.WaitGroup{}

	for i := 0; i < gNum; i++ {

		wg.Add(1)

		go func(idx int) {

			for {

				item, err := rb.Dequeue()

				if err != nil && err != ErrQueueEmpty {
					panic(fmt.Sprintf("Dequeue: g idx is %d , err is %s", idx, err.Error()))
				}

				if err == ErrQueueEmpty {

					select {
					case _, ok := <-writeCtx.Done():
						if !ok {
							b.Logf("Dequeue: g idx is %d ,dequeue done \n", idx)
							wg.Done()
							return
						}
					default:
						time.Sleep(1 * time.Millisecond)
						continue
					}

				}

				if item == nil {
					panic(fmt.Sprintf("Dequeue: item is nil g idx is %d", idx))
				}

				// b.Logf("Dequeue: g idx is %d , item is %v \n", idx, item)

			}

		}(i)

	}

	wg.Wait()
	cancel()
	b.Logf("read done")
}
