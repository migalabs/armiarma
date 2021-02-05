# armiarma
Armiarma is a Eth2 network Analyzer based on the [Rumor](https://github.com/protolambda/rumor) client debugger. The current version of Armiarma is based on the latest update on [Rumor](https://github.com/protolambda/rumor/commit/d42e0da5729ca887e26f43e8cf4f290a61dbdc26).

## Requisites
To use the tool the following tools needs to be installed on the machine:
- Go on its 1.15 version or above. Go needs to be executable from the terminal and despite previous versions might work, we recomend ussing the 1.15 for a better preformance of the crawler. The current [dv5.1]() version will not work with lower versions than the 1.15.
- Python3 needs to be installed and executable from the shell
- The viertualenv tool needs to be installed for the metrics analyzer 

## Usage
The crawler can be easily executed from the `armiarma.sh` file (make sure that is an executable file). The tool has a `-h` help panel that explains the basic usage. It looks like the following:

```
  Armiarma is the ETH2 network visualizer. Please, make sure that go on its version 15 or above is installed.
  To run armiarma please follow the scheme:
      ./armiarma.sh [option] [parameters]
      options:
          -h      Print this help.
          -c      To run the crawler on the ETH2 Mainnet network.
                  parameters for -c [network] [name]
          -p      Run the analyzer part of the tool, analyze the generate metrics and generate plots with the results.
                  parameters for -p [name]

      parameters:
          [network]   The ETH2 network where the crawler will be running
                      Currently supported networks:
                          -> mainnet
          [name]      Specify the name of the folder where the metrics and plots will be stored. Find them on 'armiarma/examples/[name]'.

```

## NOTES
Please, note that the tool is currently on a developing stage, any bugs reports or suggestions will be accepted.



