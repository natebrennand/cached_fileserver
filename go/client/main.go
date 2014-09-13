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

// cliArgs checks parses in the CLI arguments and validates them
func cliArgs() (string, int, string) {
	if len(os.Args) != 4 {
		log.Println("usuage: ./tcp_client <ip> <port> <filename>")
		os.Exit(1)
	}

	// validate port
	port, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		log.Println("usuage: ./tcp_client <ip> <port> <filename>")
		log.Println("the port provided must be an integer")
		os.Exit(1)
	}

	return os.Args[1], int(port), os.Args[3]
}

func queryServer(address, filename string) error {
	// open the tcp connection
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to open TCP connection => %s", err.Error())
	}
	defer conn.Close()
	fmt.Fprintf(conn, filename) // sends filename to server
	fmt.Fprintf(conn, "\n")

	filename = path.Base(filename)

	// open a new file to write the data to
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
	for { // read the whole file in
		if err != nil || n == 0 {
			break
		}
		n, err = io.Copy(dlFile, conn)
	}

	log.Printf("%s saved", filename)
	return nil
}

func main() {
	ip, port, filename := cliArgs()
	if err := queryServer(fmt.Sprintf("%s:%d", ip, port), filename); err != nil {
		log.Fatalf("Failed to query server properly => %s", err.Error())
	}
}
