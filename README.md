# armiarma
Armiarma is a Eth2 network Analyzer based on the [Rumor](https://github.com/protolambda/rumor) client debugger. The current version of Armiarma is based on the latest update on Rumor on its [commit](https://github.com/protolambda/rumor/commit/d42e0da5729ca887e26f43e8cf4f290a61dbdc26).

### Requisites
To use the tool the following tools needs to be installed on the machine:
- Go on its 1.15 version or above. Go needs to be executable from the terminal and despite previous versions might work, we recomend ussing the 1.15 for a better preformance of the crawler. The current [dv5.1](https://github.com/ethereum/devp2p/blob/master/discv5/discv5.md) version will not work with lower versions than the 1.15.
- Python3 needs to be installed and executable from the shell.
- The viertualenv tool needs to be installed for the metrics analyzer. 

### Usage
The crawler can be easily executed from the `armiarma.sh` file (make sure that is an executable file). 
The executable `armiarma.sh` launching shell script supports three different commands that correspond to the main functionalities of the tool.
Those commands are:

- `./armiarma.sh -h` to display the help text of the tool. 
- `./armiarma.sh -c [network] [project-name]` to launch the armiarma crawler on the given network. The given project-name is the name of the folder where the gathered metrics and where all the related information will be saved. The folder will be placed on the `./examples` folder. 
- `./armiarma.sh -p [project-name]` to launch the armiarma analyzer over the metrics from the given project. The generted plots of the analysis will be saved on the `./examples/[project-name]/plots` folder.

```
    ./armiarma.sh --> -h
                  '-> -c [network] [project-name]
                  '-> -p [project-name]
```

The tool will get the Go packages and compile Rumor as generate the virtual env and install the python3 dependencies for the user.  

Currently supported networks:
    - Mainnet

### NOTES
Please, note that the tool is currently on a developing stage, any bugs reports or suggestions will be accepted.

### LICENSE

MIT, see [LICENSE](https://github.com/Cortze/armiarma/LICENSE) file.
