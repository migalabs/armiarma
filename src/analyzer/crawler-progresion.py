# This file compiles the code to analyze the gathered metrics from each of the 
# folders in the $ARMIARMAPATH/examples folder, providing a global overview of the 
# Peer distributions and client types over the dayly analysis

import os, sys
import json
import time
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime

inittime = 0

def mainexecution():
    projectsFolder = sys.argv[1]
    outFolder = sys.argv[2]
    print("reading folder:", projectsFolder)
    print("output folder:", outFolder)
    outJson = outFolder + '/' + 'client-distribution.json'

    # So far, the generated panda will just contain the client count
    clientDist = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    stimationDist = {'date': [], 'Lighthouse': [], 'Teku': [], 'Nimbus': [], 'Prysm': [], 'Lodestar': [], 'Unknown': []}
    # Concatenation of the Json Values
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
            ps = fold + '/peerstore.json'
            if os.path.exists(f):
                # Load the readed values from the json into the panda
                poblatePandaObservedClients(clientDist, f, cf)
                poblatePandaStimatedClients(stimationDist, f, j, cf, ps)
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

    # Export the Concatenated json to the given folder
    with open(outJson, 'w', encoding='utf-8') as f:
        json.dump(j, f, ensure_ascii=False, indent=4)

    outputFigsFolder = outFolder + '/' + 'plots'
    clientList = ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown']
    figSize = (10,6)
    titleSize = 24
    textSize = 20
    labelSize = 20
    # Plot the images
    plotStackedChart(df, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'client-distribution.png',                               
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
            'figTitle': 'client-estimation-distribution.png',                                 
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
    
     # Plot the images
    plotStackedChart(sf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'Client-Distribution.png',                                 
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

    global inittime

    crwlig = 0
    crwtek = 0
    crwnim = 0
    crwpry = 0
    crwlod = 0
    crwunk = 0
    cnt = 0
    for k in jsonValues:
        peer = jsonValues[k]
        if 'MetadataRequest' in peer:
            if peer['MetadataRequest'] == True:
                cnt = cnt + 1
                if 'lig' in peer['ClientType'].lower():
                    crwlig = crwlig + 1
                elif 'teku' in peer['ClientType'].lower():
                    crwtek = crwtek + 1
                elif 'nimbus' in peer['ClientType'].lower():
                    crwnim = crwnim + 1
                elif 'prysm' in peer['ClientType'].lower():
                    crwpry = crwpry + 1
                elif 'js-libp2p' in peer['ClientType'].lower():
                    crwlod = crwlod + 1
                elif 'unk' in peer['ClientType'].lower():
                    crwunk = crwunk + 1
                else:
                    crwunk = crwunk + 1
        else:
            return
    
    print("total in metrics:", len(jsonValues))
    print("total requested:", cnt)

    if cnt == 0:
        print(jsonValues)
        return

    lig = round((crwlig*100)/cnt, 3)
    tek = round((crwtek*100)/cnt, 3)
    nim = round((crwnim*100)/cnt, 3)
    pry = round((crwpry*100)/cnt, 3)
    lod = round((crwlod*100)/cnt, 3)
    unk = round((crwunk*100)/cnt, 3)

    if cfValues['StopTime']['Month'] < 10:
	    month = "0" + str(cfValues['StopTime']['Month'])
    else: 
	    month = str(cfValues['StopTime']['Month'])

    if cfValues['StopTime']['Day'] < 10:
        day = "0" + str(cfValues['StopTime']['Day'])
    else: 
	    day = str(cfValues['StopTime']['Day'])

    if cfValues['StopTime']['Hour'] < 10:
	    hour = "0" + str(cfValues['StopTime']['Hour'])
    else: 
	    hour = str(cfValues['StopTime']['Hour'])
	
    if cfValues['StopTime']['Minute'] < 10:
	    minutes = "0" + str(cfValues['StopTime']['Minute'])
    else: 
	    minutes = str(cfValues['StopTime']['Minute'])

    """
    cday = str(cfValues['StopTime']['Year']) + '/' + month + '/' + day + '-' + hour + '-' +  minutes
    s = time.mktime(datetime.strptime(cday, "%Y/%m/%d-%H-%M").timetuple())
    if inittime == 0:
        inittime = s
    h = (s -inittime)/(60*60) # to get it in Hours
    clientDist['date'].append(h)
    """
    cday = str(cfValues['StopTime']['Year']) + '/' + month + '/' + day
    clientDist['date'].append(cday)
    clientDist['Lighthouse'].append(lig)
    clientDist['Teku'].append(tek)
    clientDist['Nimbus'].append(nim)
    clientDist['Prysm'].append(pry)
    clientDist['Lodestar'].append(lod)
    clientDist['Unknown'].append(unk)

def poblatePandaStimatedClients(stimatedDist, jsonFile, j, customFile, peerstore):
    # Read the Json
    jsf = open(jsonFile)
    jsonValues = json.load(jsf)
    jsf.close()
    
    cf = open(customFile)
    cfValues = json.load(cf)
    cf.close()

    ps = open(peerstore)
    psValues = json.load(ps)
    ps.close()

    j.append(cfValues)

    global inittime

    # Aux Variables
    tcp13000 = 0
    total = 0 
    
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
                if 'lig' in peer['ClientType'].lower():
                    crwlig = crwlig + 1
                elif 'teku' in peer['ClientType'].lower():
                    crwtek = crwtek + 1
                elif 'nimbus' in peer['ClientType'].lower():
                    crwnim = crwnim + 1
                elif 'js-libp2p' in peer['ClientType'].lower():
                    crwlod = crwlod + 1
                elif 'unk' in peer['ClientType'].lower():
                    crwunk = crwunk + 1
            """
            if '/13000' in peer['Addrs']:
                tcp13000 = tcp13000 + 1
            """
        else:
            return

    # iterate through the peerstore
    for k in psValues:
        peer = psValues[k]
        total = total + 1
        try:
            if '/13000' in peer['addrs'][0]:
                tcp13000 = tcp13000 + 1 
        except:
            pass
        

    #total = len(jsonValues)
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

    if cfValues['StopTime']['Month'] < 10:
	    month = "0" + str(cfValues['StopTime']['Month'])
    else: 
	    month = str(cfValues['StopTime']['Month'])

    if cfValues['StopTime']['Day'] < 10:
        day = "0" + str(cfValues['StopTime']['Day'])
    else: 
	    day = str(cfValues['StopTime']['Day'])

    if cfValues['StopTime']['Hour'] < 10:
	    hour = "0" + str(cfValues['StopTime']['Hour'])
    else: 
	    hour = str(cfValues['StopTime']['Hour'])
	
    if cfValues['StopTime']['Minute'] < 10:
	    minutes = "0" + str(cfValues['StopTime']['Minute'])
    else: 
	    minutes = str(cfValues['StopTime']['Minute'])

    """
    cday = str(cfValues['StopTime']['Year']) + '/' + month + '/' + day + '-' + hour + '-' +  minutes
    s = time.mktime(datetime.strptime(cday, "%Y/%m/%d-%H-%M").timetuple())
    if inittime == 0:
        inittime = s
    h = (s -inittime)/(60*60) # to get it in Hours
    stimatedDist['date'].append(h)
    """
    cday = str(cfValues['StopTime']['Year']) + '/' + month + '/' + day
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