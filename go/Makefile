
all: clean server client

server:
	cd server && \
	go build -o tcp_server && \
	mv tcp_server ..

client:
	cd client && \
	go build -o tcp_client && \
	mv tcp_client ..

clean:
	go clean


.PHONY: all server client clean
