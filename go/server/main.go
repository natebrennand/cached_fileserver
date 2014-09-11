package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	maxFilenameSize = 1000
)

// cliArgs checks parses in the CLI arguments and validates them
func cliArgs() (int, string) {
	if len(os.Args) != 3 {
		log.Println("usuage: ./tcp_server <port> <path to serving directory>")
		os.Exit(1)
	}

	port, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Println("usuage: ./tcp_server <port> <path to serving directory>")
		log.Println("the port provided must be an integer")
		os.Exit(1)
	}

	fi, err := os.Stat(os.Args[2])
	if err != nil {
		log.Printf("Error examining directory => %s", err.Error())
		log.Println("usuage: ./tcp_server <port> <path to serving directory>")
		os.Exit(1)
	} else if !fi.IsDir() {
		log.Printf("path provided is not a directory")
		log.Println("usuage: ./tcp_server <port> <path to serving directory>")
		os.Exit(1)
	}

	// NOTE: we use Name() to standardize to a dir name w/o a trailing slash
	return int(port), fi.Name()
}

// HandleFileRequest parses the file request then queries the cache for the requested file
func HandleFileRequest(conn net.Conn, cache *LRUCache) {
	defer conn.Close()
	buf := make([]byte, maxFilenameSize)

	// read in the request
	nRead, err := conn.Read(buf)
	filename := string(buf)
	if err != nil || nRead == 0 {
		log.Printf("Error reading in TCP request from %s => %s", conn.RemoteAddr().String(), err.Error())
	}

	log.Printf("Client %s is requesting file %s", conn.RemoteAddr().String(), filename)
	if err := cache.WriteToConn(conn, filename); err != nil {
		log.Printf("Failed to write to client connection => %s", err.Error())
	}
}

func main() {
	// set up arguments and cache
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
		go HandleFileRequest(conn, cache)
	}
}
