// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Cache is like a Go sync.Map but with generic is safe for concurrent use
// by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
//
// The Cache type is specialized. Most code should use a plain Go map instead,
// with separate locking or coordination, for better type safety and to make it
// easier to maintain other invariants along with the map content.
//
// The Cache type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate Mutex or RWMutex.
//
// The zero Cache is empty and ready for use. A Cache must not be copied after first use.
//
// In the terminology of the Go memory model, Cache arranges that a write operation
// “synchronizes before” any read operation that observes the effect of the write, where
// read and write operations are defined as follows.
// Load, LoadAndDelete, LoadOrStore, Swap, CompareAndSwap, and CompareAndDelete
// are read operations; Delete, LoadAndDelete, Store, and Swap are write operations;
// LoadOrStore is a write operation when it returns loaded set to false;
// CompareAndSwap is a write operation when it returns swapped set to true;
// and CompareAndDelete is a write operation when it returns deleted set to true.
type Cache[K comparable, E any] struct {
	mu sync.Mutex

	// read contains the portion of the cache's contents that are safe for
	// concurrent access (with or without mu held).
	//
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	//
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-expunged entry requires that the entry be copied to the dirty
	// map and unexpunged with mu held.
	read atomic.Pointer[readOnly[K, E]]

	// dirty contains the portion of the cache's contents that require mu to be
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// it also includes all of the non-expunged entries in the read map.
	//
	// Expunged entries are not stored in the dirty map. An expunged entry in the
	// clean map must be unexpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	dirty map[K]*entry[E]

	// misses counts the number of loads since the read map was last updated that
	// needed to lock mu to determine whether the key was present.
	//
	// Once enough misses have occurred to cover the cost of copying the dirty
	// map, the dirty map will be promoted to the read map (in the unamended
	// state) and the next store to the cache will make a new dirty copy.
	misses int
}

// ComparableCache is like Cache but its element type restricted by comparable.
type ComparableCache[K, E comparable] struct {
	Cache[K, E]
}

// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly[K comparable, E any] struct {
	m       map[K]*entry[E]
	amended bool // true if the dirty map contains some key not in m.
}

// expunged is an arbitrary pointer that marks entries which have been deleted
// from the dirty map.
var expunged = unsafe.Pointer(new(any))

// An entry is a slot in the cache corresponding to a particular key.
type entry[E any] struct {
	// p points to the E value stored for the entry.
	//
	// If p == nil, the entry has been deleted, and either m.dirty == nil or
	// m.dirty[key] is e.
	//
	// If p == expunged, the entry has been deleted, m.dirty != nil, and the entry
	// is missing from m.dirty.
	//
	// Otherwise, the entry is valid and recorded in m.read.m[key] and, if m.dirty
	// != nil, in m.dirty[key].
	//
	// An entry can be deleted by atomic replacement with nil: when m.dirty is
	// next created, it will atomically replace nil with expunged and leave
	// m.dirty[key] unset.
	//
	// An entry's associated value can be updated by atomic replacement, provided
	// p != expunged. If p == expunged, an entry's associated value can be updated
	// only after first setting m.dirty[key] = e so that lookups using the dirty
	// map find the entry.
	p atomic.Pointer[E]
}

type comparableEntry[E comparable] struct {
	entry[E]
}

func newEntry[E any](a E) *entry[E] {
	e := &entry[E]{}
	e.p.Store(&a)
	return e
}

func newComparableEntry[E comparable](a E) *comparableEntry[E] {
	e := &comparableEntry[E]{}
	e.p.Store(&a)
	return e
}

func (c *Cache[K, E]) loadReadOnly() readOnly[K, E] {
	if p := c.read.Load(); p != nil {
		return *p
	}

	return readOnly[K, E]{}
}

// Load returns the value stored in the cache for a key, or zero value if no
// value is present.
// The ok result indicates whether value was found in the cache.
func (c *Cache[K, E]) Load(key K) (value E, ok bool) {
	read := c.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		c.mu.Lock()

		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read = c.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = c.dirty[key]

			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			c.missLocked()
		}
		c.mu.Unlock()
	}

	if !ok {
		return value, false
	}

	return e.load()
}

func (e *entry[E]) load() (value E, ok bool) {
	p := e.p.Load()
	if nil == p || (*E)(expunged) == p {
		return value, false
	}

	return *p, true
}

// Store sets the value for a key.
func (c *Cache[K, E]) Store(key K, value E) {
	_, _ = c.Swap(key, value)
}

