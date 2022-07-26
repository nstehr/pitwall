.PHONY:	proto
proto:
	protoc --proto_path=proto --go_out=vm --go-grpc_out=vm --go_opt=paths=source_relative proto/vm.proto
	protoc --proto_path=proto --go_out=orchestrator --go-grpc_out=orchestrator --go_opt=paths=source_relative proto/orchestrator.proto

clean:
	rm -rf dist/

dist:
	env GOOS=linux GOARCH=amd64 go build -o dist/pitwall