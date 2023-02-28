GOCC=go
MKDIR_P=mkdir -p
GIT_SUBM=git submodule

BIN_PATH=./build
BIN="./build/armiarma"

DOCKER_VOLUMES="./app-data/"

.PHONY: build dependencies install clean

build:
	$(GOCC) get
	$(GOCC) build -o $(BIN)

dependencies:
	$(GIT_SUBM) update --init
	cd go-libp2p-pubsub && git checkout origin/armiarma && git pull origin armiarma
	cd ../

install:
	$(GOCC) install
	
clean:
	rm -r $(BIN_PATH)
	# remove also the docker-related volumes if they exist
	sudo rm -rf $(DOCKER_VOLUMES)*_db

