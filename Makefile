.PHONY: all protogen client server clean

all: protogen client server

protogen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative message/message.proto


client:
	go build -o dist/cli cmd/cli/cli.go

server:
	go build -o dist/server cmd/server/server.go

clean:
	rm -rf dist/client dist/server
