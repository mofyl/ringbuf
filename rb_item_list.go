package ringbuf

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	itemListPool = &sync.Pool{New: func() any {
		return &rbItemList{}
	}}
)

type rbItemList struct {
	value any
	next  unsafe.Pointer
}

func getRbItemNode() *rbItemList {

	return itemListPool.Get().(*rbItemList)
}

func putRbItemNode(item *rbItemList) {
	itemListPool.Put(item)
}

func atomicLoadToRbItem(p *unsafe.Pointer) *rbItemList {
	return (*rbItemList)(atomic.LoadPointer(p))
}

func rbItemAtomicCas(p *unsafe.Pointer, old, new *rbItemList) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
