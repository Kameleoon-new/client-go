package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Kameleoon/client-go/v3/logging"
)

type DataMapStorage[K comparable, V any] interface {
	Enumerate(f func(V) bool)
	Get(key K) V
	Len() int
}

func enumerateMap[K comparable, V any](m map[K]V, f func(V) bool) bool {
	for _, v := range m {
		if !f(v) {
			return false
		}
	}
	return true
}

type DataMapStorageImpl[K comparable, V any] struct {
	mx   *sync.RWMutex
	data *map[K]V
}

func NewDataMapStorageImpl[K comparable, V any](mx *sync.RWMutex, data *map[K]V) *DataMapStorageImpl[K, V] {
	return &DataMapStorageImpl[K, V]{
		mx:   mx,
		data: data,
	}
}

func (s DataMapStorageImpl[K, V]) String() string {
	var values []string
	s.Enumerate(func(v V) bool {
		values = append(values, logging.ObjectToString(v))
		return true
	})
	return fmt.Sprintf("DataMapStorageImpl{values:%s}", strings.Join(values, ","))
}

func (s *DataMapStorageImpl[K, V]) Enumerate(f func(V) bool) {
	if s.data != nil {
		s.mx.RLock()
		defer s.mx.RUnlock()
		enumerateMap[K, V](*s.data, f)
	}
}

func (s *DataMapStorageImpl[K, V]) Get(key K) V {
	if s.data != nil {
		s.mx.RLock()
		defer s.mx.RUnlock()
		if out, contains := (*s.data)[key]; contains {
			return out
		}
	}
	var defaultV V
	return defaultV
}

func (s *DataMapStorageImpl[K, V]) Len() int {
	if s.data != nil {
		s.mx.RLock()
		defer s.mx.RUnlock()
		return len(*s.data)
	}
	return 0
}
