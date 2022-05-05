package ringbuf

import (
	"sync/atomic"
	"unsafe"
)

type rbItemList struct {
	value any
	next  unsafe.Pointer
}

func atomicLoadToRbItem(p *unsafe.Pointer) *rbItemList {
	return (*rbItemList)(atomic.LoadPointer(p))
}

func rbItemAtomicCas(p *unsafe.Pointer, old, new *rbItemList) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
