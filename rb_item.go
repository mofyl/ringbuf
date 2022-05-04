package ringbuf

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/sys/cpu"
)

type RbItemFlagType = uint32

var (
	Initialization RbItemFlagType = 1 // 表示 data[i] 位置 可以写 初始状态
	Used           RbItemFlagType = 2 // 表示 data[i] 位置 写入完毕 是 Initialization -> Used -> Dequeue
	Dequeue        RbItemFlagType = 3 // 表示 data[i] 已经写入有效数据，并且当前可以 dequeue
	DequeueDone    RbItemFlagType = 4 // 表示 data[i] 位置 刚 dequeue Dequeue -> DequeueDone -> Initialization

	MaxUint32 = ^uint32(0) // head 和tail 的中间状态
)

type rbItem struct {
	_ cpu.CacheLinePad

	value any
	flag  RbItemFlagType
}

func (e *rbItem) Init() {
	e.flag = Initialization
	e.value = nil
}

func (e *rbItem) Reset() {
	e.Init()
}

func (e *rbItem) enqueue(item any) error {

	if !atomic.CompareAndSwapUint32(&e.flag, Initialization, Used) {
		if atomic.LoadUint32(&e.flag) == Initialization {
			return ErrItemNeedReTry
		}
		return ErrItemGosched
	}

	e.value = item

	if !atomic.CompareAndSwapUint32(&e.flag, Used, Dequeue) {
		fmt.Println(11111)
	}

	return nil
}

func (e *rbItem) dequeue() (any, error) {
	if !atomic.CompareAndSwapUint32(&e.flag, Dequeue, DequeueDone) {
		if atomic.LoadUint32(&e.flag) == Dequeue {
			return nil, ErrItemNeedReTry
		}
		return nil, ErrItemGosched
	}

	item := e.value
	e.value = nil

	if !atomic.CompareAndSwapUint32(&e.flag, DequeueDone, Initialization) {
		fmt.Println(22222)
	}

	return item, nil
}
