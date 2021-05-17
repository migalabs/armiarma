# This file compiles the code to analyze the gathered metrics from each of the 
# folders in the $ARMIARMAPATH/examples folder, providing a global overview of the 
# Peer distributions and client types over the dayly analysis

import os, sys
import json
import pandas as pd
import matplotlib.pyplot as plt


def mainexecution():
    projectsFolder = sys.argv[1]

    # So far, the generated panda will just contain the client count
    p = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    print(projectsFolder)
    for root, dirs, _ in os.walk(projectsFolder):
        print(dirs)
        # We just need to access the folders inside examples
        for d in dirs:
            fold = os.path.join(root, d)
            f = fold + '/metrics/custom-metrics.json'
            if os.path.exists(f):
                # Load the readed values from the json into the panda
                poblatePandaCustomMetrics(p, f)
        break
    # After filling the dict with all the info from the JSONs, generate the panda
    df = pd.DataFrame (p, columns = ['date', 'Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown'])
    print(df)

    clientList = ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown']
    figSize = (10,6)
    titleSize = 24
    textSize = 20
    labelSize = 20
    # Plot the images
    plotStackedChart(df, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'CrawlSummary.png',                                          
            'align': 'center',
            'textSize': textSize,                                                  
            'title': "Evolution of the observed client distribution",           
            'ylabel': 'Client concentration (%)',                                                         
            'xlabel': None,
            'legendLabels': clientList,                    
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                     
            'tickRotation': 90,                                                     
            'show': True})

def poblatePandaCustomMetrics(p, jsonFile):
    print('poblating panda')
    # Read the Json
    jsf = open(jsonFile)
    jsonValues = json.load(jsf)
    jsf.close()

    # Compose the date of the crawling day
    cday = str(jsonValues['StartTime']['Year']) + '-' + str(jsonValues['StartTime']['Month']) + '-' + str(jsonValues['StartTime']['Day'])
    p['date'].append(cday)
    ps = jsonValues['PeerStore']['ConnectionSucceed']['Total']
    # Get percentages for each of the clients
    lig = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Lighthouse']['Total'])) / ps
    tek = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Teku']['Total'])) / ps
    nim = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Nimbus']['Total'])) / ps
    pry = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Prysm']['Total'])) / ps
    lod = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Lodestar']['Total'])) / ps
    unk = (100 * int(jsonValues['PeerStore']['ConnectionSucceed']['Unknown']['Total'])) / ps

    p['Lighthouse'].append(lig)
    p['Teku'].append(tek)
    p['Nimbus'].append(nim)
    p['Prysm'].append(pry)
    p['Lodestar'].append(lod)
    p['Unknown'].append(unk)


def plotStackedChart(p, opts):
    
    ax = p.plot.area(figsize = opts['figSize'], x='date', stacked=True)

    # labels
    if opts['ylabel'] is not None:     
        plt.ylabel(opts['ylabel'], fontsize=opts['labelSize'])
    if opts['xlabel'] is not None:
        plt.xlabel(opts['xlabel'], fontsize=opts['labelSize'])
    
    handles,legends = ax.get_legend_handles_labels()
    ax.legend(handles=handles, labels=opts['legendLabels'], loc='upper center', bbox_to_anchor=(0.5, -0.05), fancybox=True, shadow=True, ncol=6)

    plt.title(opts['title'], fontsize = opts['titleSize'])
    plt.tight_layout()
    #plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()

if __name__ == '__main__':
    mainexecution()