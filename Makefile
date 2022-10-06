GOCC=go
MKDIR_P=mkdir -p

BIN_PATH=./build
BIN="./build/armiarma"

.PHONY: check dependencies build install clean

build: 
	$(GOCC) build -o $(BIN)


install:
	$(GOCC) install
	

clean:
	rm -r $(BIN_PATH)

