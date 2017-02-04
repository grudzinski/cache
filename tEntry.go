package cache

import (
	"container/list"
	"sync/atomic"
)

type tEntry struct {
	value   atomic.Value
	hit     uint32
	element *list.Element
}
