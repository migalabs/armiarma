# Armiarma. A distributed network monitoring tool

## Motivation
Distributed p2p networks are gaining popularity with the popularization of blockchain applications. In a scenario where the critical message exchange entirely relies on the p2p network underneath, a deep and complete analysis of the real-time network status can directly prevent or spot vulnerabilities on the application protocol.

With this idea in mind, from Miga Labs, we want to provide a tool able to join p2p networks (check the protocol compatibility list), share, and adopt the used protocols to provide the datasets and the tools to study the new blockchain networks generation.

## Who are we?
[MigaLabs](https://migalabs.io/) is a research group specialized in next-generation Blockchain technology, focusing on in-depth studies and solutions for Blockchain Scalability, Security, and Sustainability. Our team comprises technical experts in the Ethereum protocol, offering consulting services covering all technical aspects of Blockchain technology.

## Getting started
The project offers a network crawler able to monitor the Eth2 p2p network. 

### Prerequisites
To use the tool, the following requirements need to be installed in the machine:
- git
- gcc - C compiler
- [go](https://go.dev/doc/install) on its 1.17 version (upper versions might fail to compile, we are working on it). Go needs to be executable from the terminal. Lower versions will report a dependency import error for the package `io/fs`.
- PostgreSQL DB 

Alternatively, the tool can also be executed from:
- [docker](https://docs.docker.com/get-docker/) / [docker-compose](https://docs.docker.com/compose/install/)

OPTIONAL for data visualization:
- [prometheus](https://prometheus.io/docs/prometheus/latest/installation/) time-series database.
- [grafana](https://grafana.com/grafana/download) dashboard.


###  Binary compilation from source-code 
To run the tool from the source code, follow these steps:

```
# Donwload the git repository of the tool
git clone https://github.com/migalabs/armiarma.git && cd armiarma

# Compile the tool generating the armiarma binary
make dependencies
make build

# Ready to call the tool
./build/armiarma [options] [FLAGS]

```

### Execution
At the moment, the tool only offers a single command for the crawler. Check the description below.
```

EXECUTION:
    ./build/armiarma [OPTIONS] [FLAGS]

OPTIONS:
    eth2     crawl the given Ethereum CL network (selected by fork_digest)
    help, h  Shows a list of commands or help for one command
```
## Docker installation
We also provide a Dockerfile and Docker-Compose file that can be used to run the crawler without having to compile it manually. The docker-compose file spaws the following docker images:
- `Armiarma` instance with the configuration file provided by arguments
- `PostgreSQL` instance to store crawling data
- `Prometheus` instance to read the metrics exported by `Armiarma`
- `Grafana` instance as a dashboard to monitor the crawl (Eth2 dashboard)

List of ports that are going to be used:

| Instance | Port | Credentials | 
| -------- | ---- | --------- |
| Armiarma | `9020` & `9080` | - |
| PostgreSQL | `5432` | user=`user` & password=`password` | 
| Prometheus | `9090` | - | 
| Grafana | `3000` | user=`admin` & password=`admin` | 


To spawn up the entire set-up, make a copy of the `.env_template` to configure it at your wish, and just run the following command in the root directory

```
# Call docker-compose in the root of the repository, and that's all
docker-compose up --env-file <.env-copy-path> 
```
Docker-compose will generate the Docker images for you and will run the crawler and its requirements in your machine. 
Please note that, by running the tool through the `docker-compose --env-file <.env-file-name> up` command.

Remember that all these default configurations could be modified from the `docker-compose.yaml` file and from the `.env` file.

Feel free to copy the `.env_template` with `cp .env_template .my_env` and play around with the configuration of the crawler.

NOTEs: 
- you might need to run `docker-compose up` with `sudo` privileges if the Linux user doesn't belong to the docker group. 
- If the DB and Prometheus containers fail on start because of a permisions problems, make sure you grant the correct permises to the content of the folder `./app-data` doing:
```
sudo chmod 777 ./app-data/*_db
```

### Supported networks
Currently supported protocols:
```
Ethereum CL      Different networks or forks can be crawled by defining the 'ForkDigest' in the --fork-digest flag  
```

[List](./pkg/networks/ethereum/network_info.go) of fork digests.


### Custom configuration of the tool
The crawler has several fields that can be customized anytime before the launch of the crawler. The fields correspond to the following flags:

```
USAGE:
   ./build/armiarma eth2 [options...]

OPTIONS:
   --log-level value           Verbosity level for the Crawler's logs (default: info) [$ARMIARMA_LOG_LEVEL]
   --priv-key value            String representation of the PrivateKey to be used by the crawler [$ARMIARMA_PRIV_KEY]
   --ip value                  IP in the machine that we want to asign to the crawler (default: 0.0.0.0) [$ARMIARMA_IP]
   --port value                TCP and UDP port that the crawler with advertise to establish connections (default: 9020) [$ARMIARMA_PORT]
   --metrics-ip value          IP in the machine that will expose the metrics of the crawler (default: localhost) [$ARMIARMA_METRICS_IP]
   --metrics-port value        Port that the crawler with to expose pprof and prometheus metrics (default: 9080) [$ARMIARMA_METRICS_PORT]
   --user-agent value          Agent name that will identify the crawler in the network (default: Armiarma Crawler) [$ARMIARMA_USER_AGENT]
   --psql-endpoint value       PSQL enpoint where the crwaler will submit the all the gathered info (default: postgres://user:password@localhost:5432/armiarmadb) [$ARMIARMA_PSQL]
   --peers-backup value        Time interval that will be use to backup the peer_ids into a single table - allowing to recontruct the network in past-crawled times (default: 12h) [$ARMIARMA_BACKUP_INTERVAL]
   --remote-cl-endpoint value  Remote Ethereum Consensus Layer Client to request metadata (experimental) [$ARMIARMA_REMOTE_CL_ENDPOINT]
   --fork-digest value         Fork Digest of the Ethereum Consensus Layer network that we want to crawl (default: 0x4a26c58b) [$ARMIARMA_FORK_DIGEST]
   --bootnode value            List of boondes that the crawler will use to discover more peers in the network (One --bootnode <bootnode> per bootnode) [$ARMIARMA_BOOTNODES]
   --gossip-topic value        List of gossipsub topics that the crawler will subscribe to [$ARMIARMA_GOSSIP_TOPICS]
   --subnet value              List of subnets (gossipsub topics) that we want to subscribe the crawler to (One --subnet <subnet_id> per subnet) [$ARMIARMA_SUBNETS]
   --persist-msgs              Decide whether we want to track the msgs-metadata into the DB (default: false) [$ARMIARMA_PERSIST_MSGS]
   --val-pubkeys value         Path of the file that has the pubkeys of those validators that we want to track (experimental) [$ARMIARMA_VAL_PUBKEYS]
   --help, -h                  show help (default: false)

```

## Data visualization
The combination of Prometheus and Grafana is the one that we have chosen to display the network data. In the repository, both configuration files are provided. In addition, the crawler, by default, exports all the metrics to Prometheus in port 9080. 

The results of our analysis are also openly available on our website [migalabs.es](https://migalabs.es/beaconnodes).

## Contact
To get in contact with us, feel free to reach us through our [email](migalabs@protonmail.com), and don't forget to follow our latest news on [Twitter](https://twitter.com/miga_labs). 

## Notes
Please, note that the tool is currently in a developing stage. Any bugs report and/or suggestions are very welcome.

## Maintainer
@cortze

## License
MIT, see [LICENSE](https://github.com/Cortze/armiarma/blob/master/LICENSE) file.
