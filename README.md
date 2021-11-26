# Armiarma. A distributed network monitoring tool

## Motivation
Distributed p2p networks are gaining popularity with the popularization of blockchain applications. On a case scenario, where the critical message exchange complitely relies on the p2p network underneath, a deep and complete analysis on the real-time network status can directly prevent or spot vulnerabilities on the application protocol.

With this idea in mind, from Miga Labs, we want to provide a tool able to join p2p networks (check the protocol compatibility list), share and adopt the using protocols so that we can provide the datasets and the tools to study the new generation blockchain networks.

## Who are we?
[Miga Labs](http://migalabs.es/) is a young department of the Barcelona Supercomputing Center (BSC). Miga Labs is a group specialized in next-generation Blockchain technology, with a focus on Sharding and Proof-of-Stake protocols.

## Binary insstallation

### Requisites
For using the tool, the following requirements need to be installed on the machine:
- Go on its 1.17 version or above. Go needs to be executable from the terminal. Lower versions will report a dependency import error for the package `io/fs`.


### Steps
In order to execute it, download the repository and build the project
```
# Donwload the git repository of the tool
git clone git@github.com:Cortze/armiarma.git

# Switch from master branch to integral-refactor (newst stable version)
git checkout integral-refactor

# Compile the tool generating the armiarma binary
go build -o armiarma

# Ready to use calling
./armiarma

```

### Execution
In order to execute it, download the repository and build the project
```

EXECUTION:
    ./armiarma [command] [options]

COMMANDS
    crawler     Crawls around the network looking for peers and their status

OPTIONS:
    --config-file   Load the configuration from the file into the executable.
                    Find a config.json example in ./config-files/config.json

```
## Docker installation
We provide a Dockerfile that can be used to 

## The project

Our main goal is to build a crawler using the standard libp2p and ethereum go libraries.
By using the standard libraries we aim to build and know our own project, providing the abilty to fully customize the crawler to adjuts it to specific cases.
Having the possibility to fully customize the parameters of the crawler, we will be able to test specific and rare scenarios.

## Structure

Our module is divided into several packages, each of them having a unique goal.

#### Config
The package is intended to read information from the configuration file (provided through command line) and expose the information so that it can be read into the module

#### Info
The package is intended to read information from the configuration file, making use of the Config package, and importing into an object which will then be used by other packages to access the information. This package also validates and formats the information to be easily read by other components of this project.

#### Host
The package is intended to create and start a host using the LibP2P standard library. This host will then be used, along with the node, to interact with the Ethereum 2.0 network.

#### Enode
The package is intended to create a node compatible with the Ethereum network.

#### Discovery
The package is intended to configure and start the discovery5 service in order to find other nodes in the network.
This package will have all needed methods to interact with the discovery5 library for our specific scenarios.


![alt text](https://github.com/Cortze/armiarma/blob/integral-refactor/doc/modules/Armiarma_packages.drawio.png)

## NOTES
Please, note that the tool is currently on a developing stage, any bugs reports or suggestions will be accepted.

## LICENSE
MIT, see [LICENSE](https://github.com/Cortze/armiarma/blob/master/LICENSE) file.
