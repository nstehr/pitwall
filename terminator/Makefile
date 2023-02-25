.PHONY:	proto
proto:
	protoc --proto_path=../orchestrator/proto --go_out=vm --go-grpc_out=vm --go_opt=paths=source_relative ../orchestrator/proto/vm.proto
clean:
	rm -rf dist/

dist:
	env GOOS=linux GOARCH=amd64 go build -o dist/terminator