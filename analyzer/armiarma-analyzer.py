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
import datetime 
import collections

# values that will determine the beginning and end of the crawling period
startingTime = 0
finishingTime = 0

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
    global startingTime 
    global finishingTime

    print("Panda from Json")
    peerMetrics = getDictFromJson(inputFile)
    fileTime = getModificationTimeOfFile(inputFile)

    ## Temp -to get the Location of the IPs
    cont=0
    ##
    
    # Define the panda 
    pMetrics = {'PeerId': [], 'NodeId':[], 'ClientType':[], 'Pubkey':[], 'Addrs':[], 'Ip':[], 'Country':[], 'City':[], 'Latency':[], 'Connections':[], 'Disconnections':[], 'ConnectedTime':[], 'BeaconBlockCnt':[], 'BeaconAggregateProofCnt':[], 'VoluntaryExitCnt':[], 'ProposerSlashingCnt':[], 'AttesterSlashingCnt':[], 'TotalMessages': []}
    
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
            actualLatency = 0.0

        print("Peer", peer, "is form Client:", actualClientType) 
        print("Peer", peer, "has latency:", actualLatency)
        pMetrics['ClientType'].append(actualClientType)
        pMetrics["Latency"].append(actualLatency)
        pMetrics['Pubkey'].append(peerMetrics[peer]['Pubkey'])
        
        addrsCnt = 0
        print("ip of the peer", peerMetrics[peer]['Ip'])
        print("location: ", peerMetrics[peer]['Country'])
        
        auxcountry  = ""
        auxcity     = ""
        auxaddrs    = ""
        auxip       = ""

        try:
            if peerMetrics[peer]['Country'].lower() == 'unknown' or peerMetrics[peer]['Country'] == "":
                for idx, address in enumerate(peerstoreMetrics[peer]['addrs']):
                    ipx = address.replace('/ip4/', '')
                    ipx = ipx.split('/')[0]
                    ipx=IP(ipx)
                    if ipx.iptype() == 'PUBLIC':
                        country, city = getLocationFromIp(ipx)
                        auxcountry  = country
                        auxcity     = city
                        auxaddrs    = address
                        auxip       = str(ipx)
                        print('(Private from the beggining) Added:', ipx, ipx.iptype(),country, city)
                        ipPublic = 1
                        cont=cont+1
                        break
                    addrsCnt=addrsCnt+1

                if ipPublic == 0:
                    print("Private ip", peerMetrics[peer]['Ip'])
                    auxcountry  = 'Unknown'
                    auxcity     = 'Unknown'
                    auxaddrs    = peerMetrics[peer]['Addrs']
                    auxip       = peerMetrics[peer]['Ip']

            else:
                print("Already good from Rumor")
                auxcountry  = peerMetrics[peer]['Country']
                auxcity     = peerMetrics[peer]['City']
                auxaddrs    = peerMetrics[peer]['Addrs']
                auxip       = peerMetrics[peer]['Ip']

        except:
            print("Unknown")
            auxcountry  = 'Unknown'
            auxcity     = 'Unknown'
            auxaddrs    = peerMetrics[peer]['Addrs']
            auxip       = peerMetrics[peer]['Ip']
 
        pMetrics['Country'].append(auxcountry)
        pMetrics['City'].append(auxcity)
        pMetrics['Addrs'].append(auxaddrs)
        pMetrics['Ip'].append(auxip)

#       try:
#           peerstoreMetrics[peer]['addrs']
#           ip = IP(peerMetrics[peer]['Ip'])
#           if ip.iptype() == 'PRIVATE' or ip.iptype() == 'LOOPBACK' or peerMetrics[peer]['Country'] == '':
#               for idx, address in enumerate(peerMetrics[peer]['Addrs']):
#                   ipx = address.replace('/ip4/', '')
#                   ipx = ipx.split('/')[0]
#                   ipx=IP(ipx)
#                   if ipx.iptype() == 'PUBLIC':
#                       pMetrics['Ip'].append(str(ipx))
#                       country, city = getLocationFromIp(ipx)
#                       pMetrics['Country'].append(country)
#                       pMetrics['City'].append(city)
#                       pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][addrsCnt])
#                       #print('(Private from the beggining) Added:', ipx, ipx.iptype(),country, city)
#                       ipPublic = 1
#                       cont=cont+1
#                       break
#                   addrsCnt=addrsCnt+1