// tryCompareAndSwap compare the entry with the given old value and swaps
// it with a new value if the entry is equal to the old value, and the entry
// has not been expunged.
//
// If the entry is expunged, tryCompareAndSwap returns false and leaves
// the entry unchanged.
func (e *comparableEntry[E]) tryCompareAndSwap(old, new E) bool {
	p := e.p.Load()
	if nil == p || (*E)(expunged) == p || *p != old {
		return false
	}

	// Copy the interface after the first load to make this method more amenable
	// to escape analysis: if the comparison fails from the start, we shouldn't
	// bother heap-allocating an interface value to store.
	nc := new
	for {
		if e.p.CompareAndSwap(p, &nc) {
			return true
		}

		p = e.p.Load()
		if nil == p || (*E)(expunged) == p || *p != old {
			return false
		}
	}
}

// unexpungeLocked ensures that the entry is not marked as expunged.
//
// If the entry was previously expunged, it must be added to the dirty map
// before m.mu is unlocked.
func (e *entry[E]) unexpungeLocked() (wasExpunged bool) {
	return e.p.CompareAndSwap((*E)(expunged), nil)
}

// swapLocked unconditionally swaps a value into the entry.
//
// The entry must be known not to be expunged.
func (e *entry[E]) swapLocked(a *E) *E {
	return e.p.Swap(a)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (c *Cache[K, E]) LoadOrStore(key K, value E) (actual E, loaded bool) {
	// Avoid locking if it's a clean hit.
	read := c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	c.mu.Lock()
	read = c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			c.dirty[key] = e
		}

		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := c.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		c.missLocked()
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			c.dirtyLocked()
			c.read.Store(&readOnly[K, E]{m: read.m, amended: true})
		}

		c.dirty[key] = newEntry(value)
		actual, loaded = value, false
	}
	c.mu.Unlock()

	return actual, loaded
}

func (c *ComparableCache[K, E]) LoadOrStore(key K, value E) (actual E, loaded bool) {
	// Avoid locking if it's a clean hit.
	read := c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	c.mu.Lock()
	read = c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			c.dirty[key] = e
		}

		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := c.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		c.missLocked()
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			c.dirtyLocked()
			c.read.Store(&readOnly[K, E]{m: read.m, amended: true})
		}

		c.dirty[key] = (*entry[E])((unsafe.Pointer)(newComparableEntry(value)))
		actual, loaded = value, false
	}
	c.mu.Unlock()

	return actual, loaded
}

// tryLoadOrStore atomically loads or stores a value if the entry is not
// expunged.
//
// If the entry is expunged, tryLoadOrStore leaves the entry unchanged and
// returns with ok==false.
func (e *entry[E]) tryLoadOrStore(a E) (actual E, loaded, ok bool) {
	p := e.p.Load()
	if (*E)(expunged) == p {
		return actual, false, false
	}

	if p != nil {
		return *p, true, true
	}

	// Copy the interface after the first load to make this method more amenable
	// to escape analysis: if we hit the "load" path or the entry is expunged, we
	// shouldn't bother heap-allocating.
	ac := a
	for {
		if e.p.CompareAndSwap(nil, &ac) {
			return a, false, true
		}

		p = e.p.Load()
		if (*E)(expunged) == p {
			return actual, false, false
		}

		if p != nil {
			return *p, true, true
		}
	}
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (c *Cache[K, E]) LoadAndDelete(key K) (value E, loaded bool) {
	read := c.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		c.mu.Lock()
		read = c.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = c.dirty[key]
			delete(c.dirty, key)

			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			c.missLocked()
		}
		c.mu.Unlock()
	}

	if ok {
		return e.delete()
	}

	return value, false
}

// Delete deletes the value for a key.
func (c *Cache[K, E]) Delete(key K) {
	_, _ = c.LoadAndDelete(key)
}

func (e *entry[E]) delete() (value E, ok bool) {
	for {
		p := e.p.Load()
		if nil == p || (*E)(expunged) == p {
			return value, false
		}

		if e.p.CompareAndSwap(p, nil) {
			return *p, true
		}
	}
}

