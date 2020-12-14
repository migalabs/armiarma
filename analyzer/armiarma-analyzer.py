# Armiarma Metrics Analyzer
# Script that ussing the rumor-metrics.py package generates the plots

import os, sys
import json 
import time
import requests
from IPy import IP
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.colors as mcolors
import numpy as np
import durationpy


### TODO:
#        - Organize the plotting part
#      **- Organize the panda by groups (mainly on type of Clients)
#      **- Could be nice also to say which kind of client/version of the client are they running

def getDictFromJson(inputFile):
    print("reading json: ", inputFile)
    mf = open(inputFile)
    peerMetrics = json.load(mf)
    mf.close()

    return peerMetrics

def getModificationTimeOfFile(inputFile):
    statbuf = os.stat(inputFile)
    return statbuf.st_mtime*1000


# Generate the pandaobject of all the metrics per peer
def getPandaobjectFromJson(inputFile, peerstoreMetrics):
    print("Panda from Json")
    peerMetrics = getDictFromJson(inputFile)
    fileTime = getModificationTimeOfFile(inputFile)

    ## Temp -to get the Location of the IPs
    cont=0
    ##
    
    # Define the panda 
    pMetrics = {'PeerId': [], 'NodeId':[], 'ClientType':[], 'Pubkey':[], 'Addrs':[], 'Ip':[], 'Country':[], 'City':[], 'Latency':[], 'Connections':[], 'Disconnections':[], 'ConnectedTime':[], 'BeaconBlockCnt':[], 'BeaconAggregateProofCnt':[], 'VoluntaryExitCnt':[], 'ProposerSlashingCnt':[], 'AttesterSlashingCnt':[]}
    
    for peer in peerMetrics:
        # Flag for the private IPs
        ipPublic = 0

        pMetrics['PeerId'].append(peerMetrics[peer]["PeerId"])
        pMetrics['NodeId'].append(peerMetrics[peer]["NodeId"])
        try:
            actualClientType = peerstoreMetrics[peer]["user_agent"]
        except:
            actualClientType = "Unknown"

        try: 
            actualLatency = float(peerstoreMetrics[peer]["latency"]/1000000000) # from nanoseconds to seconds
        except:
            actualLatency = None

        print("Peer", peer, "is form Client:", actualClientType) 
        print("Peer", peer, "has latency:", actualLatency)
        pMetrics['ClientType'].append(actualClientType)
        pMetrics["Latency"].append(actualLatency)
        pMetrics['Pubkey'].append(peerMetrics[peer]['Pubkey'])
        
        addrsCnt = 0
        ip = IP(peerMetrics[peer]['Ip'])
        if ip.iptype() == 'PRIVATE' or ip.iptype() == 'LOOPBACK' or peerMetrics[peer]['Country'] == '':
            for idx, address in enumerate(peerMetrics[peer]['Addrs']):
                ipx = address.replace('/ip4/', '')
                ipx = ipx.split('/')[0]
                ipx=IP(ipx)
                if ipx.iptype() == 'PUBLIC':
                    pMetrics['Ip'].append(str(ipx))
                    country, city = getLocationFromIp(ipx)
                    pMetrics['Country'].append(country)
                    pMetrics['City'].append(city)
                    pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][addrsCnt])
                    print('(Private from the beggining) Added:', ipx, ipx.iptype(),country, city)
                    ipPublic = 1
                    cont=cont+1
                    break
                addrsCnt=addrsCnt+1

            if ipPublic == 0:
                print(ip, ip.iptype())
                pMetrics['Country'].append('Unknown')
                pMetrics['City'].append('Unknown')
                print('(Private) Added:', ip, 'Unknown, Unknown')
                pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][0])
        else:           
            pMetrics['Country'].append(peerMetrics[peer]['Country'])
            pMetrics['City'].append(peerMetrics[peer]['City'])
            pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][0])
            print('(Public from the beggining) Added:', ip, ip.iptype(), peerMetrics[peer]['Country'], peerMetrics[peer]['City'])

        connection, disconnection, ttime = GetConnectDisconnectAndConTime(peer, peerMetrics, fileTime)
        pMetrics['Connections'].append(connection)
        pMetrics['Disconnections'].append(disconnection)
        pMetrics['ConnectedTime'].append(ttime)
            
        pMetrics['BeaconBlockCnt'].append(peerMetrics[peer]['BeaconBlock']['Cnt'])
        pMetrics['BeaconAggregateProofCnt'].append(peerMetrics[peer]['BeaconAggregateProof']['Cnt'])
        pMetrics['VoluntaryExitCnt'].append(peerMetrics[peer]['VoluntaryExit']['Cnt'])
        pMetrics['ProposerSlashingCnt'].append(peerMetrics[peer]['ProposerSlashing']['Cnt'])
        pMetrics['AttesterSlashingCnt'].append(peerMetrics[peer]['AttesterSlashing']['Cnt'])

        # To dont exeed the limit of petitions per minute
        if cont >= 35:
                time.sleep(70)
                cont=0

    print('len PeerId:', pMetrics['PeerId'])
    print('len ClientType:', pMetrics['ClientType'])
    print('len Addrs:', pMetrics['Addrs'])
    print('len Country:', pMetrics['Country'])
    print('len City:', pMetrics['City'])
    print('len Latency:', pMetrics['Latency'])
    print('len Connections:', pMetrics['Connections'])

    pandaObject = pd.DataFrame(pMetrics, columns = ['PeerId', 'NodeId', 'ClientType', 'Pubkey', 'Addrs', 'Country',
     'City', 'Latency', 'Connections', 'Disconnections', 'ConnectedTime', 'BeaconBlockCnt', 'BeaconAggregateProofCnt',
      'VoluntaryExitCnt', 'ProposerSlashingCnt', 'AttesterSlashingCnt'])
    return pandaObject

      