#               if ipPublic == 0:
#                   print(ip, ip.iptype())
#                   pMetrics['Country'].append('Unknown')
#                   pMetrics['City'].append('Unknown')
#                   #print('(Private) Added:', ip, 'Unknown, Unknown')
#                   pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][0])
#           else:           
#               pMetrics['Country'].append(peerMetrics[peer]['Country'])
#               pMetrics['City'].append(peerMetrics[peer]['City'])
#               pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'][0])
#               #print('(Public from the beggining) Added:', ip, ip.iptype(), peerMetrics[peer]['Country'], peerMetrics[peer]['City'])
#       except:
#           pMetrics['Country'].append('Unknown')
#           pMetrics['City'].append('Unknown')
#           pMetrics['Addrs'].append(peerMetrics[peer]['Addrs'])
#           pMetrics['Ip'].append(peerMetrics[peer]['Ip'])

        connection, disconnection, ttime = GetConnectDisconnectAndConTime(peer, peerMetrics, fileTime)
        pMetrics['Connections'].append(connection)
        pMetrics['Disconnections'].append(disconnection)
        pMetrics['ConnectedTime'].append(ttime)
            
        pMetrics['BeaconBlockCnt'].append(peerMetrics[peer]['BeaconBlock']['Cnt'])
        pMetrics['BeaconAggregateProofCnt'].append(peerMetrics[peer]['BeaconAggregateProof']['Cnt'])
        pMetrics['VoluntaryExitCnt'].append(peerMetrics[peer]['VoluntaryExit']['Cnt'])
        pMetrics['ProposerSlashingCnt'].append(peerMetrics[peer]['ProposerSlashing']['Cnt'])
        pMetrics['AttesterSlashingCnt'].append(peerMetrics[peer]['AttesterSlashing']['Cnt'])

        print(peerMetrics[peer]['BeaconBlock']['Cnt'], type(peerMetrics[peer]['BeaconBlock']['Cnt']))

        tmess = int(peerMetrics[peer]['BeaconBlock']['Cnt'] + peerMetrics[peer]['BeaconAggregateProof']['Cnt'] + peerMetrics[peer]['VoluntaryExit']['Cnt'] + peerMetrics[peer]['ProposerSlashing']['Cnt'] + peerMetrics[peer]['AttesterSlashing']['Cnt'])
        pMetrics['TotalMessages'].append(tmess)

        # To dont exeed the limit of petitions per minute
        if cont >= 40:
                time.sleep(70)
                cont=0

    print('len PeerId:', len(pMetrics['PeerId']))
    print('len ClientType:', len(pMetrics['ClientType']))
    print('len Addrs:', len(pMetrics['Addrs']))
    print('len Country:', len(pMetrics['Country']))
    print('len City:', len(pMetrics['City']))
    print('len Latency:', len(pMetrics['Latency']))
    print('len Connections:', len(pMetrics['Connections']))

    pandaObject = pd.DataFrame(pMetrics, columns = ['PeerId', 'NodeId', 'ClientType', 'Pubkey', 'Addrs', 'Country',
     'City', 'Latency', 'Connections', 'Disconnections', 'ConnectedTime', 'BeaconBlockCnt', 'BeaconAggregateProofCnt',
      'VoluntaryExitCnt', 'ProposerSlashingCnt', 'AttesterSlashingCnt', 'TotalMessages'])

    # Get the initial time and the end time of the crawling
    print('initial date:', datetime.datetime.fromtimestamp(startingTime/1000))
    print('final date:  ', datetime.datetime.fromtimestamp(finishingTime/1000))    

    return pandaObject

    # Generate the pandaobject of all the metrics per peer
