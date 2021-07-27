# armiarma-server
TODO:

cd into proto/
If protoc is not found, see: https://stackoverflow.com/questions/57700860/protoc-gen-go-program-not-found-or-is-not-executable

Generate the protobuf files
```console
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    aggregator.proto
```

## Run the server

```
go run main.go --rpc-port=5000
```

## Run the client
Note that this is currently hardcoded to localhost
```console
./armiarma.sh -c mainnet run1
```
