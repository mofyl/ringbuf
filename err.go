package ringbuf

import "errors"

var (
	ErrQueueFull     = errors.New("queue full")
	ErrQueueEmpty    = errors.New("queue empty")
	ErrRaced         = errors.New("queue race")
	ErrQueueNotReady = errors.New("queue not ready")

	ErrItemNeedReTry = errors.New("queue item need re try")
	ErrItemGosched   = errors.New("queue item go sched")
)
