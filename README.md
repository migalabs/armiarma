# Armiarma. A distributed network monitoring tool

## Motivation
Distributed p2p networks are gaining popularity with the popularization of blockchain applications. In a scenario where the critical message exchange entirely relies on the p2p network underneath, a deep and complete analysis of the real-time network status can directly prevent or spot vulnerabilities on the application protocol.

With this idea in mind, from Miga Labs, we want to provide a tool able to join p2p networks (check the protocol compatibility list), share, and adopt the used protocols to provide the datasets and the tools to study the new blockchain networks generation.

## Who are we?
[Miga Labs](http://migalabs.es/) is a young department of the Barcelona Supercomputing Center (BSC), specialized in next-generation Blockchain technology, focusing on Sharding and Proof-of-Stake protocols.


## Getting started
The project offers a network crawler able to monitor the Eth2 p2p network. 

### Prerequisites
To use the tool, the following requirements need to be installed in the machine:
- git
- gcc - C compiler
- [go](https://go.dev/doc/install) on its 1.17 version or above. Go needs to be executable from the terminal. Lower versions will report a dependency import error for the package `io/fs`.
- PostgreSQL DB 

Alternatively, the tool can also be executed from:
- [docker](https://docs.docker.com/get-docker/)
- [docker-compose](https://docs.docker.com/compose/install/)

OPTIONAL for data visualization:
- [prometheus](https://prometheus.io/docs/prometheus/latest/installation/) time-series database.
- [grafana](https://grafana.com/grafana/download) dashboard.


###  Binary compilation from source-code 
To run the tool from the source code, follow these steps:

```
# Donwload the git repository of the tool
git clone https://github.com/migalabs/armiarma.git && cd armiarma

# Compile the tool generating the armiarma binary
go build -o armiarma

# Ready to call the tool
./armiarma [options]

```

### Execution
At the moment, the tool only offers a single command for the crawler. Check the description below.
```

EXECUTION:
    ./armiarma [options]

OPTIONS:
    --config-file   Load the configuration from the file into the executable.
                    Find a config.json example in ./config-files/config.json

```
## Docker installation
We also provide a Dockerfile that can be used to run the crawler without having to compile it manually.
```
# Call docker-compose in the root of the repository, and that's all
docker-compose up 

```
Docker-compose will generate the Docker image for you and will run the crawler in your machine. 
Please note that, by running the tool through the `docker-compose` command, the default config-file will serve as reference `config-files/config.json` for the tool configuration. The resulting `metrics.csv` and `peerstore.db` will be taken/generated from the folder `./peerstore`.

Remember that all these default configurations could be modified from the `docker-compose.yaml` file. 

NOTE: you might need to run `docker-compose up` with `sudo` privileges if the Linux user doesn't belong to the docker group. 

### Supported networks
Currently supported protocols:
```
Ethereum 2      Different networks or forks can be crawled by defining the 'ForkDigest' in the 'config.json' file  
Gnosis          Gnosis fork from the Eth2 Network. Add '56fdb5e0' Gnosis ForkDigest in 'config.file' to discover and crawl the network.
```

### Custom configuration of the tool
The crawler, by default, reads the configuration file located in `config-files/config.json`. The file contains several fields that can be customized anytime before the launch of the crawler. The fields correspond to the following features:

```
IP:             IP that wants to be assigned to the crawler (default = "0.0.0.0") 
TcpPort:        Port that will be used to establish TCP connections (default = 9020)
UdpPort:        Port that will be used to establish UDP connections (default = 9020)
TopicArray:     List of GossipSub topics that the tool will be subscribed to. Leave empty [] to get default ones (Eth2 topics)
Network:        Name of the Eth2 Network that the crawler will join (default = "mainnet")
DBEndpoint:     Psql endpoint with the credentials and DB name information. (Example: 'postgresql://user:password@localhost:5432/dbname')
Eth2Endpoint:   Endpoint to an Eth2 beacon node such as Infura. Used to dynamically calculate the fork-digest of the Eth2 mainnet (default = "" since default fork-digest = Eth2 Altair)
ForkDigest:     4 byte hexadecimal code of the Network's ForkDigest (default = "afcaaba0")
UserAgent:      Name that will identify the crawler in the joined network (default = "bsc-crawler")
LogLevel:       Level of logs that will be printed in the terminal ("trace", "debug", "info", "warn", "error") (default = "info")
PrivateKey:     hexadecimal encoded libp2p privkey that will be used to create a peerID for the crawler in the network (will generate a new one by default, can be copy-pasted from the printed one in the terminal)
BootNodesFile:  List of boot-nodes that will be used for the peer discovery service (recommended = "./src/discovery/official-eth2-bootnodes.json")
OutputPath:     Output folder that will contain the CSV and DB of the crawler (default = "./peerstore")
```

## Data visualization
The combination of Prometheus and Grafana is the one that we have chosen to display the network data. In the repository, both configuration files are provided. In addition, the crawler, by default, exports all the metrics to Prometheus in port 9080. 

The results of our analysis are also openly available on our website [migalabs.es](https://migalabs.es/crawler/dashboard).

## Contact
To get in contact with us, feel free to reach us through our [email](migalabs@protonmail.com), and don't forget to follow our latest news on [Twitter](https://twitter.com/miga_labs). 

## Notes
Please, note that the tool is currently in a developing stage. Any bugs report and/or suggestions are very welcome.



## License
MIT, see [LICENSE](https://github.com/Cortze/armiarma/blob/master/LICENSE) file.
