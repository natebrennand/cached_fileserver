package main

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	maxCacheSize = 64000000
)

type cacheFile struct {
	Name string
	Size int
}

// LRUCache represents a LRU file cache with a max size of 64 MB
type LRUCache struct {
	dir      string                  // directory to look for files
	contents *list.List              // ordered list of files in cache
	size     int                     // current size of the contents of the cache
	data     map[string]bytes.Buffer // actual data in the cache, key refers to the filename
}

// NewLRUCache instantiates a new LRU cache
func NewLRUCache(dir string) *LRUCache {
	return &LRUCache{
		dir:      dir,
		contents: list.New(),
		data:     make(map[string]bytes.Buffer),
	}
}

// pops the last (last used) element in the list and updates the cache size
func (c *LRUCache) pop() error {
	tail := c.contents.Back()
	if tail == nil {
		return errors.New("empty list")
	}

	evicted := c.contents.Remove(tail).(cacheFile)
	c.size -= evicted.Size
	delete(c.data, evicted.Name)
	log.Printf("%s evicted from cache", evicted.Name)
	return nil
}

// promote moves the specified element to the front of the list
func (c *LRUCache) promote(name string) error {
	for e := c.contents.Front(); e != nil; e = e.Next() {
		if e.Value.(cacheFile).Name == name {
			c.contents.MoveToFront(e)
			return nil
		}
	}
	return errors.New("element not found")
}

// set attempts to add a file to the cache
func (c *LRUCache) set(name string, data *bytes.Buffer) error {
	// don't attempt to cache if too large
	if data.Len() > maxCacheSize {
		return nil
	}

	// pop items until enough space
	for data.Len()+c.size > maxCacheSize {
		if err := c.pop(); err != nil {
			log.Printf("Rejected file %s from entering the cache due to size limitations", name)
		}
	}

	c.data[name] = *data
	c.size += data.Len()
	c.contents.PushFront(cacheFile{
		Name: name,
		Size: data.Len(),
	})
	return nil
}

// Get returns a byte array of the requested file.
// The cache will be updated in the background.
func (c *LRUCache) Get(name string) ([]byte, error) {
	data, exists := c.data[name]
	if exists {
		c.promote(name)
		log.Printf("Cache hit. File %s sent to the client", name)
		return data.Bytes(), nil
	}

	buf, err := getFile(c.dir, name)
	if err != nil {
		return []byte{}, err
	}

	log.Printf("Cache miss. File %s sent to the client", name)
	return buf.Bytes(), c.set(name, &buf)
}

// getFile looks for the file based on the name and loads it into a buffer which is returned
// An error is returned in the case of an empty file because we do not plan to cache it
func getFile(dir, name string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	// combine the directory and filename when reading
	fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, name))
	if err != nil {
		log.Printf("File %s does not exist", err.Error())
		return buf, fmt.Errorf("Cache error => %s", err.Error())
	}

	n, err := buf.Write(fileBytes)
	if n == 0 {
		return buf, errors.New("No data found in file")
	} else if err != nil {
		return buf, fmt.Errorf("Failure to read in file data => %s", err.Error())
	}

	return buf, nil
}
