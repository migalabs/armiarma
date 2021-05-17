# This file compiles the code to analyze the gathered metrics from each of the 
# folders in the $ARMIARMAPATH/examples folder, providing a global overview of the 
# Peer distributions and client types over the dayly analysis

import os, sys
import json
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np


def mainexecution():
    projectsFolder = sys.argv[1]
    outFolder = sys.argv[2]

    outJson = outFolder + '/' + 'client-distribution.json'
    # So far, the generated panda will just contain the client count
    p = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    # Concatenation of the Json values
    j = []
    print(projectsFolder)
    for root, dirs, _ in os.walk(projectsFolder):
        print(dirs)
        dirs.sort()
        # We just need to access the folders inside examples
        for d in dirs:
            fold = os.path.join(root, d)
            f = fold + '/metrics/custom-metrics.json'
            if os.path.exists(f):
                # Load the readed values from the json into the panda
                poblatePandaCustomMetrics(p, j, f)
        break

    # Export the Concatenated json to the given folder
    with open(outJson, 'w', encoding='utf-8') as f:
        json.dump(j, f, ensure_ascii=False, indent=4)

    # After filling the dict with all the info from the JSONs, generate the panda
    df = pd.DataFrame (p, columns = ['date', 'Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown'])
    df = df.sort_values(by="date")
    print(df)

    outputFigsFolder = outFolder + '/' + 'plots'
    clientList = ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown']
    figSize = (10,6)
    titleSize = 24
    textSize = 20
    labelSize = 20
    # Plot the images
    plotStackedChart(df, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'Client-Distribution.png',  
            'figtitle': 'client-distribution.png',                                 
            'outputPath': outputFigsFolder,                                        
            'align': 'center',
            'grid': 'y',    
            'textSize': textSize,                                                  
            'title': "Evolution of the projected client distribution",           
            'ylabel': 'Client concentration (%)',                                                         
            'xlabel': None, 
            'yticks': np.arange(0, 110, 10),
            'legendLabels': clientList,                    
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                     
            'tickRotation': 90,                                                     
            'show': True})

def poblatePandaCustomMetrics(p, j, jsonFile):
    # Read the Json
    jsf = open(jsonFile)
    jsonValues = json.load(jsf)
    jsf.close()
    
    # Add readed Json to the previous ones (concatenate them) 
    j.append(jsonValues)

    # Compose the date of the crawling day
    cday = str(jsonValues['StartTime']['Year']) + '-' + str(jsonValues['StartTime']['Month']) + '-' + str(jsonValues['StartTime']['Day'])
    p['date'].append(cday)
    ps = jsonValues['PeerStore']['Total']
    # Get percentages for each of the clients
    crwlig = jsonValues['PeerStore']['ConnectionSucceed']['Lighthouse']['Total']
    crwtek = jsonValues['PeerStore']['ConnectionSucceed']['Teku']['Total']
    crwnim = jsonValues['PeerStore']['ConnectionSucceed']['Nimbus']['Total']
    crwlod = jsonValues['PeerStore']['ConnectionSucceed']['Lodestar']['Total']
    crwunk = jsonValues['PeerStore']['ConnectionSucceed']['Unknown']['Total']

    prysPort = jsonValues['PeerStore']['Port13000']
    res = ps - prysPort

    noPrysm = crwlig + crwtek + crwnim + crwlod + crwunk
    
    estimlig = (res*crwlig)/noPrysm
    estimtek = (res*crwtek)/noPrysm
    estimnim = (res*crwnim)/noPrysm
    estimlod = (res*crwlod)/noPrysm
    estimunk = (res*crwunk)/noPrysm

    lig = round((estimlig*100)/ps, 2)
    tek = round((estimtek*100)/ps, 2)
    nim = round((estimnim*100)/ps, 2)
    pry = round((prysPort*100)/ps, 2)
    lod = round((estimlod*100)/ps, 2)
    unk = round((estimunk*100)/ps, 2)

    p['Lighthouse'].append(lig)
    p['Teku'].append(tek)
    p['Nimbus'].append(nim)
    p['Prysm'].append(pry)
    p['Lodestar'].append(lod)
    p['Unknown'].append(unk)


def plotStackedChart(p, opts):
    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']    

    ax = p.plot.area(figsize = opts['figSize'], x='date', stacked=True)

    # labels
    if opts['ylabel'] is not None:     
        plt.ylabel(opts['ylabel'], fontsize=opts['labelSize'])
    if opts['xlabel'] is not None:
        plt.xlabel(opts['xlabel'], fontsize=opts['labelSize'])
    
    handles,legends = ax.get_legend_handles_labels()
    ax.legend(handles=handles, labels=opts['legendLabels'], loc='upper center', bbox_to_anchor=(0.5, -0.05), fancybox=True, shadow=True, ncol=6)

    if opts['grid'] != None:
        ax.grid(which='major', axis=opts['grid'], linestyle='--')

    plt.yticks(opts['yticks'])

    plt.title(opts['title'], fontsize = opts['titleSize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()

if __name__ == '__main__':
    mainexecution()