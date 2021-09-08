# Armiarma
Armiarma is a Eth2 network Analyzer based on the [Rumor](https://github.com/protolambda/rumor) client debugger. The current version of Armiarma is based on the latest update on Rumor on its [commit](https://github.com/protolambda/rumor/commit/d42e0da5729ca887e26f43e8cf4f290a61dbdc26).

### Requisites
To use the tool the following tools needs to be installed on the machine:
- Go on its 1.15 version or above. Go needs to be executable from the terminal and despite previous versions might work, we recomend ussing the 1.15 for a better preformance of the crawler. The current [dv5.1](https://github.com/ethereum/devp2p/blob/master/discv5/discv5.md) version will not work with lower versions than the 1.15.
- Python3 and pip3 needs to be installed and executable from the shell.
- The viertualenv tool needs to be installed for the metrics analyzer.

### Usage

#### Crawler
The crawler can be easily executed from the `armiarma.sh` file (make sure that is an executable file).
The executable `armiarma.sh` launching shell script supports five different commands that correspond to the main functionalities of the crawler tool.
Those commands are:

- `./armiarma.sh -h` to display the help text of the tool.
- `./armiarma.sh -c [network] [project-name]` to launch the armiarma crawler on the given network. The given project-name is the name of the folder where the gathered metrics and where all the related information will be saved. The folder will be placed on the `./examples` folder.
- `./armiarma.sh -f [network] [project-name] [time]` to launch the armiarma crawler on the given network for the given `time` in minutes. The given project-name is the name of the folder where the gathered metrics and where all the related information will be saved. The folder will be placed on the `./examples` folder.

```
    ./armiarma.sh --> -h
                  '-> -c [network] [project-name]
                  '-> -f [network] [project-name] [time](minutes)
```

Currently supported networks:
    - Mainnet

The tool will get the necessary Go packages and compile Rumor for the user.
The tool also exports by default some metrics to a Prometheus endpoint. (Port 9080)

#### Analyzer
The Analyzer part of the tool can be found at `analyzer/`.

We recommend to install the necessary dependencies on a `virtualenv` by doing `python3 -m virtualenv venv` (or equivalent) where after activating with `source venv/bin/activate` the necessary Python dependencies can be installed by doing `pip3 install -r analyzer/requirements`.

To perform the analysis of a given `[project-name]`, it can be done calling:
```

```


### NOTES
Please, note that the tool is currently on a developing stage, any bugs reports or suggestions will be accepted.

### LICENSE

MIT, see [LICENSE](https://github.com/migalabs/armiarma/blob/master/LICENSE) file.
