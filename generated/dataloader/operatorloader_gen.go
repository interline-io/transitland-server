// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloader

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// OperatorLoaderConfig captures the config to create a new OperatorLoader
type OperatorLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []int) ([]*model.Operator, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewOperatorLoader creates a new OperatorLoader given a fetch, wait, and maxBatch
func NewOperatorLoader(config OperatorLoaderConfig) *OperatorLoader {
	return &OperatorLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// OperatorLoader batches and caches requests
type OperatorLoader struct {
	// this method provides the data for the loader
	fetch func(keys []int) ([]*model.Operator, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[int]*model.Operator

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *operatorLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type operatorLoaderBatch struct {
	keys    []int
	data    []*model.Operator
	error   []error
	closing bool
	done    chan struct{}
}

// Load a Operator by key, batching and caching will be applied automatically
func (l *OperatorLoader) Load(key int) (*model.Operator, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a Operator.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *OperatorLoader) LoadThunk(key int) func() (*model.Operator, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (*model.Operator, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &operatorLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (*model.Operator, error) {
		<-batch.done

		var data *model.Operator
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *OperatorLoader) LoadAll(keys []int) ([]*model.Operator, []error) {
	results := make([]func() (*model.Operator, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	operators := make([]*model.Operator, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		operators[i], errors[i] = thunk()
	}
	return operators, errors
}

// LoadAllThunk returns a function that when called will block waiting for a Operators.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *OperatorLoader) LoadAllThunk(keys []int) func() ([]*model.Operator, []error) {
	results := make([]func() (*model.Operator, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]*model.Operator, []error) {
		operators := make([]*model.Operator, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			operators[i], errors[i] = thunk()
		}
		return operators, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *OperatorLoader) Prime(key int, value *model.Operator) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
		// and end up with the whole cache pointing to the same value.
		cpy := *value
		l.unsafeSet(key, &cpy)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *OperatorLoader) Clear(key int) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *OperatorLoader) unsafeSet(key int, value *model.Operator) {
	if l.cache == nil {
		l.cache = map[int]*model.Operator{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *operatorLoaderBatch) keyIndex(l *OperatorLoader, key int) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *operatorLoaderBatch) startTimer(l *OperatorLoader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *operatorLoaderBatch) end(l *OperatorLoader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
