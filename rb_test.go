package ringbuf

import (
	"context"
	"testing"
	"time"
)

// func TestRingBufOneByOne(t *testing.T) {

// 	cap := uint32(100)

// 	rb := NewRingBuf(WithCap(cap))

// 	for i := uint32(0); i < cap; i++ {
// 		err := rb.Enqueue(i)

// 		if err != nil {
// 			panic("enqueue " + err.Error())
// 		}
// 	}

// 	for i := uint32(0); i < cap; i++ {
// 		item, err := rb.Dequeue()

// 		if err != nil {
// 			panic("dequeue " + err.Error())
// 		}

// 		if item != i {
// 			panic(fmt.Sprintf("dequeue item value is  %v , expect value is %d", item, i))
// 		}
// 	}

// 	t.Log("done")

// }

// func TestRingBufMutil(t *testing.T) {

// 	wgWrite := &sync.WaitGroup{}
// 	wgRead := &sync.WaitGroup{}
// 	writeDone := uint32(1)

// 	gNum := runtime.NumCPU()
// 	// gNum := 1
// 	preGWriteEleNum := 100
// 	runtime.GOMAXPROCS(gNum)

// 	cap := uint32(100)

// 	rb := NewRingBuf(WithCap(cap))

// 	for i := 0; i < gNum; i++ {
// 		wgWrite.Add(1)

// 		go func(idx int) {

// 			writeNum := idx*10 + preGWriteEleNum

// 			for j := idx * 10; j < writeNum; j++ {
// 				err := rb.Enqueue(j)

// 				if err != nil && err != ErrQueueFull {
// 					panic(fmt.Sprintf("enqueue g idx is %d , err is %s", idx, err.Error()))
// 				}

// 				if err == ErrQueueFull {
// 					time.Sleep(1 * time.Millisecond)
// 					continue
// 				}

// 				t.Logf("enqueue g idx is %d ,  item is %d", idx, j)

// 			}

// 			t.Logf("enqueue g idx is %d , enqueue done \n", idx)

// 			wgWrite.Done()

// 		}(i)

// 	}

// 	for i := 0; i < 3; i++ {

// 		wgRead.Add(1)

// 		go func(idx int) {

// 			for {

// 				item, err := rb.Dequeue()

// 				if err != nil && err != ErrQueueEmpty {
// 					panic(fmt.Sprintf("dequeue g idx is %d , err is %s", idx, err.Error()))
// 				}
// 				if err == ErrQueueEmpty {

// 					if atomic.LoadUint32(&writeDone) == 2 {
// 						t.Logf("g idx is %d ,dequeue done \n", idx)
// 						wgRead.Done()
// 						return
// 					}

// 					time.Sleep(1 * time.Millisecond)
// 					continue
// 				}
// 				if item == nil {
// 					panic(fmt.Sprintf("dequeue item is nil g idx is %d", idx))
// 				}
// 				t.Logf("dequeue g idx is %d , item is %v \n", idx, item)
// 			}

// 		}(i)

// 	}

// 	wgWrite.Wait()
// 	atomic.StoreUint32(&writeDone, 2)
// 	t.Logf("write Done \n")

// 	wgRead.Wait()

// 	t.Logf("read Done \n")
// }

func BenchmarkRingBuf_Put16384(b *testing.B) {

	b.ResetTimer()

	ctxWrite, writeCancel := context.WithCancel(context.Background())

	ctxRead, readCancel := context.WithCancel(context.Background())

	cap := uint32(100)
	rb := NewRingBuf(WithCap(cap))

	go enqueueRoutine(b, rb, writeCancel, b.N)

	go dequeueRoutine(b, rb, readCancel)

	<-ctxWrite.Done()

	<-ctxRead.Done()

}

func enqueueRoutine(t *testing.B, rb RingBuf, cancel context.CancelFunc, maxN int) {

	var err error
	for i := 0; i < maxN; i++ {

		err = rb.Enqueue(i)
		if err != nil {
			if err == ErrQueueFull {
				// block till queue not full
				time.Sleep(1 * time.Millisecond)
				continue
			}

			t.Fatalf("[Enqueue] failed on i=%v. err: %s.", i, err.Error())
		}

	}

	cancel()
	t.Log("[Enqueue] END")
}

func dequeueRoutine(t *testing.B, rb RingBuf, cancel context.CancelFunc) {

	for {

		_, err := rb.Dequeue()
		if err != nil {
			if err == ErrQueueEmpty {
				// block till queue not full
				break
			}

			t.Fatalf("[Dequeue] failed err: %s.", err.Error())
		}

	}

	cancel()
	t.Log("[Dequeue] END")
}
