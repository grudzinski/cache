package cache

import "sync/atomic"

type tEntry struct {
	value atomic.Value
	hit   uint32
}
