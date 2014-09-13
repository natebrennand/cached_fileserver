package main

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

const (
	maxCacheSize = 64 * 1000 * 1000
	bufferSize   = 8 * 1000 * 1000
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
		log.Printf("Rejected file %s from entering the cache due to size limitations", name)
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

// WriteToConn writes the specified file to the network connection passed in
func (c *LRUCache) WriteToConn(conn io.Writer, name string) error {
	filename := path.Join(c.dir, name)
	matchParent, err1 := path.Match("../*", filename)
	matchRoot, err2 := path.Match("/*", filename)
	if matchParent || matchRoot || err1 != nil || err2 != nil {
		return errors.New("Illegal file, cannot be in parent dir")
	}

	data, exists := c.data[name]
	if exists {
		c.promote(name)
		log.Printf("Cache hit. File %s sent to the client", name)
		_, err := conn.Write(data.Bytes())
		return err
	}

	// write the file to the connection
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("%#v\n", err)
		return fmt.Errorf("Error opening file => %s", err.Error())
	}
	defer file.Close()

	for {
		n, err := io.Copy(conn, file)
		if n == 0 || err == io.EOF || err != nil {
			break
		}
	}

	// see if file should be cached
	fi, err := os.Stat(filename)
	if err != nil {
		// will not attempt to cache
		log.Printf("Error reading file stats for %s => %s", filename, err.Error())
		return nil
	} else if fi.Size() <= maxCacheSize { // if file can fit in the cache
		buf, err := getFile(filename)
		if err != nil {
			// will drop attempt to cache
			log.Printf("WARN: Error reading in file for caching => %s", err.Error())
			return nil
		}
		c.set(name, &buf)
	}

	log.Printf("Cache miss. File %s sent to the client", name)
	return nil
}

// getFile looks for the file based on the name and loads it into a buffer which is returned
// An error is returned in the case of an empty file because we do not plan to cache it
func getFile(filename string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("File '%s' does not exist", filename)
		return buf, fmt.Errorf("Cache error => %s", err.Error())
	}
	defer file.Close()

	n, err := buf.ReadFrom(file)

	if err != nil {
		return buf, fmt.Errorf("Failure to read in file data => %s", err.Error())
	} else if n == 0 {
		log.Println(n, err)
		return buf, errors.New("No data found in file")
	}

	return buf, nil
}
