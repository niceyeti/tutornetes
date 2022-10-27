package lru_cache

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type foo struct {
	id int
}

func (f *foo) ID() int {
	return f.id
}

/*
Full interview strategy:
- delay locking, esoteric error handling with TODOs
- employ test-driven development:
  - focus first on high level api details
  - strongly focus on data objects (the node interface problem)
  - don't write generics, don't bother. Write simple code first, genericize later
  - pivot from test back to code rapidly
  - don't worry terribly about superfluous code paths, mark as TODO
*/
func TestList(t *testing.T) {
	Convey("List tests", t, func() {
		Convey("TrimRight tests", func() {
			l := newDoublyLinkedList()
			nodes := []*node{
				{item: &foo{id: 1}},
				{item: &foo{id: 2}},
				{item: &foo{id: 3}},
			}
			l.Prepend(nodes[2])
			l.Prepend(nodes[1])
			l.Prepend(nodes[0])

			Convey("When list is [1,2,3] and TrimRight(0) is called", func() {
				evicted := l.TrimRight(0)
				So(evicted, ShouldEqual, nodes[0])
				So(l.count, ShouldEqual, 0)
			})

			Convey("When list is [1,2,3] and TrimRight(2) is called", func() {
				evicted := l.TrimRight(2)
				So(evicted, ShouldEqual, nodes[2])
				So(l.count, ShouldEqual, 2)
			})

			Convey("When list is [1,2,3] and TrimRight(1) is called", func() {
				evicted := l.TrimRight(1)
				So(evicted, ShouldEqual, nodes[1])
				So(l.count, ShouldEqual, 1)
			})
		})

		Convey("RotateFront tests", func() {
			l := newDoublyLinkedList()

			Convey("When list is [1,2,3] and RotateFront is called on the last item", func() {
				nodes := []*node{
					{item: &foo{id: 1}},
					{item: &foo{id: 2}},
					{item: &foo{id: 3}},
				}
				l.Prepend(nodes[2])
				l.Prepend(nodes[1])
				l.Prepend(nodes[0])

				err := l.RotateFront(nodes[2])
				So(err, ShouldBeNil)
				So(l.head, ShouldEqual, nodes[2])
				So(l.tail, ShouldEqual, nodes[1])
				So(l.count, ShouldEqual, 3)
			})

			Convey("When nil is passed", func() {
				err := l.RotateFront(nil)
				So(err, ShouldBeError, errItemNil)
			})

			Convey("When only one item is in the list and RotateFront is called", func() {
				item := &node{item: &foo{id: 1}}
				l.Prepend(item)
				err := l.RotateFront(item)
				So(err, ShouldBeNil)
				So(l.head, ShouldEqual, item)
				So(l.tail, ShouldEqual, item)
				So(l.count, ShouldEqual, 1)
			})
		})

		Convey("Initialization tests", func() {
			l := newDoublyLinkedList()
			So(l.count, ShouldEqual, 0)
			So(l.head, ShouldBeNil)
			So(l.tail, ShouldBeNil)
		})

		Convey("Removal tests", func() {
			l := newDoublyLinkedList()
			So(l.count, ShouldEqual, 0)

			nodes := []*node{
				{item: &foo{id: 1}},
				{item: &foo{id: 2}},
				{item: &foo{id: 3}},
			}
			l.Prepend(nodes[2])
			l.Prepend(nodes[1])
			l.Prepend(nodes[0])

			Convey("When last node is removed", func() {
				err := l.Remove(nodes[2])
				So(err, ShouldBeNil)
				So(l.head, ShouldEqual, nodes[0])
				So(l.tail, ShouldEqual, nodes[1])
				So(l.count, ShouldEqual, 2)
			})

			Convey("When first node is removed", func() {
				err := l.Remove(nodes[0])
				So(err, ShouldBeNil)
				So(l.head, ShouldEqual, nodes[1])
				So(l.tail, ShouldEqual, nodes[2])
				So(l.count, ShouldEqual, 2)
			})

			Convey("When middle node is removed", func() {
				err := l.Remove(nodes[1])
				So(err, ShouldBeNil)
				So(l.head, ShouldEqual, nodes[0])
				So(l.tail, ShouldEqual, nodes[2])
				So(l.count, ShouldEqual, 2)
			})

			Convey("When all nodes are removed", func() {
				for i := 0; i < len(nodes); i++ {
					err := l.Remove(nodes[1])
					So(err, ShouldBeNil)
				}

				So(l.head, ShouldBeNil)
				So(l.tail, ShouldBeNil)
				So(l.count, ShouldEqual, 0)
			})

			Convey("When user attempts to remove a nil node", func() {
				err := l.Remove(nil)
				So(err, ShouldBeError, errItemNil)
			})
		})

		Convey("Prepend tests", func() {
			l := newDoublyLinkedList()
			So(l.count, ShouldEqual, 0)

			nodes := []*node{
				{item: &foo{id: 1}},
				{item: &foo{id: 2}},
				{item: &foo{id: 3}},
			}

			// Prepending to empty list
			l.Prepend(nodes[0])
			So(l.count, ShouldEqual, 1)
			So(l.head, ShouldEqual, nodes[0])
			So(l.tail, ShouldEqual, nodes[0])

			// Prepend an additional item
			l.Prepend(nodes[1])
			So(l.count, ShouldEqual, 2)
			So(l.head, ShouldEqual, nodes[1])
			So(l.tail, ShouldEqual, nodes[0])

			// Prepend a third item
			l.Prepend(nodes[2])
			So(l.count, ShouldEqual, 3)
			So(l.head, ShouldEqual, nodes[2])
			So(l.tail, ShouldEqual, nodes[0])
		})
	})
}