def getPandaobjectFromPeerstoreJson(inputFile):
    global startingTime 
    global finishingTime

    print("Panda from Json")
    peerstoreMetrics = getDictFromJson(inputFile)

    ## Temp -to get the Location of the IPs
    cont=0
    ##
    
    # Define the panda 
    pMetrics = {'ClientType':[]}
    
    for peer in peerstoreMetrics:
        try:
            print(peerstoreMetrics[peer]["user_agent"])
            actualClientType = peerstoreMetrics[peer]["user_agent"]
        except:
            actualClientType = "Unknown"
            print('Unknown')

        #print("Peer", peer, "is form Client:", actualClientType) 

        pMetrics['ClientType'].append(actualClientType)



    pandaObject = pd.DataFrame(pMetrics, columns = ['ClientType'])

    # Get the initial time and the end time of the crawling
    print('initial date:', datetime.datetime.fromtimestamp(startingTime/1000))
    print('final date:  ', datetime.datetime.fromtimestamp(finishingTime/1000))    

    return pandaObject

      
# request the location from api
def getLocationFromIp(ipAddress):
    composedUrl = f"http://ip-api.com/json/{ipAddress}"
    resp = requests.get(url=composedUrl) 
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
    global startingTime
    global finishingTime

    prevEvent = 'Disconnection'
    prevTime = 0
    timeRange = 500 # milliseconds
    contConn = 0
    contDisc = 0
    ttime = 0
    ctime = 0 # aux variable to calculate the final time

    for connection in peerMetrics[peer]["ConnectionEvents"]:
        if prevEvent != connection["ConnectionType"] or connection['TimeMili'] >= (prevTime + timeRange):                                
            if connection["ConnectionType"] == 'Connection':
                contConn = contConn +1                                  
                ctime = connection["TimeMili"] # millis
                prevEvent = connection["ConnectionType"]
                prevTime = connection['TimeMili']                      
            elif connection["ConnectionType"] == 'Disconnection':
                contDisc = contDisc +1
                ttime = ttime + (connection["TimeMili"] - ctime) #millis 
                ctime = connection["TimeMili"] # millis                                  
                prevEvent = connection["ConnectionType"]                      
                prevTime = connection['TimeMili']
        if startingTime == 0:
            startingTime = connection["TimeMili"]
            finishingTime = connection['TimeMili']  
        else:
            if startingTime > connection['TimeMili']:
                startingTime = connection['TimeMili']
            if finishingTime < connection['TimeMili']:
                finishingTime = connection['TimeMili']

#        if connection["ConnectionType"] == "Connection" :
#            connectionCounter += 1
#            if timeFlag == 0:
#                ctime = connection["TimeMili"] # secs
#                timeFlag = 1
#        if connection["ConnectionType"] == "Disconnection" :
#            disconnectionCounter += 1
#            if timeFlag == 1:
#                connectionTotalTime = connectionTotalTime + (connection["TimeMili"] - ctime)
#                timeFlag = 0
#    # if the flag is 1, means that on the moment of taking the metrics we were connected
#    if timeFlag == 1:
#        connectionTotalTime = connectionTotalTime + (fileTime - ctime)
    return contConn, contDisc, ttime/60000 # from millis to minutes ( /60*1000)
 
########### ------------------ Ploting Stage/Code AKA Wonderland ------------
# TODO: - At one point would be nice to add the ploting stuff on a library itself


def plotBarsFromPandas(panda, opts):
    print("Bar Graph from Panda")

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    if opts['xmetrics'] != None:
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

    # Check id the grid has been set
    if opts['grid'] != None:
        ax.grid(which='major', axis=opts['grid'], linestyle='--')

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

