package cache

import (
	"bytes"
	"testing"
)

func makeTestCache() *LRUCache {
	c := NewLRUCache()
	c.contents.PushFront(cacheFile{
		Name: "test1",
		Size: 10,
	})
	c.contents.PushFront(cacheFile{
		Name: "test2",
		Size: 20,
	})
	c.contents.PushFront(cacheFile{
		Name: "test3",
		Size: 30,
	})
	c.size = 60
	return c
}

func TestPop(t *testing.T) {
	x := makeTestCache()
	if err := x.pop(); err != nil {
		t.Errorf("nil should be returned from pop(), recieved %s", err.Error())
	} else if x.size != 50 {
		t.Error("size of cache should be 50")
	}

	x.pop() // removing test2
	x.pop() // removing test3

	if nil == x.pop() {
		t.Error("error should be returned from empty cache")
	} else if x.size != 0 {
		t.Error("size of cache should be 0 when empty")
	}
}

func TestPromote(t *testing.T) {
	x := makeTestCache()
	x.promote("test1")

	if err := x.pop(); err != nil {
		t.Errorf("nil should be returned from pop(), recieved %s", err.Error())
	} else if x.size != 40 {
		t.Error("size of cache should be 40 since test2 was removed")
	}

	if x.promote("testX") == nil {
		t.Error("non-existent file should return error")
	}
}

func TestSet(t *testing.T) {
	x := makeTestCache()
	var buf bytes.Buffer
	buf.Write([]byte("0123456789"))

	x.set("test4", &buf)
	if x.size != 60+10 {
		t.Errorf("Size not properly updated on set(), expected 70, found %d", x.size)
	} else if _, exists := x.data["test4"]; !exists {
		t.Error("new element should be added to data map")
	}

	// cache should reach full state
	c := makeTestCache()
	c.size = maxCacheSize - 10
	c.set("test4", &buf)
	if c.contents.Len() != 4 {
		t.Error("No cache elements should've been purged")
	}
	// cache should purge one file to make space
	c.set("test5", &buf)
	if c.contents.Len() == 5 {
		t.Error("There should not be space in the cache to insert w/o purging\nCache contents:")
		for e := c.contents.Front(); e != nil; e = e.Next() {
			t.Errorf("%#v\n", e.Value.(cacheFile))
		}
		t.Errorf("%#v\n", c)
	}
}

func TestGet(t *testing.T) {
	x := makeTestCache()
	var buf bytes.Buffer
	buf.Write([]byte("0123456789"))

	x.set("test4", &buf)
	if x.size != 70 {
		t.Errorf("Size not properly updated on set(), expected 70, found %d", x.size)
	} else if _, exists := x.data["test4"]; !exists {
		t.Error("new element should be added to data map")
	}

	x.set("test5", &buf)
	if x.size != 80 {
		t.Errorf("Size not properly updated on set(), expected 70, found %d", x.size)
	} else if _, exists := x.data["test5"]; !exists {
		t.Error("new element should be added to data map")
	}

	// test that proper data is returned
	data, err := x.Get("test4")
	if err != nil {
		t.Fatal("failed to retrieve data from cache")
	} else if string(data) != buf.String() {
		t.Fatal("failed to retrieve correct data from cache")
	}

	// make sure that element was promoted in cache
	if x.contents.Front().Value.(cacheFile).Name != "test4" {
		t.Error("failed to promote queried file")
	}
}

func TestGetFile(t *testing.T) {
	x := makeTestCache()
	data, err := x.Get("test.txt")
	testData := "123456789\n"

	if err != nil {
		t.Errorf("test file should be found and returned w/o error, err => %s", err.Error())
	}
	if string(data) != testData {
		t.Errorf("test file should be properly read in, found => %s, expected => %s", string(data), testData)
	}

	// test non-existant file
	if _, err = x.Get("non_existent.txt"); err == nil {
		t.Error("error should be thrown on non-existent file")
	}
}
