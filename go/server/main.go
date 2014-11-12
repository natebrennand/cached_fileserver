package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	maxFilenameSize = 1000
)

func die(reason string) {
	log.Printf(reason)
	log.Println("usuage: ./tcp_server <port> <path to serving directory>")
	os.Exit(1)
}

// cliArgs checks parses in the CLI arguments and validates them
func cliArgs() (int, string) {
	if len(os.Args) != 3 {
		die("2 arguments required")
	}

	port, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		die("the port provided must be an integer")
	}

	fi, err := os.Stat(os.Args[2])
	if err != nil {
		die(fmt.Sprintf("Error examining directory => %s\n", err.Error()))
	} else if !fi.IsDir() {
		die("path provided is not a directory")
	}

	// NOTE: we use Name() to standardize to a dir name w/o a trailing slash
	return int(port), fi.Name()
}

// HandleFileRequest parses the file request then queries the cache for the requested file
func HandleFileRequest(conn net.Conn, cache *LRUCache) {
	defer conn.Close()

	// read in the request
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		log.Printf("Error reading in TCP request from %s => %s", conn.RemoteAddr().String(), scanner.Err().Error())
		return
	}
	filename := scanner.Text() // grab the first string from the scanner

	log.Printf("Client %s is requesting file %s", conn.RemoteAddr().String(), filename)
	if err := cache.WriteFile(conn, filename); err != nil {
		log.Printf("Failed to write to client connection => %s", err.Error())
	}
}

func main() {
	// set up arguments and instantiate cache
	port, dir := cliArgs()
	cache := NewLRUCache(dir)

	// start up TCP server
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicf("Failed to open tcp port => %s", err.Error())
	}

	// start fielding incoming requests
	log.Printf("listening on port %d", port)
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Failed to open client connection to %s => %s", conn.RemoteAddr().String(), err.Error())
		}
		go HandleFileRequest(conn, cache) // start a go-routine to handle the request
	}
}
