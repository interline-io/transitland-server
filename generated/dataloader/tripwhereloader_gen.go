// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloader

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// TripWhereLoaderConfig captures the config to create a new TripWhereLoader
type TripWhereLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []model.TripParam) ([][]*model.Trip, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewTripWhereLoader creates a new TripWhereLoader given a fetch, wait, and maxBatch
func NewTripWhereLoader(config TripWhereLoaderConfig) *TripWhereLoader {
	return &TripWhereLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// TripWhereLoader batches and caches requests
type TripWhereLoader struct {
	// this method provides the data for the loader
	fetch func(keys []model.TripParam) ([][]*model.Trip, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[model.TripParam][]*model.Trip

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *tripWhereLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type tripWhereLoaderBatch struct {
	keys    []model.TripParam
	data    [][]*model.Trip
	error   []error
	closing bool
	done    chan struct{}
}

// Load a Trip by key, batching and caching will be applied automatically
func (l *TripWhereLoader) Load(key model.TripParam) ([]*model.Trip, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a Trip.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *TripWhereLoader) LoadThunk(key model.TripParam) func() ([]*model.Trip, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() ([]*model.Trip, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &tripWhereLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() ([]*model.Trip, error) {
		<-batch.done

		var data []*model.Trip
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
func (l *TripWhereLoader) LoadAll(keys []model.TripParam) ([][]*model.Trip, []error) {
	results := make([]func() ([]*model.Trip, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	trips := make([][]*model.Trip, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		trips[i], errors[i] = thunk()
	}
	return trips, errors
}

// LoadAllThunk returns a function that when called will block waiting for a Trips.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *TripWhereLoader) LoadAllThunk(keys []model.TripParam) func() ([][]*model.Trip, []error) {
	results := make([]func() ([]*model.Trip, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([][]*model.Trip, []error) {
		trips := make([][]*model.Trip, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			trips[i], errors[i] = thunk()
		}
		return trips, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *TripWhereLoader) Prime(key model.TripParam, value []*model.Trip) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
		// and end up with the whole cache pointing to the same value.
		cpy := make([]*model.Trip, len(value))
		copy(cpy, value)
		l.unsafeSet(key, cpy)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *TripWhereLoader) Clear(key model.TripParam) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *TripWhereLoader) unsafeSet(key model.TripParam, value []*model.Trip) {
	if l.cache == nil {
		l.cache = map[model.TripParam][]*model.Trip{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *tripWhereLoaderBatch) keyIndex(l *TripWhereLoader, key model.TripParam) int {
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

func (b *tripWhereLoaderBatch) startTimer(l *TripWhereLoader) {
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

func (b *tripWhereLoaderBatch) end(l *TripWhereLoader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
