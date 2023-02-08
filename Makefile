GOCC=go1.17.7
MKDIR_P=mkdir -p
GIT_SUBM=git submodule

BIN_PATH=./build
BIN="./build/armiarma"

.PHONY: build dependecies install clean

build:
	$(GOCC) get
	$(GOCC) build -o $(BIN)

dependecies:
	$(GIT_SUBM) update --init
	cd go-libp2p-pubsub && git checkout origin/armiarma && git pull origin armiarma

install:
	$(GOCC) install
	
clean:
	rm -r $(BIN_PATH)

