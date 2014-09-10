package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
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
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to open TCP connection => %s", err.Error())
	}

	fmt.Fprintf(conn, filename)

	var buf bytes.Buffer
	n, err := buf.ReadFrom(conn)
	if err != nil {
		log.Fatal("Error reading from TCP connection => %s", err.Error())
	} else if n == 0 { // warn if file is empty
		log.Println("File %s does not exist on the server")
		conn.Close()
		return nil
	}
	if err := conn.Close(); err != nil {
		log.Printf("Failed to close connection => %s", err.Error())
	}

	if err = ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		log.Fatalf("failed to write file => %s", err.Error())
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