# Sort xarray and y array By Values from Max to Min
def sortArrayMaxtoMin(xarray, yarray):
    iterations = len(xarray)
    x = []
    y = []
    for i in range(iterations):
        maxV   = max(yarray)
        maxIdx = yarray.index(maxV)
        x.append(xarray[maxIdx])
        y.append(maxV)
        print("New line:", xarray[maxIdx], maxV)
        xarray.pop(maxIdx)
        yarray.pop(maxIdx)
    return x, y

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
        aux = plt.get_cmap(item, len(xarray[idx]))
        auxarray = aux(range(len(xarray[idx])))
        if cnt == 0:
            innercolors = auxarray[::-1]
        else:
            innercolors = np.concatenate((innercolors, auxarray[::-1]), axis=0)
        cnt = cnt + 1

    if opts['autopct'] == 'values':
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1-size, colors=opts['outercolors'], labels=opts['outerlabels'], 
                    labeldistance=opts['labeldistance'], autopct=autopct_format(valsouter), pctdistance=opts['pctdistance'],
                     wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))
    elif opts['autopct'] == 'pcts':
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1-size, colors=opts['outercolors'], labels=opts['outerlabels'], 
                    labeldistance=opts['labeldistance'], autopct='%1.1f', pctdistance=opts['pctdistance'], 
                    wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))

    elif opts['autopct'] == False:
        patches1, labels1, autotext = ax.pie(x=valsouter, radius=1-size, colors=opts['outercolors'], labels=None, 
                    labeldistance=opts['labeldistance'], autopct=autopct_format(valsouter), pctdistance=opts['pctdistance'], 
                    wedgeprops=dict(width=size, edgecolor=opts['edgecolor']))

    for idx, _ in enumerate(labels1):
        autotext[idx].set_fontsize(opts['labelsize'])
        autotext[idx].set_c(opts['outercolors'][idx])
        if opts['autopct'] != False:
            labels1[idx].remove()
            #labels1[idx].set_fontsize(opts['labelsize'])
            #labels1[idx].set_c(opts['outercolors'][idx])

    # , labels=opts['innerlabels']
    patches2, labels2 = ax.pie(valsinner, radius=1, colors=innercolors,
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


def plotColumn(panda, opts):

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    fig = plt.figure(figsize = opts['figSize'])
    ax = fig.add_subplot(111)

    # TODO: add the sortting nativelly to the plot function
    if opts['sortmetrics'] != None:
        print('Sorting') 
        sortedPanda = panda.sort_values(by=opts['sortmetrics'], ascending=False)
        if opts['xMetrics'] != None:
            sortedPanda.plot(ax=ax, logx=opts['xlog'], logy=opts['ylog'], x=opts['xMetrics'], y=opts['yMetrics'], style=opts['markerStyle'], marker=opts['marker'], markersize=opts['markerSize'], label=opts['legendLabel'])
        else: 
            print(sortedPanda[opts['sortmetrics']])
            sortedPanda[opts['yMetrics']].sort_values(by=opts['sortmetrics'], ascending=False).plot(ax=ax, logx=opts['xlog'], logy=opts['ylog'], style=opts['markerStyle'], marker=opts['marker'], markersize=opts['markerSize'], label=opts['legendLabel'])
        print('Done')
    else:
        panda.plot(ax=ax, logx=opts['xlog'], logy=opts['ylog'], x=opts['xMetrics'], y=opts['yMetrics'], style=opts['markerStyle'], marker=opts['marker'], markersize=opts['markerSize'], label=opts['legendLabel'])
    
    ax.set_ylabel(opts['yLabel'], fontsize=opts['labelSize'])
    ax.set_xlabel(opts['xLabel'], fontsize=opts['labelSize'])

    ax.tick_params(axis='both', labelsize=opts['tickSize'])
    
    # Check if the legend was enabled
    if opts['legendLabel'] != None:
        # Adding opts['legendSize'] as markerscale might not be the best option, try and see how it looks
        # if it doesn't look nice, change by adding a new flag 
        ax.legend(markerscale=opts['legendSize'], loc=opts['legendPosition'], ncol=ncol, prop={'size':opts['legendSize']})
    else:
        ax.get_legend().remove()
    
    # Set/No the grids if specified
    if opts['hGrids'] != False:
        ax.grid(which='major', axis='y', linestyle='--')
    if opts['vGrids'] != False:
        ax.grid(which='major', axis='x', linestyle='--')

    # Check if any limit was set for the x axis 
    if opts['xLowLimit'] != None and opts['xUpperLimit'] != None: # For X axis
        print("Both X limits set")
        ax.xaxis.set_ticks(np.arange(opts['xLowLimit'], opts['xUpperLimit'], opts['xRange']))
        ax.set_xlim(left=opts['xLowLimit'], right=opts['xUpperLimit'])
    elif opts['xLowLimit'] != None:
        print("Only xLow limit set")
        ax.xaxis.set_ticks(np.arange(opts['xLowLimit'], panda[opts['xMetrics']].iloc[-1]+1, opts['xRange']))
        ax.set_xlim(left=opts['xLowLimit'], right=panda[opts['xMetrics']].iloc[-1]+1)
    elif opts['xUpperLimit'] != None:
        print("Only xUpper limit set")
        ax.xaxis.set_ticks(np.arange(0, opts['xUpperLimit'], opts['xRange']))
        ax.set_xlim(left=0, right=opts['xUpperLimit'])
    else:
        print("Non xLimit set") 
        ax.xaxis.set_ticks(np.arange(0, panda[opts['xMetrics']].iloc[-1]+1, opts['xRange']))
        ax.set_xlim(left=0, right=panda[opts['xMetrics']].iloc[-1]+1)

    if opts['yLowLimit'] != None:
        ax.set_ylim(bottom=opts['yLowLimit'])
    if opts['yUpperLimit'] != None:
        ax.set_ylim(top=opts['yUpperLimit'])
    #if opts['yRange'] != None:

    if opts['xticks'] == None:
        ax.get_xaxis().set_ticks([])

    # Set horizontal and vertical lines if needed
    if opts['hlines'] != None:
        for item in opts['hlines']:
            plt.axhline(y=item, color=opts['hlineColor'], linestyle=opts['hlineStyle'])
    if opts['vlines'] != None:
        for item in opts['vlines']:
            plt.axvline(x=item, color=opts['vlineColor'], linestyle=opts['vlineStyle'])
    plt.title(opts['title'], fontsize=opts['titleSize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    #plt.show()


# Funtion that gives length of the panda
def getLengthOfPanda(panda):
    return len(panda)

# Function that gets the data (counter, sum, avg) of the given metric from the panda
def getDataFromPanda(panda, ymetrics, xmetrics, xarray, flag):
    yarray = []
    print(xarray)
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
                    auxCnt = auxCnt + float(row[ymetrics])
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
        peerstorePanda = getPandaobjectFromPeerstoreJson(peerstoreFile)

        rumorMetricsPanda = pd.read_csv(rumorMetricsFile)


        # plot just the peerstore
        xarray, yarray = getDataFromPanda(peerstorePanda, None, "ClientType", ['ligh', 'teku', 'nim', 'prysm', 'lod', 'Unknown'], 'counter')

        plotBarsFromArrays(xarray, yarray, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'PeerstoreClientType.png',                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'barColor': ['tab:blue', 'tab:orange', 'tab:green', 'tab:red', 'tab:purple', 'k' ],
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Client Types on the Peerstore",                             
            'xlabel': ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown'],                                   
            'ylabel': 'Number of peers',                                                
            'xticks': xarray,                                                           
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                           
            'tickRotation': 0,
            'show': False}) 



    
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
        'title': "Number of Peers Connected from the entire Peerstore",                  
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

    lightarray = []
    tekuarray = []
    nimbusarray = []
    prysmarray = []
    lodesarray = []
    unknarray = []
    # iterate through the items on the received names
    for idx, item in enumerate(clientVersList):
        if 'ligh' in item.lower():
            aux = item.rsplit('/', 1)[0]
            aux = aux.rsplit('-', 1)[0]
            try: 
                i = lightarray.index(aux); 
            except ValueError: 
                lightarray.append(aux)
        if 'teku' in item.lower():
            aux = item.rsplit('/', 2)[0]
            aux = aux.rsplit('+', 1)[0]
            try: 
                i = tekuarray.index(aux); 
            except ValueError: 
                tekuarray.append(aux)
        if 'nim' in item.lower():
            aux = item.rsplit('/', 1)[0]
            aux = aux.rsplit('+', 1)[0]
            try: 
                i = nimbusarray.index(aux); 
            except ValueError: 
                nimbusarray.append(aux)

        if 'pry' in item.lower():
            aux = item.rsplit('/', 1)[0]
            aux = aux.rsplit('+', 1)[0]
            try: 
                i = prysmarray.index(aux); 
            except ValueError: 
                prysmarray.append(aux)
        if 'lod' in item.lower():
            aux = item.rsplit('/', 1)[0]
            aux = aux.rsplit('+', 1)[0]
            try: 
                i = lodesarray.index(aux); 
            except ValueError: 
                lodesarray.append(aux)
        if 'unkn' in item.lower():
            aux = item.rsplit('/', 1)[0]
            aux = aux.rsplit('+', 1)[0]
            try: 
                i = unknarray.index(aux.lower())
                
            except ValueError: 
                try:
                    p = prysmarray.index(aux)
                except:
                    unknarray.append(aux.lower())

    print('light:', len(lightarray))
    print('Teku:', len(tekuarray))
    print('Nimbus:', len(nimbusarray))
    print('Prysm', len(prysmarray))
    print('Lodestar', len(lodesarray))
    print('Unknown', len(unknarray))

    clientVersList = lightarray + tekuarray + nimbusarray + prysmarray + lodesarray + unknarray
    print(clientVersList)


    xarray, yarray = getDataFromPanda(rumorMetricsPanda, None, "ClientType", clientVersList, 'counter')
    namesarray, valuesarray = sortArrayByNames(xarray, yarray, clientList)

    plotDoublePieFromArray(valuesarray, opts={                                   
        'figsize': figSize,                                                      
        'figtitle': 'PeersPerClient.png',                                    
        'outputpath': outputFigsFolder,
        'piesize': 0.3,                                                      
        'autopct': "pcts", #False,
        'pctdistance': 1.65,
        'edgecolor': 'w',
        'innerlabels': clientVersList,
        'outerlabels': clientList,
        'labeldistance': 1.25,
        'innercolors': innerColors,
        'outercolors': clientColors,
        'shadow': None,
        'startangle': 90,                                                  
        'title': "Number of Peers From Each Client and Their Versions",                   
        'titlesize': titleSize,                                                        
        'labelsize': labelSize, 
        'legend': True,                                                       
        'lengendposition': None,                                                   
        'legendsize': labelSize,                                                     
        'show': False})
    
    print("{:<30} {:<15}".format('ClientVersion', 'NumbersPeers'))
    for idx, iten in enumerate(xarray):
        print("{:<30} {:<15}".format(xarray[idx], yarray[idx]))



    # get the number of peers per country 
    countriesList = getItemsFromColumn(rumorMetricsPanda, 'Country') 
    auxxarray, auxyarray = getDataFromPanda(rumorMetricsPanda, None, "Country", countriesList, 'counter') 
    # Remove the Countries with less than X peers
    countryLimit = 10
    xarray = []
    yarray = []
    for idx, item in enumerate(auxyarray):
        if auxyarray[idx] >= countryLimit:
            yarray.append(item)
            xarray.append(auxxarray[idx])
    
    print("X before ->", xarray)                                                       
    print("Y before ->", yarray) 
    xarray, yarray = sortArrayMaxtoMin(xarray, yarray)
    # Get Color Grid
    print("X ->", xarray)
    print("Y ->", yarray)
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
        'title': "Number of Peers Connected from each Country",                             
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
        'title': "Number of Peers From Each Client and Their Versions",                   
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

    # get the average latency per client
    # since few of the clients dont hace latency
    # the calculus are made by hand
    xarray = clientList
    yarray = []
    xmetrics = 'ClientType'
    ymetrics = 'Latency'
    for _, item in enumerate(xarray):            
        auxCnt = 0
        for index, row in rumorMetricsPanda.iterrows():
            if item.lower() in str(row[xmetrics]).lower():
                if row[ymetrics] != 0:
                    auxCnt = auxCnt + float(row[ymetrics])
        auxAmount = rumorMetricsPanda.apply(lambda x: True if item.lower() in str(x[xmetrics]).lower() else False, axis=1)
        if auxCnt != 0:
            yarray.append(round((auxCnt/(len(auxAmount[auxAmount == True].index))),1))
        else:
            yarray.append(0)

    print(xarray)
    print(yarray)

    plotBarsFromArrays(xarray, yarray, opts={                                            
        'figSize': figSize,                                                          
        'figTitle': 'AverageLatencyPerClientType.png',                                
        'outputPath': outputFigsFolder,                                                    
        'align': 'center', 
        'barValues': True,
        'barColor': clientColors,
        'textSize': textSize,                                                         
        'yLowLimit': 0,                                                             
        'yUpperLimit': None,                                                        
        'title': "Average Latency per Client Type",                             
        'xlabel': None,                                   
        'ylabel': 'latency (seconds)',                                                
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
        'title': "Number of Received BeaconBlock Msgs",                             
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
        'title': "Average of Received BeaconBlock Msgs",                             
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
        'title': "Number of Received BeaconAggregateAndProof Msgs",                             
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
        'title': "Average of Received BeaconAggregateAndProof Msgs",                             
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
        'title': "Average of Received ProposerSlashing Messages",                             
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
        'grid': None,                                                   
        'title': "Number of Connections with each Peer",                  
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
        'grid': None,                                                
        'title': "Number of Disconnections with each Peer",                  
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
        'grid': None,                                                    
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
        'yUpperLimit': 5,
        'grid': 'y',                                                    
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

    # Message distributions among the peers

    barColor = 'black'
    messagesDics = {}
    for index, row in rumorMetricsPanda.iterrows():
        if row['BeaconBlockCnt'] in messagesDics:
            messagesDics[row['BeaconBlockCnt']] = messagesDics[row['BeaconBlockCnt']] + 1
        else:
            messagesDics[row['BeaconBlockCnt']] = 1

    #print(messagesDics)

    sortedDict = collections.OrderedDict(sorted(messagesDics.items()))

    print(sortedDict)

    xarray = []
    yarray = []
    for item in sortedDict:
        xarray.append(item)
        yarray.append(sortedDict[item])

    print(xarray)
    print(yarray)

#    plotBarsFromPandas(rumorMetricsPanda, opts={                                   
#        'figSize': wideFigSize,                                                      
#        'figTitle': 'BeaconBlockMessagePerClient.png',                                    
#        'outputPath': outputFigsFolder,
#        'legend': False,                                         
#        'align': 'center',
#        'ylog': True,
#        'xmetrics': ['BeaconBlockCnt'],                                                      
#        'barValues': None,   
#        'barColor': barColor,                                               
#        'yLowLimit': 0,                                                         
#        'yUpperLimit': None,
#        'grid': 'y',                                                    
#        'title': "Number of Beacon Blocks Received from each Peer",                  
#        'xlabel': "Peers Connected",                                                         
#        'ylabel': 'Number of Messages Sent',                                      
#        'xticks': None,                                                       
#        'titleSize': titleSize +2,                                                        
#        'labelSize': labelSize +2,                                                        
#        'lengendPosition': 1,                                                   
#        'legendSize': labelSize +2,                                                       
#        'xticksSize': ticksSize +2,                                                       
#        'yticksSize': ticksSize +2,                                                     
#        'tickRotation': 0,                                                     
#        'show': False}) 

    print(rumorMetricsPanda.loc[rumorMetricsPanda['BeaconBlockCnt'].idxmax()])

    print(rumorMetricsPanda['BeaconBlockCnt'])

    plotColumn(rumorMetricsPanda, opts={
        'figSize': wideFigSize, 
        'figTitle': 'BeaconBlockMessagePerClient.png',
        'outputPath': outputFigsFolder,
        'xlog': False,
        'ylog': True,
        'xMetrics': None,
        'yMetrics': ['BeaconBlockCnt'],
        'sortmetrics': 'BeaconBlockCnt',
        'xticks': None,
        'xLowLimit': 0,
        'xUpperLimit': len(rumorMetricsPanda),
        'xRange': 1,
        'yLowLimit': 10**0,
        'yRange': None,
        'yUpperLimit': None,
        'title': "Number of Beacon Blocks Received from each Peer",
        'xLabel': "Peers Connected",
        'yLabel': 'Number of Messages Received',
        'legendLabel': None,
        'titleSize': titleSize +2,
        'labelSize': labelSize + 2,
        'lableColor': 'tab:orange',
        'hGrids': True,
        'vGrids': False,
        'hlines': [1000],
        'vlines': None,
        'hlineColor': 'r',
        'vlineColor': 'r',
        'hlineStyle': None,
        'vlineStyle': '--',
        'marker': '.',
        'markerStyle': ',',
        'markerSize': 4,
        'lengendPosition': 1,
        'legendSize': 16,
        'tickSize': 16})



#    plotBarsFromArrays(xarray, yarray, opts={                                            
#        'figSize': figSize,                                                          
#        'figTitle': 'BeaconBlockMessageDistribution.png',                                
#        'outputPath': outputFigsFolder,                                                    
#        'align': 'center', 
#        'barValues': None,
#        'barColor': barColor,
#        'textSize': textSize,                                                         
#        'yLowLimit': 0,                                                             
#        'yUpperLimit': None,                                                        
#        'title': "Beacon Block Messages Distribution among the Peers",                             
#        'xlabel': "Number of Messages Sent ",                                   
#        'ylabel': 'Number of Peers',                                                
#        'xticks': xarray,                                                           
#        'titleSize': titleSize,                                                        
#        'labelSize': labelSize,                                                        
#        'lengendPosition': 1,                                                   
#        'legendSize': labelSize,                                                       
#        'xticksSize': ticksSize-2,                                                       
#        'yticksSize': ticksSize,                                                             
#        'tickRotation': 90,
#        'show': False}) 

    barColor = 'black'

    auxPanda = rumorMetricsPanda.sort_values(by='BeaconBlockCnt', ascending=True)
    cont = 0
#    for index, row in auxPanda.iterrows():
#        print(row['BeaconBlockCnt'])
            


    auxrow = rumorMetricsPanda.loc[rumorMetricsPanda['ConnectedTime'].idxmax()]
    maxX = auxrow['ConnectedTime'] 

    plotColumn(rumorMetricsPanda, opts={
        'figSize': wideFigSize, 
        'figTitle': 'TotalMesagesPerTimeConnected.png',
        'outputPath': outputFigsFolder,
        'xlog': False,
        'ylog': True,
        'xMetrics': 'ConnectedTime',
        'yMetrics': ['TotalMessages'],
        'sortmetrics': None,
        'xticks': 1,
        'xLowLimit': 0,
        'xUpperLimit': maxX,
        'xRange': 250,
        'yLowLimit': 10**0,
        'yRange': None,
        'yUpperLimit': None,
        'title': "Total of Messages for Connected Time",
        'xLabel': "Connected Time (Minutes)",
        'yLabel': 'Number of Messages Received',
        'legendLabel': None,
        'titleSize': titleSize +2,
        'labelSize': labelSize + 2,
        'lableColor': 'tab:orange',
        'hGrids': True,
        'vGrids': True,
        'hlines': None,
        'vlines': None,
        'hlineColor': None,
        'vlineColor': 'r',
        'hlineStyle': None,
        'vlineStyle': '--',
        'marker': '.',
        'markerStyle': ',',
        'markerSize': 4,
        'lengendPosition': 1,
        'legendSize': 16,
        'tickSize': 16})




    # ------ End of Get data -------

    print("Finished!, Ciao!")



if __name__ == '__main__':
    main()
