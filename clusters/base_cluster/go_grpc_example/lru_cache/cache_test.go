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

func TestCache(t *testing.T) {
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
		/*
			Convey("Given a cache containing a few items, Add succeeds", func() {

			})
			Convey("Given a cache at capacity, Add succeeds", func() {

			})
		*/
	})
}
