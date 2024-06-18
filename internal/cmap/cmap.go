package cmap

import (
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// CMap is a goroutine-safe map.
type CMap struct {
	m map[string]any
	l cmtsync.RWMutex
}

func NewCMap() *CMap {
	return &CMap{
		m: make(map[string]any),
	}
}

func (cm *CMap) Set(key string, value any) {
	cm.l.Lock()
	cm.m[key] = value
	cm.l.Unlock()
}

func (cm *CMap) Get(key string) any {
	cm.l.RLock()
	val := cm.m[key]
	cm.l.RUnlock()
	return val
}

func (cm *CMap) Has(key string) bool {
	cm.l.RLock()
	_, ok := cm.m[key]
	cm.l.RUnlock()
	return ok
}

func (cm *CMap) Delete(key string) {
	cm.l.Lock()
	delete(cm.m, key)
	cm.l.Unlock()
}

func (cm *CMap) Size() int {
	cm.l.RLock()
	size := len(cm.m)
	cm.l.RUnlock()
	return size
}

func (cm *CMap) Clear() {
	cm.l.Lock()
	cm.m = make(map[string]any)
	cm.l.Unlock()
}

func (cm *CMap) Keys() []string {
	cm.l.RLock()
	keys := make([]string, 0, len(cm.m))
	for k := range cm.m {
		keys = append(keys, k)
	}
	cm.l.RUnlock()
	return keys
}

func (cm *CMap) Values() []any {
	cm.l.RLock()
	items := make([]any, 0, len(cm.m))
	for _, v := range cm.m {
		items = append(items, v)
	}
	cm.l.RUnlock()
	return items
}
