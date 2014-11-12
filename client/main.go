package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
)

const (
	bufferSize = 8 * 1000 * 1000
)

func die(reason string) {
	log.Printf(reason)
	log.Println("usuage: ./tcp_client <ip> <port> <filename>")
	os.Exit(1)
}

// cliArgs checks parses in the CLI arguments and validates them
func cliArgs() (string, int, string) {
	if len(os.Args) != 4 {
		die("3 arguments required")
	}

	// validate port
	port, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		die("the port provided must be an integer")
	}

	ip, filename := os.Args[1], os.Args[3]
	return ip, int(port), filename
}

func queryServer(address, filename string) error {
	// open the tcp connection
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to open TCP connection => %s", err.Error())
	}
	defer conn.Close()

	// sends filename to server
	fmt.Fprintf(conn, filename)
	fmt.Fprintf(conn, "\n")

	// open a new file to write the data to
	filename = path.Base(filename) // strip filename of directory
	dlFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Cannot open file => %s", err.Error())
	}
	defer dlFile.Close()

	// read from the connection into the buffer and write to the new file
	n, err := io.Copy(dlFile, conn)
	if err != nil || n == 0 { // check for no file
		log.Printf("File %s does not exist", filename)
		os.Remove(filename)
		return nil
	}

	// read the whole file in
	for {
		if err != nil || n == 0 {
			break
		}
		n, err = io.Copy(dlFile, conn)
	}

	log.Printf("File %s saved", filename)
	return nil
}

func main() {
	ip, port, filename := cliArgs()
	if err := queryServer(fmt.Sprintf("%s:%d", ip, port), filename); err != nil {
		log.Fatalf("Failed to query server properly => %s", err.Error())
	}
}
