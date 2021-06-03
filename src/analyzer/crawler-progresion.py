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
    print("reading folder:", projectsFolder)
    print("output folder:", outFolder)
    outJson = outFolder + '/' + 'client-distribution.json'

    # So far, the generated panda will just contain the client count
    clientDist = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    stimationDist = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    # Concatenation of the Json values
    j = []
    print(projectsFolder)
    for root, dirs, _ in os.walk(projectsFolder):
        print(dirs)
        dirs.sort()
        # We just need to access the folders inside examples
        for d in dirs:
            fold = os.path.join(root, d)
            f = fold + '/gossip-metrics.json' 
            cf = fold + '/custom-metrics.json'
            if os.path.exists(f):
                # Load the readed values from the json into the panda
                poblatePandaObservedClients(clientDist, f, cf)
                poblatePandaStimatedClients(stimationDist, f, cf)
                print(f)
        break

    # After filling the dict with all the info from the JSONs, generate the panda
    df = pd.DataFrame (clientDist, columns = ['date', 'Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown'])
    df = df.sort_values(by="date")
    print(df)

      # After filling the dict with all the info from the JSONs, generate the panda
    sf = pd.DataFrame (stimationDist, columns = ['date', 'Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown'])
    sf = sf.sort_values(by="date")
    print(sf)

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
            'title': "Evolution of the experienced client distribution",           
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

    # Plot the images
    plotStackedChart(sf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'Client estimation distribution',  
            'figtitle': 'client-estimation-distribution.png',                                 
            'outputPath': outputFigsFolder,                                        
            'align': 'center',
            'grid': 'y',    
            'textSize': textSize,                                                  
            'title': "Evolution of the estimated client distribution",           
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

def poblatePandaObservedClients(clientDist, jsonFile, customFile):
    # Read the Json
    jsf = open(jsonFile)
    jsonValues = json.load(jsf)
    jsf.close()
    
    cf = open(customFile)
    cfValues = json.load(cf)
    cf.close()


    # Aux Variables
    tcp13000 = 0

    crwlig = 0
    crwtek = 0
    crwnim = 0
    crwlod = 0
    crwpry = 0
    crwunk = 0
    cnt = 0
    for k in jsonValues:
        peer = jsonValues[k]
        if 'MetadataRequest' in peer:
            if peer['MetadataRequest'] == True:
                cnt = cnt + 1
                if 'Light' in peer['ClientType']:
                    crwlig = crwlig + 1
                elif 'Teku' in peer['ClientType']:
                    crwtek = crwtek + 1
                elif 'Nimbus' in peer['ClientType']:
                    crwnim = crwnim + 1
                elif 'Prysm' in peer['ClientType']:
                    crwpry = crwpry + 1
                elif 'Lod' in peer['ClientType']:
                    crwlod = crwlod + 1
                elif 'Unk' in peer['ClientType']:
                    crwunk = crwunk + 1
        else:
            return
    
    total = crwlig + crwtek + crwtek + crwpry + crwlod + crwunk
    print("total in metrics:", len(jsonValues))
    print("total requested:", cnt)
    print("total of client sum:", total)

    if total == 0:
        print(jsonValues)
        return

    lig = round((crwlig*100)/total, 2)
    tek = round((crwtek*100)/total, 2)
    nim = round((crwnim*100)/total, 2)
    pry = round((crwpry*100)/total, 2)
    lod = round((crwlod*100)/total, 2)
    unk = round((crwunk*100)/total, 2)

    cday = str(cfValues['StopTime']['Year']) + '/' + str(cfValues['StopTime']['Month']) + '/' + str(cfValues['StopTime']['Day']) + '-' + str(cfValues['StopTime']['Hour'])
    clientDist['date'].append(cday)
    clientDist['Lighthouse'].append(lig)
    clientDist['Teku'].append(tek)
    clientDist['Nimbus'].append(nim)
    clientDist['Prysm'].append(pry)
    clientDist['Lodestar'].append(lod)
    clientDist['Unknown'].append(unk)

def poblatePandaStimatedClients(stimatedDist, jsonFile, customFile):
    # Read the Json
    jsf = open(jsonFile)
    jsonValues = json.load(jsf)
    jsf.close()
    
    cf = open(customFile)
    cfValues = json.load(cf)
    cf.close()

    # Aux Variables
    tcp13000 = 0

    crwlig = 0
    crwtek = 0
    crwnim = 0
    crwlod = 0
    crwunk = 0
    cnt = 0
    for k in jsonValues:
        peer = jsonValues[k]
        if 'MetadataRequest' in peer:
            if peer['MetadataRequest'] == True:
                cnt = cnt + 1
                if 'Light' in peer['ClientType']:
                    crwlig = crwlig + 1
                elif 'Teku' in peer['ClientType']:
                    crwtek = crwtek + 1
                elif 'Nimbus' in peer['ClientType']:
                    crwnim = crwnim + 1
                elif 'Lod' in peer['ClientType']:
                    crwlod = crwlod + 1
                elif 'Unk' in peer['ClientType']:
                    crwunk = crwunk + 1
            if '/13000' in peer['Addrs']:
                tcp13000 = tcp13000 + 1
        else:
            return

    total = len(jsonValues)
    print("total in metrics:", len(jsonValues))
    print("total tcp 13000:", tcp13000)

    res = total - tcp13000

    noPrysm = crwlig + crwtek + crwnim + crwlod + crwunk
    
    if total == 0 or noPrysm == 0:
        print(jsonValues)
        return

    estimlig = (res*crwlig)/noPrysm
    estimtek = (res*crwtek)/noPrysm
    estimnim = (res*crwnim)/noPrysm
    estimlod = (res*crwlod)/noPrysm
    estimunk = (res*crwunk)/noPrysm

    lig = round((estimlig*100)/total, 2)
    tek = round((estimtek*100)/total, 2)
    nim = round((estimnim*100)/total, 2)
    pry = round((tcp13000*100)/total, 2)
    lod = round((estimlod*100)/total, 2)
    unk = round((estimunk*100)/total, 2)

    cday = str(cfValues['StopTime']['Year']) + '/' + str(cfValues['StopTime']['Month']) + '/' + str(cfValues['StopTime']['Day']) + '-' + str(cfValues['StopTime']['Hour'])
    stimatedDist['date'].append(cday)
    stimatedDist['Lighthouse'].append(lig)
    stimatedDist['Teku'].append(tek)
    stimatedDist['Nimbus'].append(nim)
    stimatedDist['Prysm'].append(pry)
    stimatedDist['Lodestar'].append(lod)
    stimatedDist['Unknown'].append(unk)


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