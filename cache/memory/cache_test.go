package memory

import (
	"testing"

	"github.com/rdumont/loki"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCache(t *testing.T) {
	Convey("Set", t, func() {
		Convey("just calls inner cache", func() {
			inner := new(fakeCache)
			c := &cache{Cache: inner}
			key := "http://my-host.com/api/v2/documents/search?foo-bar"

			c.Set(key, nil)

			So(inner.SetCalls.GetCall()[0], ShouldEqual, key)
			So(c.rootKeys, ShouldBeNil)
		})

		Convey("stores key if it refers to a root API call", func() {
			inner := new(fakeCache)
			c := &cache{Cache: inner}
			key := "http://my-host.com/api/v2?proxy-host=foobar"

			c.Set(key, nil)

			So(inner.SetCalls.GetCall()[0], ShouldEqual, key)
			So(c.rootKeys, ShouldResemble, []string{key})
		})
	})

	Convey("DeleteAllRootKeys", t, func() {
		Convey("deletes all recorded root keys", func() {
			inner := new(fakeCache)
			keys := []string{"one", "two", "three"}
			c := &cache{Cache: inner, rootKeys: keys}

			c.DeleteAllRootKeys()

			So(inner.DeleteCalls.CallCount(), ShouldEqual, len(keys))
			for i, k := range keys {
				So(inner.DeleteCalls.GetNthCall(i)[0], ShouldEqual, k)
			}
		})
	})
}

type fakeCache struct {
	SetCalls    loki.Method
	DeleteCalls loki.Method
}

func (c *fakeCache) Get(key string) (responseBytes []byte, ok bool) {
	panic("not implemented")
}

func (c *fakeCache) Set(key string, responseBytes []byte) {
	c.SetCalls.Receive(key, responseBytes)
}

func (c *fakeCache) Delete(key string) {
	c.DeleteCalls.Receive(key)
}
