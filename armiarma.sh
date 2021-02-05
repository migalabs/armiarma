#!/bin/bash
# ------- ARMIARMA --------
# BSC-ETH2 TEAM
# VERSION: v0.1
version="v0.1.0"




s ---------- DEFINITION OF FUNTIONS -----------

ARMIARMA="./src/bin/armiarma"
BIN="./src/bin"

### Help function to show how to run the too
Help()
{
    # Display Help
    echo ""
    echo "  Armiarma is the ETH2 network visualizer. Please, make sure that go on its version 15 or above is installed"
    echo "  To run armiarma please follow the scheme:"
    echo "      ./armiarma.sh [option] [parameters]"
    echo "      options:"
    echo "          -h      Print this help."
    echo "          -c      To run the crawler on the ETH2 Mainnet network."
    echo "                  parameters for -c [network] [name]"
    echo "          -p      Run the analyzer part of the tool, analyze the generate metrics and generate plots with the results."
    echo "                  parameters for -p [name]"
    echo ""
    echo "      parameters:"
    echo "          [network]   The ETH2 network where the crawler will be running"
    echo "                      Currently supported networks:"
    echo "                          -> mainnet" 
    echo "          [name]      Specify the name of the folder where the metrics and plots will be stored. Find them on 'armiarma/examples/[name]'."
    echo ""
}

# Function that calls the crawler
# Arguments (1)->Path of the folder 
CheckCompileRumor(){

    if [ -f "$ARMIARMA" ]
    then 
        # Rumor has been already compiled
        echo
        echo "Rumor was already compiled"
        echo
    else
        # Rumor needs to be compiled
        echo 
        echo "Compiling Rumor ..."
        cd ./src
        # Check if the ./src/bin folder is already there
        if [[ -d "./bin" ]]; then
            echo "..."  
        else
            mkdir bin
        fi
        go build -o ./bin/armiarma
        cd "$1"
        echo "Rumor Compiled!"
        echo
    fi
}

# Generate a plain launcher.rumor on the current PATH
# Receives (1)->Example folder
TouchLauncher(){
    echo "Generating $1 Rumor Launcher"
    touch launcher.rumor
    echo "# $1 launcher" >> launcher.rumor
    echo "source config.sh" >> launcher.rumor
    echo "" >> launcher.rumor
    echo "# cd to the crawler directory" >> launcher.rumor
    echo "cd ../../src/crawler" >> launcher.rumor
    echo "# Launch the crawler" >> launcher.rumor
    echo "include crawler.rumor" >> launcher.rumor
}


# -------- END OF FUNCTION DEFINITION ---------

# Execution Secuence

# 0. Get the options

if [[ -d ./examples ]]; then
    echo "----------- ARMIARMA $version -----------"
else
    echo "----------- ARMIARMA $version -----------"
    echo ""
    echo "Generating ./examples folder"
    mkdir ./examples  
fi 



while getopts ":hcp" option; do
    case $option in
        h)  # display Help
            Help
            exit;;
        c)  # execute rumor (if its compiled)
            # Save [name]
            networkName="$2"
            folderName="$3" 
            folderPath="$PWD"
            # Flag to see if rumor needs to be installed/called
            rumorFlag="0"
            executableNetwork=""

            echo
            echo "  Crawler selected"
            echo "  network:        $networkName"
            echo "  metrics-folder: $folderName"
            echo "  folder path:    $folderPath"
            echo

            if [[ "$networkName" == "mainnet" ]]
            then
                echo "Mainnet network selected"
                # source the config file for the Eth2 Mainnet network
                source ./networks/mainnet/config.sh
                rumorFlag=$((rumorFlag+1))
                executableNetwork="mainnet-launcher.rumor"
            else
                echo "Invalid newtork."
                echo "Available networks:"
                echo "  -> mainnet"
                exit
            fi

            # Check if rumor flag has been activated to Run/compiled Rumor
            if [[ $rumorFlag -eq 1 ]]
            then
                # Rumor would need to be run/compiled
                CheckCompileRumor "$folderPath"
                
                # Switch to the crawler folder where the ".rumor" files are located
                #cd ./src/crawler
                
                # Generate the full-path for the metrics folder
                metricsFolder="${folderPath}/examples/${folderName}"
                
                # Check if the directory already exists 
                if [[ -d $metricsFolder ]]; then
                    echo
                    echo "Error. Project with name $folderName already exist"
                    echo
                    exit
                else
                    echo "Getting the env ready"
                    mkdir $metricsFolder
                fi
                
                # Make a temporary copy of the config.sh file on the new folder
                echo "Generating the config.sh"
                cp "./networks/${networkName}/config.sh" "./examples/${folderName}"
                
                # Move to the example folder
                cd "./examples/${folderName}"

                # Append the bash env variables to the temp file
                echo "metricsFolder=\"${metricsFolder}\"" >> config.sh
                echo "armiarmaPath=\"${folderPath}\"" >> config.sh

                cd $metricsFolder
                TouchLauncher "$folderName"

                echo
                echo "Executing file $executableNetwork"
                echo "Exporting metrics at $metricsFolder"
                echo
                
                # Finaly launch Rumor form the Network File
                ../../src/bin/armiarma file launcher.rumor 
                
                # Maybe wait untill forcing the exit        
            fi
                 
            
            echo "Armiarma Finished!";;

        p)  # option for the ploter/analyzer (Temporary)
            analyzeFolder="$2"
            echo
            echo " Folder to be analyzed: $analyzeFolder"
            echo
            
            # Check if the virtual environment has been created
            if [[ -d ./src/analyzer/venv ]]; then
                echo "venv already created"
            else
                echo "Generating the virtual env"
                python3 -m virtualenv ./src/analyzer/venv
            fi
            
            # Source the virtual env 
            source ./src/analyzer/venv/bin/activate  
            pip3 install -r ./src/analyzer/requirements.txt
            
            aux="$analyzeFolder"
            # Set the Paths for the gossip-metrics.json peerstore.json and output
            metrics="./examples/${analyzeFolder}/metrics/gossip-metrics.json"
            peerstore="./examples/${aux}/metrics/peerstore.json"
            plots="./examples/${aux}/plots"
            csvs="./examples/${aux}/csvs"

            echo "metrics $peerstore"
            if [[ -d $plots ]]; then
                echo ""
            else
                mkdir "$plots"
            fi

            if [[ -d $csvs ]]; then
                echo ""
            else
                mkdir "$csvs"
            fi
            csvs="${csvs}/armiarma-metrics.csv"

            # Run the Analyzer
            python3 ./src/analyzer/armiarma-analyzer.py "json" "$peerstore" "$metrics" "$plots" "$csvs"
            echo "Analyzer Finished!"
            exit;;

        \?)         # incorrect option
            echo "Invalid option"
            echo
            Help
            exit;;        
    esac
done

echo ""
# 1. Check if the armiarma go executable exists on the /src forder



