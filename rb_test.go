package ringbuf

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

//
//func TestRingBufOneByOne(t *testing.T) {
//
//	rbCap := uint32(100)
//
//	rb := NewRingBuf(WithCap(rbCap))
//
//	for i := uint32(0); i < rbCap; i++ {
//		err := rb.Enqueue(i)
//
//		if err != nil {
//			panic("enqueue " + err.Error())
//		}
//	}
//
//	for i := uint32(0); i < rbCap; i++ {
//		item, err := rb.Dequeue()
//
//		if err != nil {
//			panic("dequeue " + err.Error())
//		}
//
//		if item != i {
//			panic(fmt.Sprintf("dequeue item value is  %v , expect value is %d", item, i))
//		}
//	}
//
//	t.Log("done")
//
//}
//
//func TestRingBufMutil(t *testing.T) {
//
//	wgWrite := &sync.WaitGroup{}
//	wgRead := &sync.WaitGroup{}
//	writeDone := uint32(1)
//
//	gNum := runtime.NumCPU()
//	// gNum := 1
//	preGWriteEleNum := 100
//	runtime.GOMAXPROCS(gNum)
//
//	rbCap := uint32(100)
//
//	rb := NewRingBuf(WithCap(rbCap))
//
//	for i := 0; i < gNum; i++ {
//		wgWrite.Add(1)
//
//		go func(idx int) {
//
//			writeNum := idx*10 + preGWriteEleNum
//
//			for j := idx * 10; j < writeNum; j++ {
//				err := rb.Enqueue(j)
//
//				if err != nil && err != ErrQueueFull {
//					panic(fmt.Sprintf("enqueue g idx is %d , err is %s", idx, err.Error()))
//				}
//
//				if err == ErrQueueFull {
//					time.Sleep(1 * time.Millisecond)
//					continue
//				}
//
//				t.Logf("enqueue g idx is %d ,  item is %d", idx, j)
//
//			}
//
//			t.Logf("enqueue g idx is %d , enqueue done \n", idx)
//
//			wgWrite.Done()
//
//		}(i)
//
//	}
//
//	for i := 0; i < 3; i++ {
//
//		wgRead.Add(1)
//
//		go func(idx int) {
//
//			for {
//
//				item, err := rb.Dequeue()
//
//				if err != nil && err != ErrQueueEmpty {
//					panic(fmt.Sprintf("dequeue g idx is %d , err is %s", idx, err.Error()))
//				}
//				if err == ErrQueueEmpty {
//
//					if atomic.LoadUint32(&writeDone) == 2 {
//						t.Logf("g idx is %d ,dequeue done \n", idx)
//						wgRead.Done()
//						return
//					}
//
//					time.Sleep(1 * time.Millisecond)
//					continue
//				}
//				if item == nil {
//					panic(fmt.Sprintf("dequeue item is nil g idx is %d", idx))
//				}
//				t.Logf("dequeue g idx is %d , item is %v \n", idx, item)
//			}
//
//		}(i)
//
//	}
//
//	wgWrite.Wait()
//	atomic.StoreUint32(&writeDone, 2)
//	t.Logf("write Done \n")
//
//	wgRead.Wait()
//
//	t.Logf("read Done \n")
//}
//
//func TestRingBuf_Len(t *testing.T) {
//
//	rbCap := uint32(10)
//	rb := NewRingBuf(WithCap(rbCap))
//
//	for i := 0; i < 10; i++ {
//		rb.Enqueue(i)
//	}
//
//	if rb.Len() != 10 {
//		panic(fmt.Sprintf("fir len was wrong , cur len is %d", rb.Len()))
//	}
//
//	for i := 0; i < 5; i++ {
//		rb.Dequeue()
//	}
//
//	for i := 0; i < 3; i++ {
//		rb.Enqueue(i)
//	}
//
//	if rb.Len() != 8 {
//		panic(fmt.Sprintf("sec len was wrong , cur len is %d", rb.Len()))
//	}
//
//	t.Log("done.")
//
//}
//  go test -run "rb_test.go" -bench "BenchmarkRingBuf_Put16384" -cpuprofile cpu.profile -memprofile mem.profile -benchmem -v
func BenchmarkRingBuf_Put16384(b *testing.B) {
	b.ResetTimer()

	ctxWrite, writeCancel := context.WithCancel(context.Background())

	ctxRead, readCancel := context.WithCancel(context.Background())

	gNum := runtime.NumCPU()
	runtime.GOMAXPROCS(gNum)

	rb := NewRingBufList()

	enqueueMultiGorutineRingBuf(b, rb, writeCancel, gNum, b.N)

	dequeueMultiGorutineRingBuf(b, rb, ctxWrite, readCancel, gNum)

	<-ctxWrite.Done()
	<-ctxRead.Done()
}

func enqueueMultiGorutineRingBuf(b *testing.B,
	rb RingBuf,
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

func dequeueMultiGorutineRingBuf(b *testing.B,
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
