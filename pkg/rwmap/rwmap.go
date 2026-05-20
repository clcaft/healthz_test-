package rwmap

import (
	"reflect"
	"sync"
)

// MapInterface is the interface Map implements.
type MapInterface[K comparable, V any] interface {
	Load(key K) (V, bool)
	Store(key K, value V)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	LoadAndDelete(key K) (value V, loaded bool)
	Delete(key K)
	Swap(key K, value V) (previous V, loaded bool)
	CompareAndSwap(key K, oldValue, newValue V) (swapped bool)
	CompareAndDelete(key K, oldValue V) (deleted bool)
	Range(func(key K, value V) (shouldContinue bool))
}

type MapWithClearInterface[K comparable, V any] interface {
	MapInterface[K, V]
	// ClearByCallback если callback возвращает true - значение удаляется из мапы
	ClearByCallback(callback func(key K, value V) bool)
}

// RWMutexMap is an implementation of mapInterface using a sync.RWMutex.
type RWMutexMap[K comparable, V any] struct {
	dirty map[K]V
	mu    sync.RWMutex
}

func (m *RWMutexMap[K, V]) Load(key K) (value V, ok bool) {
	m.mu.RLock()
	value, ok = m.dirty[key]
	m.mu.RUnlock()

	return
}

func (m *RWMutexMap[K, V]) Store(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dirty == nil {
		m.dirty = make(map[K]V)
	}

	m.dirty[key] = value
}

func (m *RWMutexMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	actual, loaded = m.dirty[key]

	if loaded {
		return
	}

	actual = value

	if m.dirty == nil {
		m.dirty = make(map[K]V)
	}

	m.dirty[key] = value

	return
}

func (m *RWMutexMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dirty == nil {
		m.dirty = make(map[K]V)
	}

	previous, loaded = m.dirty[key]
	m.dirty[key] = value

	return
}

func (m *RWMutexMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, loaded = m.dirty[key]
	if !loaded {
		return
	}

	delete(m.dirty, key)

	return
}

func (m *RWMutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	delete(m.dirty, key)
	m.mu.Unlock()
}

func (m *RWMutexMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) (swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dirty == nil {
		return false
	}

	value, loaded := m.dirty[key]
	if loaded && reflect.DeepEqual(value, oldValue) {
		m.dirty[key] = newValue

		return true
	}

	return false
}

func (m *RWMutexMap[K, V]) CompareAndDelete(key K, oldValue V) (deleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dirty == nil {
		return false
	}

	value, loaded := m.dirty[key]
	if loaded && reflect.DeepEqual(value, oldValue) {
		delete(m.dirty, key)

		return true
	}

	return false
}

func (m *RWMutexMap[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	m.mu.RLock()

	keys := make([]K, 0, len(m.dirty))
	for k := range m.dirty {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	for _, k := range keys {
		v, ok := m.Load(k)
		if !ok {
			continue
		}

		if !f(k, v) {
			break
		}
	}
}

func (m *RWMutexMap[K, V]) ClearByCallback(callback func(key K, value V) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.dirty {
		if callback(k, v) {
			delete(m.dirty, k)
		}
	}
}
