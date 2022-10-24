package lru_cache

import (
	"errors"
)

// TODO: specify type instead of any
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
// and evicts any items over capacity. The
func (list *doublyLinkedList) Prepend(newNode *node) {
	// List is empty
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
		newNode.prev, newNode.next = nil, nil
		list.count = 1
		return
	}

	head := list.head
	newNode.next = head
	head.prev = newNode
	list.head = newNode
	list.count++
}

func (list *doublyLinkedList) RotateFront(target *node) {
	// Node is already at front, simply return
	if target.prev == nil {
		return
	}

	// TODO: revist for simplification. This func is stateful since the list
	// could be in an invalid state (empty) when called. This error case might
	// be prevented by refactor. The general case is when called with a node not in list
	// (perhaps removed previously, or other stateful ops).

	target.prev.next = target.next
	target.next.prev = target.prev
	list.count--
	list.Prepend(target)
}

// Slice the list at the zero-based nth position and return the first node from that position.
func (list *doublyLinkedList) TrimRight(n int) (evicted *node) {
	// NOt at capacity, so just return.
	if list.count <= n {
		return
	}

	// TODO: reconsider list w/out count variable. I don't like trusting
	// that I can iterate the list to nth position w/out nil checks.
	evicted = list.head
	for i := 0; i < n; i++ {
		evicted = evicted.next
	}

	list.tail = evicted.prev
	list.tail.next = nil
	evicted.prev = nil

	// Check if evicted is the only item in the list.
	if list.head == evicted {
		list.head = nil
	}

	return
}

var errItemNil error = errors.New("node cannot be nil")

// Remove removes the passed list node from the list and returns an
// error if target it nil, otherwise returns nil on success. If successful,
// no longer use the passed node to allow it to be removed.
func (list *doublyLinkedList) Remove(target *node) (err error) {
	defer func() {
		// If no error, nullify target's pointers to prevent memory leaks via stale references.
		if err == nil {
			target.prev = nil
			target.next = nil
			list.count--
		}
	}()

	if target == nil {
		return errItemNil
	}

	// Target is the only list item
	if target.prev == nil && target.next == nil {
		list.head = nil
		list.tail = nil
		return
	}
	// Target is the first item in the list
	if target.prev == nil {
		list.head = target.next
		return
	}
	// Target is last item in the list
	if target.next == nil {
		list.tail = target.prev
		return
	}
	// Target is in the middle of a list containing multiple items
	prev := target.prev
	next := target.next
	prev.next = next
	next.prev = prev

	return
}

// CacheKey must be a type that resolves to either int or string.
// The logic is that the key should resemble a primary key in a database,
// a C#-style hashcode, or GUID. Constraining to these use-cases enforces
// strong object-identity patterns.
type CacheKey interface {
	~int | ~string
}

// A CacheObject implements an ID() method for use as a map key.
type CacheObject interface {
	// ID() returns an efficient object id for use as a map key.
	ID() int
}

// Cache is a least-recently-used cache.
type Cache struct {
	// TODO: locking
	itemMap  map[int]*node
	itemList *doublyLinkedList
	size     int
}

var ErrInvalidSize error = errors.New("invalid cache size")

func NewCache(size int) (*Cache, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	return &Cache{
		itemMap:  make(map[int]*node, size),
		itemList: newDoublyLinkedList(),
		size:     size,
	}, nil
}

var ErrDuplicateItem error = errors.New("duplicate item")

// Add adds the passed item to the cache, returning an error
// if the insertion failed.
func (cache *Cache) Add(item CacheObject) (err error) {
	if _, ok := cache.itemMap[item.ID()]; ok {
		err = ErrDuplicateItem
		return
	}

	newNode := &node{
		item: item,
	}

	// TODO: error handling on insertion
	// TODO: locking

	// Add the item to the front of the list
	cache.itemList.Prepend(newNode)
	// Store the item in fast lookup
	cache.itemMap[item.ID()] = newNode

	// Evict any nodes over capacity
	evicted := cache.itemList.TrimRight(cache.size)
	for evicted != nil {
		// TODO: map size is not reduced after deletion, a memory leak.
		delete(cache.itemMap, evicted.item.ID())
		evicted.prev = nil
		evicted = evicted.next
	}

	return
}

// Get finds the passed item and returns it if it exists.
// If found, the item is rotated to the front of the cache.
func (cache *Cache) Get(id int) (item CacheObject, exists bool) {
	var target *node
	target, exists = cache.itemMap[id]
	if !exists {
		return
	}

	// Rotate item to front of list
	cache.itemList.RotateFront(target)
	item = target.item

	return
}

var ErrItemNotFound error = errors.New("item id not found")

func (cache *Cache) Remove(id int) error {
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
