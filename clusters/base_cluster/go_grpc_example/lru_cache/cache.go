// This is an lru cache implementation for interview practice.
// Do not use, use hashicorp or another implementation instead;
// for instance see: https://github.com/golang/groupcache/blob/master/lru/lru.go
// Caches comes in many different flavors and modifications.
// TODO/FUTURE: implement a store-backed lru  into which a store
// (postgres, minio, etc) could be injected.

package lru_cache

import (
	"errors"
	"sync"
)

var (
	// ErrInvalidCache size is returned if the cache size is not positive.
	ErrInvalidSize error = errors.New("invalid cache size")
	// ErrDuplicateItem is returned when attempting to Put a dupe.
	ErrDuplicateItem error = errors.New("duplicate item")
	// ErrItemNotFound is returned when an object id is not in the cache.
	ErrItemNotFound error = errors.New("item id not found")
)

// CacheObject implements an ID() method for use as a map key.
type CacheObject interface {
	// ID() returns an int for use as a map key.
	ID() int
}

// Cache is a least-recently-used cache.
type Cache struct {
	// TODO: locking
	itemMap  map[int]*node
	itemList *doublyLinkedList
	capacity int
	mu       sync.RWMutex
}

// NewCache initializes a cache of the passed capacity.
func NewCache(capacity int) (*Cache, error) {
	if capacity <= 0 {
		return nil, ErrInvalidSize
	}

	return &Cache{
		itemMap:  make(map[int]*node, capacity),
		itemList: newDoublyLinkedList(),
		capacity: capacity,
		mu:       sync.RWMutex{},
	}, nil
}

// Put adds the passed item to the cache and evicts old items.
// Put returns an error if the insertion failed or the object already exists.
func (cache *Cache) Put(item CacheObject) (err error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, ok := cache.itemMap[item.ID()]; ok {
		err = ErrDuplicateItem
		return
	}

	newNode := &node{
		item: item,
	}

	// TODO: error handling on insertion
	// TODO: verify if indices are off by one (e.g. list evicts too many/few nodes)

	// Add the item to the front of the list
	cache.itemList.Prepend(newNode)
	// Store the item in hash table
	cache.itemMap[item.ID()] = newNode

	// Evict least-recently-used nodes over capacity
	evicted := cache.itemList.TrimRight(cache.capacity)
	for evicted != nil {
		// TODO: underlying map size is not reduced after deletion, a memory leak.
		delete(cache.itemMap, evicted.item.ID())
		evicted.prev = nil
		evicted = evicted.next
	}

	return
}

// Get finds the passed item and returns it if it exists.
// If found, the item is rotated to the front of the cache.
func (cache *Cache) Get(id int) (item CacheObject, exists bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	var target *node
	target, exists = cache.itemMap[id]
	if !exists {
		return
	}

	// Rotate item to front of list
	_ = cache.itemList.RotateFront(target)
	item = target.item

	return
}

func (cache *Cache) Remove(id int) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	target, ok := cache.itemMap[id]
	if !ok {
		return ErrItemNotFound
	}

	if err := cache.itemList.Remove(target); err != nil {
		return err
	}

	delete(cache.itemMap, target.item.ID())

	return nil
}

type node struct {
	next *node
	prev *node
	item CacheObject
}

type doublyLinkedList struct {
	head  *node
	tail  *node
	count int
}

func newDoublyLinkedList() *doublyLinkedList {
	return &doublyLinkedList{
		head:  nil,
		tail:  nil,
		count: 0,
	}
}

// Prepend inserts the passed node to the front of the list
// and evicts any items over capacity.
func (list *doublyLinkedList) Prepend(newNode *node) {
	// List is empty
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
		newNode.prev, newNode.next = nil, nil
		list.count = 1
		return
	}

	newNode.next = list.head
	list.head.prev = newNode
	list.head = newNode
	list.count++
}

func (list *doublyLinkedList) RotateFront(target *node) (err error) {
	if target == nil {
		return errItemNil
	}

	// Node is already at front, simply return
	if target.prev == nil {
		return
	}

	// TODO: revisit for simplification. This func is stateful since the list
	// could be in an invalid state (empty) when called. This error case might
	// be prevented by refactor. The general case is when called with a node not in list
	// (perhaps removed previously, or other stateful ops).

	_ = list.Remove(target)
	list.Prepend(target)

	return
}

// Slice the list at the zero-based nth position and return the first node from that position.
func (list *doublyLinkedList) TrimRight(n int) (evicted *node) {
	// Not at capacity, so just return.
	if list.count <= n {
		return
	}

	// TODO: reconsider list w/out count variable. I don't like trusting
	// that I can iterate the list to nth position w/out nil checks.
	evicted = list.head
	for i := 0; i < n; i++ {
		evicted = evicted.next
	}

	// Evicted is the first node in the list
	if evicted == list.head {
		list.head = nil
		list.tail = nil
		list.count = 0
		return
	}

	// Evicted is the last item in the list
	if evicted == list.tail {
		list.tail.prev.next = nil
		list.tail = list.tail.prev
		evicted.prev = nil
		list.count--
		return
	}

	// Else: evicted is some node in the middle
	list.tail = evicted.prev
	list.tail.next = nil
	evicted.prev = nil
	list.count = n

	return
}

var errItemNil error = errors.New("node cannot be nil")

// Remove removes the passed list node from the list and returns an
// error if target is nil, otherwise returns nil on success.
// If successful, no longer use the passed node to allow it to be removed.
func (list *doublyLinkedList) Remove(target *node) (err error) {
	if target == nil {
		return errItemNil
	}

	defer func() {
		// If no error, nullify target's pointers to prevent memory leaks via stale references.
		if err == nil {
			target.prev = nil
			target.next = nil
			list.count--
		}
	}()

	// Target is the only list item
	if target.prev == nil && target.next == nil {
		list.head = nil
		list.tail = nil
		return
	}
	// Target is the first item in a list with successors.
	if target.prev == nil {
		list.head = target.next
		return
	}
	// Target is the last item in a list with predecessors.
	if target.next == nil {
		list.tail = target.prev
		return
	}
	// Target is in the middle of a list with predecessors and successors.
	target.prev.next = target.next
	target.next.prev = target.prev

	return
}