func TestCacheGet(t *testing.T) {
	Convey("Getter tests", t, func() {
		Convey("Given an empty cache, then Get fails", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			_, exists := cache.Get(123)
			So(exists, ShouldBeFalse)

			_, err = NewCache(0)
			So(err, ShouldBeError, ErrInvalidSize)
		})

		Convey("Given a cache with an item, then Get succeeds", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			item := &foo{
				id: 123,
			}
			err = cache.Put(item)
			So(err, ShouldBeNil)
			found, ok := cache.Get(item.ID())
			So(ok, ShouldBeTrue)
			So(found.ID(), ShouldEqual, item.ID())

			// Adding another item suceeds as well
			item2 := &foo{
				id: 345,
			}
			err = cache.Put(item2)
			So(err, ShouldBeNil)
			found, ok = cache.Get(item2.ID())
			So(ok, ShouldBeTrue)
			So(found.ID(), ShouldEqual, item2.ID())
		})

		Convey("Given an item has been removed, then Get fails", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			item := &foo{
				id: 123,
			}
			err = cache.Put(item)
			So(err, ShouldBeNil)

			target, ok := cache.Get(item.ID())
			So(ok, ShouldBeTrue)
			So(target.ID(), ShouldEqual, item.ID())

			err = cache.Remove(item.ID())
			So(err, ShouldBeNil)

			_, ok = cache.Get(item.ID())
			So(ok, ShouldBeFalse)
		})

		Convey("Given a cache with several items, getting each one rotates it to the front of list", func() {
			numItems := 10
			cache, err := NewCache(numItems)
			So(err, ShouldBeNil)

			// Add a bunch of items to the cache
			items := []*foo{}
			for i := 0; i < numItems; i++ {
				item := &foo{
					id: i,
				}
				items = append(items, item)

				err = cache.Put(item)
				So(err, ShouldBeNil)
			}

			for i := 0; i < numItems; i++ {
				item := items[i]
				target, ok := cache.Get(item.ID())
				So(ok, ShouldBeTrue)
				So(target.ID(), ShouldEqual, item.ID())
				// The fetched item should now be at front of the list.
				So(cache.itemList.head.item.ID(), ShouldEqual, item.ID())
			}
		})
	})
}

func TestCacheRemove(t *testing.T) {
	Convey("Removal tests", t, func() {
		Convey("Given an empty cache, then Remove fails", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)

			item := &foo{
				id: 123,
			}
			err = cache.Remove(item.ID())
			So(err, ShouldEqual, ErrItemNotFound)
		})

		Convey("Given a non-empty cache, then Remove succeeds", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)

			item := &foo{
				id: 123,
			}
			err = cache.Put(item)
			So(err, ShouldBeNil)

			err = cache.Remove(item.ID())
			So(err, ShouldBeNil)

			// Calling Remove again should fail
			err = cache.Remove(item.ID())
			So(err, ShouldEqual, ErrItemNotFound)
		})
	})
}

func TestCacheAdd(t *testing.T) {
	Convey("Add tests", t, func() {
		Convey("Given an empty cache, Add succeeds", func() {
			cache, err := NewCache(10)
			So(err, ShouldBeNil)
			err = cache.Put(&foo{
				id: 123,
			})
			So(err, ShouldBeNil)
		})

		Convey("Given a duplicate item is added, Add returns error", func() {
			cache, err := NewCache(10)
			So(err, ShouldBeNil)
			item := &foo{
				id: 123,
			}
			err = cache.Put(item)
			So(err, ShouldBeNil)

			err = cache.Put(item)
			So(err, ShouldEqual, ErrDuplicateItem)
		})
		Convey("Given a cache of size one, multiple Add calls succeed with evictions", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			item := &foo{
				id: 234,
			}
			err = cache.Put(item)
			So(err, ShouldBeNil)

			item2 := &foo{
				id: 123,
			}

			err = cache.Put(item2)
			So(err, ShouldBeNil)

			item3 := &foo{
				id: 456,
			}
			err = cache.Put(item3)
			So(err, ShouldBeNil)
		})
	})
}