// trySwap swaps a value if the entry has not been expunged.
//
// If the entry is expunged, trySwap returns false and leaves the entry
// unchanged.
func (e *entry[E]) trySwap(a *E) (*E, bool) {
	for {
		p := e.p.Load()
		if (*E)(expunged) == p {
			return nil, false
		}

		if e.p.CompareAndSwap(p, a) {
			return p, true
		}
	}
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (c *Cache[K, E]) Swap(key K, value E) (previous E, loaded bool) {
	read := c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if nil == v {
				return previous, false
			}

			return *v, true
		}
	}

	c.mu.Lock()
	read = c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			c.dirty[key] = e
		}

		if v := e.swapLocked(&value); v != nil {
			previous = *v
			loaded = true
		}
	} else if e, ok := c.dirty[key]; ok {
		if v := e.swapLocked(&value); v != nil {
			previous = *v
			loaded = true
		}
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			c.dirtyLocked()
			c.read.Store(&readOnly[K, E]{m: read.m, amended: true})
		}

		c.dirty[key] = newEntry(value)
	}
	c.mu.Unlock()

	return previous, loaded
}

func (c *ComparableCache[K, E]) Swap(key K, value E) (previous E, loaded bool) {
	read := c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if nil == v {
				return previous, false
			}

			return *v, true
		}
	}

	c.mu.Lock()
	read = c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			c.dirty[key] = e
		}

		if v := e.swapLocked(&value); v != nil {
			previous = *v
			loaded = true
		}
	} else if e, ok := c.dirty[key]; ok {
		if v := e.swapLocked(&value); v != nil {
			previous = *v
			loaded = true
		}
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			c.dirtyLocked()
			c.read.Store(&readOnly[K, E]{m: read.m, amended: true})
		}

		c.dirty[key] = (*entry[E])((unsafe.Pointer)(newComparableEntry(value)))
	}
	c.mu.Unlock()

	return previous, loaded
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the cache is equal to old.
func (c *ComparableCache[K, E]) CompareAndSwap(key K, old, new E) bool {
	read := c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		return (*comparableEntry[E])((unsafe.Pointer)(e)).tryCompareAndSwap(old, new)
	} else if !read.amended {
		return false // No existing value for key.
	}

	var swapped bool
	c.mu.Lock()
	read = c.loadReadOnly()
	if e, ok := read.m[key]; ok {
		swapped = (*comparableEntry[E])((unsafe.Pointer)(e)).tryCompareAndSwap(old, new)
	} else if e, ok := c.dirty[key]; ok {
		swapped = (*comparableEntry[E])((unsafe.Pointer)(e)).tryCompareAndSwap(old, new)

		// We needed to lock mu in order to load the entry for key,
		// and the operation didn't change the set of keys in the map
		// (so it would be made more efficient by promoting the dirty
		// map to read-only).
		// Count it as a miss so that we will eventually switch to the
		// more efficient steady state.
		c.missLocked()
	}
	c.mu.Unlock()

	return swapped
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
//
// If there is no current value for key in the cache, CompareAndDelete
// returns false.
func (c *ComparableCache[K, E]) CompareAndDelete(key K, old E) (deleted bool) {
	read := c.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		c.mu.Lock()
		read = c.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = c.dirty[key]
			// Don't delete key from m.dirty: we still need to do the “compare” part
			// of the operation. The entry will eventually be expunged when the
			// dirty map is promoted to the read map.
			//
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			c.missLocked()
		}
		c.mu.Unlock()
	}

	if ok {
		p := e.p.Load()
		if nil == p || (*E)(expunged) == p || *p != old {
			return false
		}

		if e.p.CompareAndSwap(p, nil) {
			return true
		}
	}

	return false
}

// Range calls f sequentially for each key and value present in the cache.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Cache's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on c.
//
// Range may be O(N) with the number of elements in the cache even if f returns
// false after a constant number of calls.
func (c *Cache[K, E]) Range(f func(key K, value E) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read := c.loadReadOnly()
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		c.mu.Lock()
		read = c.loadReadOnly()
		if read.amended {
			read = readOnly[K, E]{m: c.dirty}
			c.read.Store(&read)
			c.dirty = nil
			c.misses = 0
		}
		c.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}

		if !f(k, v) {
			break
		}
	}
}

func (c *Cache[K, E]) missLocked() {
	c.misses++
	if c.misses < len(c.dirty) {
		return
	}

	c.read.Store(&readOnly[K, E]{m: c.dirty})
	c.dirty = nil
	c.misses = 0
}

func (c *Cache[K, E]) dirtyLocked() {
	if c.dirty != nil {
		return
	}

	read := c.loadReadOnly()
	c.dirty = make(map[K]*entry[E], len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			c.dirty[k] = e
		}
	}
}

func (e *entry[E]) tryExpungeLocked() (isExpunged bool) {
	p := e.p.Load()
	for nil == p {
		if e.p.CompareAndSwap(nil, (*E)(expunged)) {
			return true
		}

		p = e.p.Load()
	}

	return (*E)(expunged) == p
}
