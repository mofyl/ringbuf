package ringbuf

import (
	"sync/atomic"
	"unsafe"
)

// =======================  rb list

type ringBufList struct {
	tail   unsafe.Pointer
	head   unsafe.Pointer
	length int32
}

func NewRingBufList() RingBuf {

	head := &rbItemList{}
	rb := &ringBufList{
		tail: unsafe.Pointer(head),
		head: unsafe.Pointer(head),
	}

	return rb

}

func (rb *ringBufList) Enqueue(item any) error {

	n := &rbItemList{value: item}

	for {
		tail := atomicLoadToRbItem(&rb.tail)
		next := atomicLoadToRbItem(&tail.next)

		if tail == atomicLoadToRbItem(&rb.tail) {
			// 表示位置正确 可以cas
			if next == nil {
				if rbItemAtomicCas(&tail.next, next, n) {
					rbItemAtomicCas(&rb.tail, tail, n)
					atomic.AddInt32(&rb.length, 1)

					return nil
				}
			} else {
				// 这里可能 是 tail 的位置不对，也就是说 tail 发生的变化，我们需要向下移动 移动到 next 的位置
				rbItemAtomicCas(&rb.tail, tail, next)
			}

		}
	}

}

func (rb *ringBufList) Dequeue() (any, error) {

	for {

		head := atomicLoadToRbItem(&rb.head)
		tail := atomicLoadToRbItem(&rb.tail)
		next := atomicLoadToRbItem(&head.next)

		if head == atomicLoadToRbItem(&rb.head) {

			if head == tail {

				if next == nil {
					// 这里表示是 空list
					return nil, ErrQueueEmpty
				}
				// 这里表示 可能 tail 发生了变化， 我们这里更新tail
				rbItemAtomicCas(&rb.tail, tail, next)
			} else {
				// 不相等 说明 可以取
				/*
					这里取的是next的value 是因为 我们在初始化的使用有一个头节点。所以 head.next 是有效节点
					我们将 next.value 返回后，就将 next 节点作为新的头节点
				*/
				v := next.value

				if rbItemAtomicCas(&rb.head, head, next) {
					atomic.AddInt32(&rb.length, -1)
					rbItemAtomicCas(&head.next, next, nil)
					return v, nil
				}
			}

		}

	}

}

/*
Len() uint32
	Cap() uint32
	Reset()

*/

func (rb *ringBufList) Len() uint32 {
	return uint32(atomic.LoadInt32(&rb.length))
}
func (rb *ringBufList) Cap() uint32 {
	return rb.Len()
}

func (rb *ringBufList) Reset() {

}
