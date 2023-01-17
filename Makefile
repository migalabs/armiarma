GOCC=go
MKDIR_P=mkdir -p

BIN_PATH=./build
BIN="./build/armiarma"

.PHONY: build install clean

build:
	$(GOCC) get
	$(GOCC) build -o $(BIN)


install:
	$(GOCC) install
	

clean:
	rm -r $(BIN_PATH)