# request the location from api
def getLocationFromIp(ipAddress):
    print(ipAddress)
    composedUrl = f"http://ip-api.com/json/{ipAddress}"
    resp = requests.get(url=composedUrl) 
    print(resp)
    try:
        array = resp.json()
        if array["status"] == "success":
            return array["country"], array["city"]
        else:
            return "Unknown", "Unknown"
    except:
        return "Unknown", "Unknown"
       
# Get Connections, Disconnections and Time from each peer
def GetConnectDisconnectAndConTime(peer, peerMetrics, fileTime):
    connectionCounter = 0
    disconnectionCounter= 0
    connectionTotalTime = 0
    ctime = 0 # aux variable to calculate the final time
    timeFlag = 0

    for connection in peerMetrics[peer]["ConnectionEvents"]:
        if connection["ConnectionType"] == "Connection" :
            connectionCounter += 1
            if timeFlag == 0:
                ctime = connection["TimeMili"] # secs
                timeFlag = 1
        if connection["ConnectionType"] == "Disconnection" :
            disconnectionCounter += 1
            if timeFlag == 1:
                connectionTotalTime = connectionTotalTime + (connection["TimeMili"] - ctime)
                timeFlag = 0
    # if the flag is 1, means that on the moment of taking the metrics we were connected
    if timeFlag == 1:
        connectionTotalTime = connectionTotalTime + (fileTime - ctime)
    return connectionCounter, disconnectionCounter, connectionTotalTime/60000

########### ------------------ Ploting Stage/Code AKA Wonderland ------------
# TODO: - At one point would be nice to add the ploting stuff on a library itself


