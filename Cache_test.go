package cache

import "testing"

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.FailNow()
	}
	if c.keyToEntry == nil {
		t.FailNow()
	}
}
