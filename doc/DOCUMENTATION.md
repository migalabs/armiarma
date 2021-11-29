# Our Project
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