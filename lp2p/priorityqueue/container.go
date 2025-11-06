package priorityqueue

import (
	"container/heap"
)

type item struct {
	value    any
	priority uint64
}

// convenient wrapper around heap.Interface
type container struct {
	comparator func(a, b uint64) bool
	items      []item
}

var _ heap.Interface = &container{}

func newContainer(maxHeap bool, cap int) *container {
	c := &container{
		items: make([]item, 0, cap),
	}

	c.comparator = minHeapComparator
	if maxHeap {
		c.comparator = maxHeapComparator
	}

	heap.Init(c)

	return c
}

func (c *container) PushItem(v any, priority uint64) {
	heap.Push(c, item{
		value:    v,
		priority: priority,
	})
}

func (c *container) PopItem() (v any, ok bool) {
	if c.Len() == 0 {
		return nil, false
	}

	raw := heap.Pop(c)

	i, ok := raw.(item)
	if !ok {
		panic("invalid type")
	}

	return i.value, true
}

func (c *container) Len() int {
	return len(c.items)
}

func (c *container) Less(i, j int) bool {
	return c.comparator(c.items[i].priority, c.items[j].priority)
}

func (c *container) Push(v any) {
	i, ok := v.(item)
	if !ok {
		panic("invalid item type")
	}

	c.items = append(c.items, i)
}

func (c *container) Pop() any {
	n := len(c.items)
	if n == 0 {
		panic("unable to Pop from an empty container")
	}

	lastItem := c.items[n-1]
	c.items = c.items[:n-1]

	return lastItem
}

func (c *container) Swap(i, j int) {
	c.items[j], c.items[i] = c.items[i], c.items[j]
}

func minHeapComparator(a, b uint64) bool {
	return a < b
}

func maxHeapComparator(a, b uint64) bool {
	return a > b
}
