package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := NewCache()
	c.Set("foo", "bar", 1*time.Minute)
	value, _, found := c.Get("foo")
	if !found {
		t.FailNow()
	}
	if value != "bar" {
		t.FailNow()
	}
}