def plotBarsFromPandas(panda, opts):
    print("Bar Graph from Panda")

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)


    #panda = pandaOriginal.sort_values(by=opts['xmetrics'], ascending=False, inplace=False)
    ax = panda[opts['xmetrics']].sort_values(by=opts['xmetrics'], ascending=False).plot(kind='bar', figsize=opts['figSize'], logy=opts['ylog'], legend=opts['legend'], color=opts['barColor']) 
    
    # labels
    if opts['ylabel'] is not None:    
        plt.ylabel(opts['ylabel'], fontsize=opts['labelSize'])
    if opts['xlabel'] is not None:
        plt.xlabel(opts['xlabel'], fontsize=opts['labelSize'])

    # Ticks LABELS
    if opts['xticks'] is not None:
        plt.xticks(range(len(panda)), opts['xticks'], rotation=opts['tickRotation'], fontsize=opts['xticksSize'])
    else: 
        ax.get_xaxis().set_ticks([])
    plt.yticks(fontsize=opts['yticksSize'])
    plt.ylim(opts['yLowLimit'], opts['yUpperLimit'])
    
    # Check is there is Value on top of the charts
    if opts['barValues'] is not None:
        for ind, value in enumerate(yarray):
            plt.text(ind, value, str(value), fontsize=opts['textSize'], horizontalalignment='center')

    # Title
    plt.title(opts['title'], fontsize = opts['titleSize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()



# CortzePlot extension to plot bar-charts
def plotBarsFromArrays(xarray, yarray, opts):
    print("Bar Graph from Arrays")

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    fig = plt.figure(figsize = opts['figSize'])

    plt.bar(range(len(xarray)), yarray, align=opts['align'], color=opts['barColor'])


    # labels
    if opts['ylabel'] is not None:    
        plt.ylabel(opts['ylabel'], fontsize=opts['labelSize'])
    if opts['xlabel'] is not None:
        plt.xlabel(opts['xlabel'], fontsize=opts['labelSize'])

    # Ticks LABELS
    if opts['xticks'] is not None:
        plt.xticks(range(len(xarray)), opts['xticks'], rotation=opts['tickRotation'], fontsize=opts['xticksSize'])
    else: 
        plt.xticks(range(len(xarray)))

    plt.margins(x=0)
    plt.yticks(fontsize=opts['yticksSize'])
    plt.ylim(opts['yLowLimit'], opts['yUpperLimit'])
    
    # Check is there is Value on top of the charts
    if opts['barValues'] is not None:
        for ind, value in enumerate(yarray):
            plt.text(ind, value, str(value), fontsize=opts['textSize'], horizontalalignment='center')

    # Title
    plt.title(opts['title'], fontsize = opts['titleSize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()


# Reuturns on 2 arrays the names found and the values, grouped by te clientNames/arrayOfNames
def sortArrayByNames(xarray, yarray, clientNames):
    namesarray = []
    valuesarray = [] 

    for idx, clientType in enumerate(clientNames):
        auxnames = []
        auxvalues = []
        for index, item in enumerate(xarray):
            if clientType.lower() in item.lower():
                auxnames.append(item)
                auxvalues.append(int(yarray[index]))
        namesarray.append(auxnames)

        if not auxvalues:
            auxvalues = [0]
        valuesarray.append(auxvalues)

    return namesarray, valuesarray

# Funtion that plots the given array into a pie chart 
def plotSinglePieFromArray(xarray, opts):
    print("Pie Graph from Arrays")

    outputFile = str(opts['outputpath']) + '/' + opts['figtitle']
    print('printing image', opts['figtitle'], 'on', outputFile)

    fig, ax = plt.subplots(figsize=opts['figsize'])

    size = opts['piesize']
    
    """
    cmap = plt.get_cmap(opts['piecolors'])
    pcolors = cmap(np.arange(1)*len(xarray))
    """

    patches1, labels1 = ax.pie(xarray, radius=1, colors=opts['piecolors'], labels=opts['labels'],
           wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))


    if opts['legend'] == True:
        plt.legend(patches1, labels1, loc=opts['lengendposition'])
    ax.set(aspect="equal")

    # Title
    plt.title(opts['title'], fontsize = opts['titlesize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()

# Autoformat to actually show the values of the pie chart
def autopct_format(values):
    def my_format(pct):
        total = sum(values)
        val = int(round(pct*total/100.0))
        return '{v:d}'.format(v=val)
    return my_format


# Funtion that plots the given array into a pie chart 
def plotDoublePieFromArray(xarray, opts):
    print("Pie Graph from Arrays")

    outputFile = str(opts['outputpath']) + '/' + opts['figtitle']
    print('printing image', opts['figtitle'], 'on', outputFile)

    fig, ax = plt.subplots(figsize=opts['figsize'])

    size = opts['piesize']
    valsouter = []
    valsinner = []
    for _, item in enumerate(xarray):
        total=0
        for _, valaux in enumerate(item):
            total=total+valaux
            valsinner.append(valaux)
        valsouter.append(total)

    cnt = 0
    # Temporal plot for the inner color_grids
    for idx, item in enumerate(opts['innercolors']):
        print(len(xarray[idx]))
        aux = plt.get_cmap(item, len(xarray[idx]))
        auxarray = aux(range(len(xarray[idx])))
        if cnt == 0:
            innercolors = auxarray
        else:
            innercolors = np.concatenate((innercolors, auxarray), axis=0)
        cnt = cnt + 1

    print(valsouter)

    if opts['autopct'] == 'values':
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1, colors=opts['outercolors'], labels=opts['outerlabels'], 
                    labeldistance=opts['labeldistance'], autopct=autopct_format(valsouter), pctdistance=opts['pctdistance'],
                     wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))
    elif opts['autopct'] == 'pcts':
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1, colors=opts['outercolors'], labels=opts['outerlabels'], 
                    labeldistance=opts['labeldistance'], autopct='%d', pctdistance=opts['pctdistance'], 
                    wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))

    elif opts['autopct'] == False:
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1, colors=opts['outercolors'], labels=None, 
                    labeldistance=opts['labeldistance'], autopct=autopct_format(valsouter), pctdistance=opts['pctdistance'], 
                    wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))

    for idx, _ in enumerate(labels1):
        autotext[idx].set_fontsize(opts['labelsize'])
        autotext[idx].set_c(opts['outercolors'][idx])
        if opts['autopct'] != False:
            labels1[idx].set_fontsize(opts['labelsize'])
            labels1[idx].set_c(opts['outercolors'][idx])

    # , labels=opts['innerlabels']
    patches2, labels2 = ax.pie(valsinner, radius=1-size, colors=innercolors,
           wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))

    if opts['legend'] == True:
        plt.legend(opts['outerlabels'], bbox_to_anchor=(1, 0.75), loc=opts['lengendposition'], fontsize=opts['labelsize'], 
           bbox_transform=plt.gcf().transFigure)
        plt.subplots_adjust(left=0.0, bottom=0.1, right=0.85)
    ax.set(aspect="equal")

    # Title
    plt.title(opts['title'], fontsize = opts['titlesize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    if opts['show'] is True:
        plt.show()

# Funtion that gives length of the panda
def getLengthOfPanda(panda):
    return len(panda)

# Function that gets the data (counter, sum, avg) of the given metric from the panda
def getDataFromPanda(panda, ymetrics, xmetrics, xarray, flag):
    yarray = []
    if flag == 'counter':
        for _, item in enumerate(xarray):
            auxAmount = panda.apply(lambda x: True if item.lower() in str(x[xmetrics]).lower() else False, axis=1)
            yarray.append(len(auxAmount[auxAmount == True].index))
    elif flag == 'sum':
        for _, item in enumerate(xarray):
            item = str(item)
            auxCnt = 0
            for index, row in panda.iterrows():
                if item.lower() in str(row[xmetrics]).lower():
                    auxCnt = auxCnt + int(row[ymetrics]) 
            yarray.append(auxCnt)
    elif flag =='avg':
        for _, item in enumerate(xarray):            
            auxCnt = 0
            for index, row in panda.iterrows():
                if item.lower() in str(row[xmetrics]).lower():
                    auxCnt = auxCnt + int(row[ymetrics])
            auxAmount = panda.apply(lambda x: True if item.lower() in str(x[xmetrics]).lower() else False, axis=1)
            if auxCnt != 0:
                yarray.append(round((auxCnt/(len(auxAmount[auxAmount == True].index))),1))
            else:
                yarray.append(0)
    else:
        print("Default Aplication on getDataFromPanda")
    
    return xarray, yarray

# Funtion that gets how many different items are detected
def getItemsFromColumn(panda, ymetrics):
    itemList = []
    for index, row in panda.iterrows():
        if not itemList:
            itemList.append(row[ymetrics])
        else:
            if row[ymetrics] not in itemList:
                itemList.append(row[ymetrics])
    return itemList 


# Get the raimbow colors
def GetColorGridFromArray(yarray):
    clist = [(0, "red"), (0.125, "red"), (0.25, "orange"), (0.5, "green"), 
         (0.7, "green"), (0.75, "blue"), (1, "blue")]
    rvb = mcolors.LinearSegmentedColormap.from_list("", clist)

    N = len(yarray)
    maxVal = np.max(yarray)
    x = np.arange(N).astype(float)
    y = np.random.uniform(0, maxVal, size=(N,))
    grid = rvb(x/N)
    return grid

def GetColorGridFromPanda(panda, ymetric):
    clist = [(0, "red"), (0.125, "red"), (0.25, "orange"), (0.5, "green"), 
         (0.7, "green"), (0.75, "blue"), (1, "blue")]
    rvb = mcolors.LinearSegmentedColormap.from_list("", clist)

    N = len(panda)
    maxVal = panda[ymetric].max()
    x = np.arange(N).astype(float)
    y = np.random.uniform(0, maxVal, size=(N,))
    print(N, maxVal)
    grid = rvb(x/N)
    return grid



# MAIN FUNTION, describes the execution secuence or workflow
def main():
    progName = "Armiarma Analyzer Running!"

    # Variables for Plotting
    figSize = (10,6)
    wideFigSize = (12,7)
    titleSize = 22
    labelSize = 22
    ticksSize = 22 
    textSize = 14
    # End of plotting variables
    print(sys.argv[1])
    if sys.argv[1] == 'json':
        peerstoreFile = sys.argv[2]
        rumorMetricsFile = sys.argv[3]
        outputFigsFolder = sys.argv[4]
        outputCsvFile = sys.argv[5]

        # Start preparing the data for a later plots
        peerstoreMetrics = getDictFromJson(peerstoreFile)
        rumorMetricsPanda = getPandaobjectFromJson(rumorMetricsFile, peerstoreMetrics)
        
        rumorMetricsPanda.to_csv(outputCsvFile)
        print(rumorMetricsPanda)

    if sys.argv[1] == 'csv':
        peerstoreFile = sys.argv[2]
        rumorMetricsFile = sys.argv[3]
        outputFigsFolder = sys.argv[4]
        
        peerstoreMetrics  = getDictFromJson(peerstoreFile)
        rumorMetricsPanda = pd.read_csv(rumorMetricsFile)
    
    # ------ Get data for plotting -------
    clientList = ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown']
    clientColors = ['tab:blue', 'tab:orange', 'tab:green', 'tab:red', 'tab:purple', 'k' ]
    innerColors = ['Blues', 'Oranges', 'Greens', 'Reds', 'Purples', 'Greys' ]
    print(clientList)
    

    # get length of the peerstore
    peerstoreSize = getLengthOfPanda(peerstoreMetrics)
    print("Peerstore Size:", peerstoreSize)
    peerMetricsSize = getLengthOfPanda(rumorMetricsPanda)
    print("Peerstore Size:", peerMetricsSize)

    xarray = ['Peerstore', 'Connected Peers']
    yarray = [peerstoreSize, peerMetricsSize]
    barColor = ['tab:blue', 'tab:green']

    plotBarsFromArrays(xarray, yarray, opts={                                   
        'figSize': figSize,                                                      
        'figTitle': 'PeerstoreVsConnectedPeers.png',                                    
        'outputPath': outputFigsFolder,                                         
        'align': 'center',                                                      
        'barValues': True,
        'barColor': barColor,
        'textSize': textSize,
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Numer of Peers Connected from the entire Peerstore",                  
        'xlabel': None,                                                         
        'ylabel': 'Number of Peers',                                      
        'xticks': xarray,                       
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                       
        'tickRotation': 0,                                                     
        'show': False})
    
    """
    # get the number of peers per client
    print(clientList)
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, None, "ClientType", clientList, 'counter')
    """

    clientVersList = getItemsFromColumn(rumorMetricsPanda, 'ClientType')
    print(clientVersList)
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, None, "ClientType", clientVersList, 'counter')
    namesarray, valuesarray = sortArrayByNames(xarray, yarray, clientList)

    plotDoublePieFromArray(valuesarray, opts={                                   
        'figsize': figSize,                                                      
        'figtitle': 'PeersPerClient.png',                                    
        'outputpath': outputFigsFolder,
        'piesize': 0.3,                                                      
        'autopct': False,
        'pctdistance': 1.1,
        'edgecolor': 'w',
        'innerlabels': clientVersList,
        'outerlabels': clientList,
        'labeldistance': 1.25,
        'innercolors': innerColors,
        'outercolors': clientColors,
        'shadow': None,
        'startangle': 90,                                                  
        'title': "Numer of Peers From Each Client and Their Versions",                   
        'titlesize': titleSize,                                                        
        'labelsize': labelSize, 
        'legend': True,                                                       
        'lengendposition': None,                                                   
        'legendsize': labelSize,                                                     
        'show': False})

    """
    plotBarsFromArrays(xarray, yarray, opts={                                   
        'figSize': figSize,                                                      
        'figTitle': 'PeersPerClient.png',                                    
        'outputPath': outputFigsFolder,                                         
        'align': 'center',                                                      
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                  
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Numer of Peers Connected from each Client",                  
        'xlabel': None,                                                         
        'ylabel': 'Number of Connections',                                      
        'xticks': xarray,                       
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                        
        'tickRotation': 0,                                                     
        'show': False}) 
    """

    # get the number of peers per country 
    countriesList = getItemsFromColumn(rumorMetricsPanda, 'Country') 
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, None, "Country", countriesList, 'counter') 

    # Get Color Grid
    barColor = GetColorGridFromArray(yarray)
    
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': (12,7),                                                          
        'figTitle': 'PeersPerCountries.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': barColor,
        'textSize': textSize+2,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Numer of Peers Connected from each Country",                             
        'xlabel': None,                                   
        'ylabel': 'Number of Connections',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize+2,                                                        
        'labelSize': labelSize+2,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize+2,                                                       
        'xticksSize': ticksSize-2,                                                       
        'yticksSize': ticksSize+2,                                                            
        'tickRotation': 90,
        'show': False})   
    """
    plotSinglePieFromArray(yarray, opts={                                   
        'figsize': figSize,                                                      
        'figtitle': 'PeerstoreVsConnectedPeers.png',                                    
        'outputpath': outputFigsFolder,
        'piesize': 0.3,                                                      
        'autopct': '%f.f',
        'edgecolor': 'w',
        'piecolors': barColor,
        'labels': countriesList,
        'shadow': None,
        'startangle': 90,                                                  
        'title': "Numer of Peers From Each Client and Their Versions",                   
        'titlesize': titleSize,                                                        
        'labelsize': labelSize, 
        'legend': False,                                                       
        'lengendposition': 1,                                                   
        'legendsize': labelSize,                                                     
        'show': False})
    """

    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Connections", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'AverageOfConnectionsPerClientType.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Connections per Client Type",                             
        'xlabel': None,                                   
        'ylabel': 'Number of Connections',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                           
        'tickRotation': 0,
        'show': False}) 

    # get the average of disconnections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Disconnections", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'AverageOfDisconnectionsPerClientType.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Disconnections per Client Type",                             
        'xlabel': None,                                   
        'ylabel': 'Number of Disconnections',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                             
        'tickRotation': 0,
        'show': False}) 

    # get the average of ConnectedTime per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "ConnectedTime", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'AverageOfConnectedTimePerClientType.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Connected Time to Peers from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Time (Minutes)',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                             
        'tickRotation': 0,
        'show': False}) 

    # GossipSub
    # BeaconBlock
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "BeaconBlockCnt", "ClientType", clientList, 'sum') 
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessagesFromBeaconBlock.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Number of Received BeaconBlock Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                            
        'tickRotation': 0,
        'show': False})   


    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "BeaconBlockCnt", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessageAverageFromBeaconBlock.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Received BeaconBlock Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                         
        'tickRotation': 0,
        'show': False}) 

    # BeaconAggregateAndProof
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "BeaconAggregateProofCnt", "ClientType", clientList, 'sum') 
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessagesFromBeaconAggregateProof.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Number of Received BeaconAggregateAndProof Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                          
        'tickRotation': 0,
        'show': False})   


    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "BeaconAggregateProofCnt", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessageAverageFromBeaconAggregateProof.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Received BeaconAggregateAndProof Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                        
        'tickRotation': 0,
        'show': False}) 

    # VoluntaryExit
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "VoluntaryExitCnt", "ClientType", clientList, 'sum') 
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessagesFromVoluntaryExit.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Number of Received VoluntaryExit Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                         
        'tickRotation': 0,
        'show': False})   


    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "VoluntaryExitCnt", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessageAverageFromVoluntaryExit.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Received VoluntaryExit Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                       
        'tickRotation': 0,
        'show': False}) 

    # AttesterSlashing
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "AttesterSlashingCnt", "ClientType", clientList, 'sum') 
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessagesFromAttesterSlashing.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Number of Received AttesterSlashing Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                     
        'tickRotation': 0,
        'show': False})   


    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "AttesterSlashingCnt", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessageAverageFromAttesterSlashing.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Received AttesterSlashing Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                           
        'tickRotation': 0,
        'show': False}) 

    # ProposerSlashing
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "ProposerSlashingCnt", "ClientType", clientList, 'sum') 
    
    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessagesFromProposerSlashing.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Number of Received ProposerSlashing Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                           
        'tickRotation': 0,
        'show': False})   


    # get the average of connections per client
    xarray, yarray = getDataFromPanda(rumorMetricsPanda, "ProposerSlashingCnt", "ClientType", clientList, 'avg') 

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'MessageAverageFromProposerSlashing.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': None,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average of Received ProposerSlashing Messages from Clients",                             
        'xlabel': None,                                   
        'ylabel': 'Messages Received',                                                
        'xticks': xarray,                                                           
        'titleSize': titleSize,                                                        
        'labelSize': labelSize,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize,                                                       
        'xticksSize': ticksSize,                                                       
        'yticksSize': ticksSize,                                                             
        'tickRotation': 0,
        'show': False}) 



    # Plotting from the panda
    barColor = 'black'
    plotBarsFromPandas(rumorMetricsPanda, opts={                                   
        'figSize': wideFigSize,                                                      
        'figTitle': 'ConnectionsWithPeers.png',                                    
        'outputPath': outputFigsFolder,
        'legend': False,                                         
        'align': 'center',
        'ylog': False,
        'xmetrics': ['Connections'],                                                      
        'barValues': None,    
        'barColor': barColor,                                              
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Numer of Connections with each Peer",                  
        'xlabel': "Peers Connected",                                                         
        'ylabel': 'Number of Connections',                                      
        'xticks': None,                                                       
        'titleSize': titleSize+2,                                                        
        'labelSize': labelSize+2,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize+2,                                                       
        'xticksSize': ticksSize+2,                                                       
        'yticksSize': ticksSize+2,                                                      
        'tickRotation': 0,                                                     
        'show': False}) 

    barColor = 'black'
    plotBarsFromPandas(rumorMetricsPanda, opts={                                   
        'figSize': wideFigSize,                                                      
        'figTitle': 'DisconnectionsWithPeers.png',                                    
        'outputPath': outputFigsFolder,
        'legend': False,                                         
        'align': 'center',
        'ylog': False,
        'xmetrics': ['Disconnections'],                                                      
        'barValues': None,
        'barColor': barColor,                                                  
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Numer of Disconnections with each Peer",                  
        'xlabel': "Peers Connected",                                                         
        'ylabel': 'Number of Disconnections',                                      
        'xticks': None,                                                       
        'titleSize': titleSize+2,                                                        
        'labelSize': labelSize+2,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize+2,                                                       
        'xticksSize': ticksSize +2,                                                       
        'yticksSize': ticksSize +2,                                                        
        'tickRotation': 0,                                                     
        'show': False}) 

    barColor = 'black'
    plotBarsFromPandas(rumorMetricsPanda, opts={                                   
        'figSize': wideFigSize,                                                      
        'figTitle': 'TimeConnectedWithPeers.png',                                    
        'outputPath': outputFigsFolder,
        'legend': False,                                         
        'align': 'center',
        'ylog': False,
        'xmetrics': ['ConnectedTime'],                                                      
        'barValues': None,   
        'barColor': barColor,                                               
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Total of Time Connected with each Peer",                  
        'xlabel': "Peers Connected",                                                         
        'ylabel': 'Time (in Minutes)',                                      
        'xticks': None,                                                       
        'titleSize': titleSize +2,                                                        
        'labelSize': labelSize +2,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize +2,                                                       
        'xticksSize': ticksSize +2,                                                       
        'yticksSize': ticksSize +2,                                                     
        'tickRotation': 0,                                                     
        'show': False}) 

    barColor = 'black'

    print(rumorMetricsPanda.loc[rumorMetricsPanda['Latency'].idxmax()])

    plotBarsFromPandas(rumorMetricsPanda, opts={                                   
        'figSize': wideFigSize,                                                      
        'figTitle': 'LatencyWithPeers.png',                                    
        'outputPath': outputFigsFolder,
        'legend': False,                                         
        'align': 'center',
        'ylog': False,
        'xmetrics': ['Latency'],                                                      
        'barValues': None,   
        'barColor': barColor,                                               
        'yLowLimit': 0,                                                         
        'yUpperLimit': None,                                                    
        'title': "Latency with each Peer",                  
        'xlabel': "Peers Connected",                                                         
        'ylabel': 'Seconds',                                      
        'xticks': None,                                                       
        'titleSize': titleSize +2,                                                        
        'labelSize': labelSize +2,                                                        
        'lengendPosition': 1,                                                   
        'legendSize': labelSize +2,                                                       
        'xticksSize': ticksSize +2,                                                       
        'yticksSize': ticksSize +2,                                                     
        'tickRotation': 0,                                                     
        'show': False}) 

    # ------ End of Get data -------

    print("Finished!, Ciao!")



if __name__ == '__main__':
    main()
