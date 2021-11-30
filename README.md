# Armiarma. A distributed network monitoring tool

## Motivation
Distributed p2p networks are gaining popularity with the popularization of blockchain applications. In a case scenario where the critical message exchange entirely relies on the p2p network underneath, a deep and complete analysis of the real-time network status can directly prevent or spot vulnerabilities on the application protocol.

With this idea in mind, from Miga Labs, we want to provide a tool able to join p2p networks (check the protocol compatibility list), share, and adopt the using protocols to provide the datasets and the tools to study the new generation blockchain networks.

## Who are we?
[Miga Labs](http://migalabs.es/) is a young department of the Barcelona Supercomputing Center (BSC). Miga Labs is a group specialized in next-generation Blockchain technology, focusing on Sharding and Proof-of-Stake protocols.


## Getting started
The project offers a network crawler able to monitor the Eth2 p2p network. 

### Prerequisites
To use the tool, the following requirements need to be installed in the machine:
- Go on its 1.17 version or above. Go needs to be executable from the terminal. Lower versions will report a dependency import error for the package `io/fs`.

Alternatively, the tool can also be executed from:
- Docker
- Docker-compose

OPTIONAL for data visualization:
- Prometheus time-series database.
- Grafana dashboard.


###  Binary compilation from source-code 
To run the tool from the source code, follow these steps:

```
# Donwload the git repository of the tool
git clone https://github.com/migalabs/armiarma.git

# Switch from 'master' branch to 'integral-refactor' (newest stable version)
git checkout integral-refactor

# Compile the tool generating the armiarma binary
go build -o armiarma

# Ready to call the tool
./armiarma [options]

```

### Execution
At the moment, the tool only offers a single command for the crawler. Check the description bellow.
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
# Call docker-compose in the root of the repository and that's all
docker-compose up 
```
Docker-compose will generate the Docker image for you and will run the crawler in your machine. 
Please, note that by running the tool through the `docker-compose` command, the default config-file will serve as reference `config-files/config.json` for the tool configuration. The resulting `metrics.csv` and `peerstore.db` will be taken/generated from the folder `./peerstore`.

Remember that all these default configurations could be modified from the `docker-compose.yaml` file. 

### Supported networks
Currently supported protocols:
```
Ethereum 2      Different networks or forks can be crawled by defining the 'ForkDigest' in the 'config.json' file    
```

## Data visualization
The combination of Prometheus and Grafana is the one that we have chosen to display the network data. In the repository, both configuration files are provided. In addition, the crawler, by default, exports all the metrics to Prometheus in port 9080. 

The results of our analysis are also openly available on our website [migalabs.es](https://migalabs.es).

## Contact
To get in contact with us, feel free to reach us through our [email](migalabs@protonmail.com), and don't forget to follow our latest news on [Twitter](https://twitter.com/miga_labs). 

## Notes
Please, note that the tool is currently in a developing stage. Any bugs reports and/or suggestions are very welcome.



## License
MIT, see [LICENSE](https://github.com/Cortze/armiarma/blob/master/LICENSE) file.
