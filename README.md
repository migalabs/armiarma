# Armiarma
![example workflow](https://github.com/migalabs/armiarma/actions/workflows/tests.yaml/badge.svg)

Armiarma is an Eth2 network Analyzer based on the [Rumor](https://github.com/protolambda/rumor) client debugger. The current version of Armiarma is based on the latest update on Rumor on its [commit](https://github.com/protolambda/rumor/commit/d42e0da5729ca887e26f43e8cf4f290a61dbdc26).

## Requisites
For using the tool, the following requirements need to be installed on the machine:
- Go on its 1.17 version or above. Go needs to be executable from the terminal. Lower versions will report a dependency import of the package `io/fs`. 
- Python3 and pip3 need to be installed and executable from the shell.
- The virtualenv tool needs to be installed for the metrics analyzer.

## Usage

### Crawler
The crawler can be easily executed from the `armiarma.sh` file (make sure that it is an executable file).
The executable `armiarma.sh` launching shell script supports three different commands that correspond to the main functionalities of the crawler tool.
Those commands are:

- `./armiarma.sh -h` to display the help text of the tool.
- `./armiarma.sh -c [network] [project-name]` to launch the armiarma crawler on the given network. The given project name is the folder's name where the gathered metrics and all the related information will be saved. The folder will be placed on the `[armiarma-root]/examples` folder.
- `./armiarma.sh -f [network] [project-name] [time]` to launch the armiarma crawler on the given network for the given `time` in minutes. The given project name is the folder's name where the gathered metrics and all the related information will be saved. The folder will be placed on the `./examples` folder.

```
    ./armiarma.sh --> -h
                  '-> -c [network] [project-name]
                  '-> -f [network] [project-name] [time](minutes)
```

Currently supported networks:
    - Mainnet

The tool will get the necessary Go packages and compile Rumor for the user.
The tool also exports by default some metrics to a Prometheus endpoint(Port 9080).

### Analyzer
The Analyzer part of the tool can be found at `./analyzer`.

We recommend installing the necessary dependencies on a `virtualenv` placed inside `./analyzer`.  
To generate it:
1. Run:
```
cd analyzer
python3 -m virtualenv venv (or equivalent)
``` 
2. Activate the `venv` running:
```
source venv/bin/activate
```
3. Install the necessary Python dependencies in the `venv` by running:
```
pip3 install -r requirements.txt
```
4. After the virtualenv has been activated, to perform the analysis of a given `[project-name]` run:
```
python3 armiarma-analyzer.py [project-name]
```
Remember that the projects are stored in `[armiarma-root-folder]/examples`, and that the output `metrics.csv` and `metrics-summary.pdf` will be available in `[armiarma-root-folder]/examples/[project-name]`.


## NOTES
Please, note that the tool is currently in a developing stage. Any bugs reports and/or suggestions are very welcome.

## LICENSE

MIT, see [LICENSE](https://github.com/migalabs/armiarma/blob/master/LICENSE) file.