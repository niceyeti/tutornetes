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
	* focus first on high level api details
	* strongly focus on data objects (the node interface problem)
	* don't write generics, don't bother. Write simple code first, genericize later
	* pivot from test back to code rapidly
	* don't worry terribly about superfluous code paths, mark as TODO
*/

func TestCacheGet(t *testing.T) {
	Convey("Getter tests", t, func() {
		Convey("Given an empty cache, then Get fails", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			_, exists := cache.Get(123)
			So(exists, ShouldBeFalse)
		})

		Convey("Given a cache with an item, then Get succeeds", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			item := &foo{
				id: 123,
			}
			err = cache.Add(item)
			So(err, ShouldBeNil)
			found, ok := cache.Get(item.ID())
			So(ok, ShouldBeTrue)
			So(found.ID(), ShouldEqual, item.ID())

			// Adding another item suceeds as well
			item2 := &foo{
				id: 345,
			}
			err = cache.Add(item2)
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
			err = cache.Add(item)
			So(err, ShouldBeNil)

			target, ok := cache.Get(item.ID())
			So(ok, ShouldBeTrue)
			So(target.ID(), ShouldEqual, item.ID())

			err = cache.Remove(item.ID())
			So(err, ShouldBeNil)

			_, ok = cache.Get(item.ID())
			So(ok, ShouldBeFalse)
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
			err = cache.Add(item)
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
			err = cache.Add(&foo{
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
			err = cache.Add(item)
			So(err, ShouldBeNil)

			err = cache.Add(item)
			So(err, ShouldEqual, ErrDuplicateItem)
		})
		Convey("Given a cache of size one, multiple Add calls succeed with evictions", func() {
			cache, err := NewCache(1)
			So(err, ShouldBeNil)
			item := &foo{
				id: 234,
			}
			err = cache.Add(item)
			So(err, ShouldBeNil)

			item2 := &foo{
				id: 123,
			}

			err = cache.Add(item2)
			So(err, ShouldBeNil)

			item3 := &foo{
				id: 456,
			}
			err = cache.Add(item3)
			So(err, ShouldBeNil)
		})
	})
}
