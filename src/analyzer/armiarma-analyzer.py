# Armiarma Metrics Analyzer
# Script that ussing the rumor-metrics.py package generates the plots

import os, sys
import json 
import time
import math
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.colors as mcolors
from matplotlib.backends.backend_pdf import PdfPages
from pathlib import Path
import numpy as np
import datetime
import collections

def getDictFromJson(inputFile):
    print("reading json: ", inputFile)
    mf = open(inputFile)
    dicts = json.load(mf)
    mf.close()

    return dicts

# Generate the pandaobject of all the metrics per peer
def getPandaFromPeerstoreJson(inputFile):
    global startingTime 
    global finishingTime

    print("Panda from Json")
    peerstoreMetrics = getDictFromJson(inputFile)

    return peerstoreMetrics
      

def plotFromPandas(panda, pdf, opts):
    print("Bar Graph from Panda")

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    if opts['xmetrics'] != None:
        if opts['kind'] != None:
            ax = panda[opts['xmetrics']].sort_values(by=opts['xmetrics'], ascending=False).plot(kind=opts['kind'], figsize=opts['figSize'], logy=opts['ylog'], legend=opts['legend'], color=opts['barColor'], use_index=False) 
        else:
            ax = panda[opts['xmetrics']].sort_values(by=opts['xmetrics'], ascending=False).plot(figsize=opts['figSize'], logy=opts['ylog'], lw=opts['lw'], legend=opts['legend'], color=opts['barColor'], use_index=False)
    
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
    pdf.savefig(plt.gcf())
    if opts['show'] is True:
        plt.show()



# CortzePlot extension to plot bar-charts
def plotBarsFromArrays(xarray, yarray, pdf, opts):
    print("Bar Graph from Arrays")

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    fig = plt.figure(figsize = opts['figSize'])

    plt.bar(range(len(xarray)), yarray, log=opts['logy'], align=opts['align'], color=opts['barColor'])


    # labels
    if opts['ylabel'] is not None:    
        plt.ylabel(opts['ylabel'], fontsize=opts['labelSize'])
    if opts['xlabel'] is not None:
        plt.xlabel(opts['xlabel'], fontsize=opts['labelSize'])

    # Ticks LABELS
    if opts['xticks'] is not None:
        plt.xticks(range(len(xarray)), opts['xticks'], rotation=opts['tickRotation'], fontsize=opts['xticksSize'])
    else:
        plt.xticks(range(len(xarray)), {},)
        #plt.xticks(range(len(xarray)))

    plt.margins(x=0)
    plt.yticks(fontsize=opts['yticksSize'])
    plt.ylim(opts['yLowLimit'], opts['yUpperLimit'])

    # Set/No the grids if specified
    if opts['hGrids'] != False:
        plt.grid(which='major', axis='y', linestyle='--')
    
    # Check is there is Value on top of the charts
    if opts['barValues'] is not None:
        for ind, value in enumerate(yarray):
            plt.text(ind, value, str(value), fontsize=opts['textSize'], horizontalalignment='center')

    # Title
    plt.title(opts['title'], fontsize = opts['titleSize'])
    plt.tight_layout()
    plt.savefig(outputFile)
    pdf.savefig(plt.gcf())
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

        xarray.pop(maxIdx)
        yarray.pop(maxIdx)
    return x, y

# Funtion that plots the given array into a pie chart 
def plotSinglePieFromArray(xarray, pdf, opts):
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
    pdf.savefig(plt.gcf())
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
def plotDoublePieFromArray(xarray, pdf, opts):
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
    pdf.savefig(plt.gcf())
    if opts['show'] is True:
        plt.show()


def plotColumn(panda, pdf, opts):

    outputFile = str(opts['outputPath']) + '/' + opts['figTitle']
    print('printing image', opts['figTitle'], 'on', outputFile)

    fig = plt.figure(figsize = opts['figSize'])
    ax = fig.add_subplot(111)

    # TODO: add the sortting nativelly to the plot function
    if opts['sortmetrics'] != None:
        sortedPanda = panda.sort_values(by=opts['sortmetrics'], ascending=False)
        if opts['xMetrics'] != None:
            sortedPanda.plot(ax=ax, logx=opts['xlog'], logy=opts['ylog'], x=opts['xMetrics'], y=opts['yMetrics'], style=opts['markerStyle'], marker=opts['marker'], markersize=opts['markerSize'], label=opts['legendLabel'])
        else: 
            sortedPanda[opts['yMetrics']].sort_values(by=opts['sortmetrics'], ascending=False).plot(ax=ax, logx=opts['xlog'], logy=opts['ylog'], style=opts['markerStyle'], marker=opts['marker'], markersize=opts['markerSize'], label=opts['legendLabel'])
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
        #ax.xaxis.set_ticks(np.arange(0, panda[opts['xMetrics']].iloc[-1]+1, opts['xRange']))
        #ax.set_xlim(left=0, right=panda[opts['xMetrics']].iloc[-1]+1)

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
    pdf.savefig(plt.gcf())
    #plt.show()


# Funtion that gives length of the panda
def getLengthOfPanda(panda):
    return len(panda)

def getTypesPerName(panda, c1name, column1, column2):
    types = []
    typeCounter = []
    totalCounter = 0
    for index, row in panda.iterrows():
        if c1name.lower() == str(row[column1]).lower():
            totalCounter = totalCounter + 1
            # Check if the version is already in types
            if str(row[column2]) not in types:
                types.append(str(row[column2]))
                typeCounter.append(1)
            else:
                idx = types.index(str(row[column2]))
                typeCounter[idx] = typeCounter[idx] + 1
    if not typeCounter:
        typeCounter.append(0)
    return totalCounter, types, typeCounter

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
    csvFile = sys.argv[1]
    peerstoreFile = sys.argv[2]
    outputFigsFolder = sys.argv[3]

    pdfFile = outputFigsFolder + "/MetricsSummary.pdf"
    
    peerstorePanda = getPandaFromPeerstoreJson(peerstoreFile)
    rumorMetricsPanda = pd.read_csv(csvFile)



    # ---------- PLOT SECUENCE -----------

    # ------ Get data for plotting -------
    clientList = ['Lighthouse', 'Teku', 'Nimbus', 'Prysm', 'Lodestar', 'Unknown']
    clientColors = ['tab:blue', 'tab:orange', 'tab:green', 'tab:red', 'tab:purple', 'k' ]
    innerColors = ['Blues', 'Oranges', 'Greens', 'Reds', 'Purples', 'Greys' ]

    # get length of the peerstore
    peerstoreSize = getLengthOfPanda(peerstorePanda)
    peerMetricsSize = getLengthOfPanda(rumorMetricsPanda)

    xarray = ['Peerstore', 'Connected Peers']
    yarray = [peerstoreSize, peerMetricsSize]
    barColor = ['tab:blue', 'tab:green']

    # Temporary measurments for the Armiarma Paper

    # Count number of Prysm Peers on the Peerstore
    ff = open(peerstoreFile)
    peerstore = json.load(ff)

    cntU = 0
    cntN = 0
    
    print()
    print('Total amount of peers on the peerstore:',len(peerstore))
    for peer in peerstore:
        try:
            if '/13000/' in peerstore[peer]["addrs"][0]:
                cntU = cntU +1
        except:
            pass

    print('Number of clients with the TPC port at 1300:', cntU)
    print()
    print('percentage of "Prysm" peers from the peerstore:', (cntU*100)/len(peerstore))

    ff.close()

    # get percentage of the total of messages received

    totalMessageCounter = 0
    messageCounterArray = []
    messagePercentageArray = []

    for index, row in rumorMetricsPanda.iterrows():
        totalMessageCounter = totalMessageCounter + row['Total Messages']
        messageCounterArray.append(row['Total Messages'])

    for msgCount in messageCounterArray:
        percent = msgCount/totalMessageCounter
        messagePercentageArray.append(percent)

    messageCounterArray = sorted(messageCounterArray, reverse = True)
    messagePercentageArray = sorted(messagePercentageArray, reverse = True)

    print(messageCounterArray[:5])
    print("Total of received messages:", totalMessageCounter)

    # Total of messages
    percAcumulated = 0
    itemsCounter   = 0
    for item in messagePercentageArray:
        itemsCounter = itemsCounter + 1
        percAcumulated = percAcumulated + item
        if itemsCounter > (len(messagePercentageArray) * 0.1): # to see how many messages send the 10% of the peers 
            peerPerc = itemsCounter / len(messagePercentageArray)
            print( peerPerc*100, "% of peers send the ", percAcumulated*100,"% of the messages")
            restPeers = len(messagePercentageArray) - itemsCounter
            print((restPeers/len(messagePercentageArray))*100, "% of the peers send less than", item*totalMessageCounter, "number of messages")
            break

    percAcumulated = 0
    itemsCounter   = 0
    for item in messagePercentageArray:
        itemsCounter = itemsCounter + 1
        percAcumulated = percAcumulated + item
        if percAcumulated > 0.89:
            peerPerc = itemsCounter / len(messagePercentageArray)
            print(peerPerc, "% of peers send the ", percAcumulated*100,"% of the messages")
            restPeers = len(messagePercentageArray) - itemsCounter
            print((restPeers/len(messagePercentageArray))*100, "% of the peers send less than", item*totalMessageCounter, "number of messages")
            break
    print()
    print()

    # Beacon blocks
    totalBlockCounter = 0
    messageCounterArray = []
    messagePercentageArray = []

    for index, row in rumorMetricsPanda.iterrows():
        totalBlockCounter = totalBlockCounter + row['Beacon Blocks']
        messageCounterArray.append(row['Beacon Blocks'])

    for msgCount in messageCounterArray:
        percent = msgCount/totalBlockCounter
        messagePercentageArray.append(percent)

    messageCounterArray = sorted(messageCounterArray, reverse = True)
    messagePercentageArray = sorted(messagePercentageArray, reverse = True)

    print(messageCounterArray[:5])
    print("Total of received blocks:", totalBlockCounter)

    percAcumulated = 0
    itemsCounter   = 0
    for item in messagePercentageArray:
        itemsCounter = itemsCounter + 1
        percAcumulated = percAcumulated + item
        if itemsCounter > (len(messagePercentageArray) * 0.1): # to see how many messages send the 10% of the peers 
            peerPerc = itemsCounter / len(messagePercentageArray)
            print( peerPerc*100, "% of peers send the ", percAcumulated*100,"% of the blocks")
            restPeers = len(messagePercentageArray) - itemsCounter
            print((restPeers/len(messagePercentageArray))*100, "% of the peers send less than", item*totalBlockCounter, "number of blocks")
            break

    percAcumulated = 0
    itemsCounter   = 0
    for item in messagePercentageArray:
        itemsCounter = itemsCounter + 1
        percAcumulated = percAcumulated + item
        if percAcumulated > 0.89:
            peerPerc = itemsCounter / len(messagePercentageArray)
            print(peerPerc, "% of peers send the ", percAcumulated*100,"% of the blocks")
            restPeers = len(messagePercentageArray) - itemsCounter
            print((restPeers/len(messagePercentageArray))*100, "% of the peers send less than", item*totalBlockCounter, "number of blocks")
            break
    print()
    print()
    # End of Temporary code for the paper

    # Temp
    # Check if any of the Lodestar nodes was recognized with the  "js-libp2p/version" user agent
    for index, row in rumorMetricsPanda.iterrows():
        if row['Client'] == 'Unknown':
            try:
                if "js-libp2p" in peerstore[row['Peer Id']]["user_agent"]:
                    print("Lodestar", peerstore[row['Peer Id']]["user_agent"])
                    ua = peerstore[row['Peer Id']]["user_agent"].split('/')
                    rumorMetricsPanda.at[index,'Client'] = "Lodestar"
                    rumorMetricsPanda.at[index,'Version'] = ua[1]
            except:   
                print("*") 

    # End Temp

    with PdfPages(pdfFile) as pdf:
    
        plotBarsFromArrays(xarray, yarray, pdf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'PeerstoreVsConnectedPeers.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                         
            'align': 'center',                                                      
            'barValues': True,
            'logy': False,
            'barColor': barColor,
            'textSize': textSize,
            'yLowLimit': 0,                                                         
            'yUpperLimit': None,                                                    
            'title': "Number of Peers Connected from the entire Peerstore",                  
            'xlabel': None,                                                         
            'ylabel': 'Number of Peers',                                      
            'xticks': xarray,
            'hGrids': False,                        
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                       
            'tickRotation': 0,                                                     
            'show': False})

        clientCounter = []
        types         = []
        typesCounter  = []

        for idx, item in enumerate(clientList):
            tcnt, tp, tpc = getTypesPerName(rumorMetricsPanda, item, 'Client', 'Version')
            clientCounter.append(tcnt)
            types.append(tp)
            typesCounter.append(tpc)

        xarray = types
        yarray = typesCounter
        namesarray = clientList

        plotDoublePieFromArray(yarray, pdf, opts={                                   
            'figsize': figSize,                                                      
            'figtitle': 'PeersPerClient.png',  
            'pdf': pdfFile,                                   
            'outputpath': outputFigsFolder,
            'piesize': 0.3,                                                      
            'autopct': "pcts", #False,
            'pctdistance': 1.65,
            'edgecolor': 'w',
            'innerlabels': types,
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

        print("| {:<35}| {:<15}|".format('ClientVersion', 'NumbersPeers'))
        print("-------------------------------------------------------")
        for idx, item in enumerate(clientList):
            print("| {:<35}| {:<15}|".format(item, clientCounter[idx]))
            print("-------------------------------------------------------")
            v = types[idx]
            for j, n in enumerate(v):
                print(" -> {:<33}| {:<15}|".format(v[j], yarray[idx][j]))
            print("-------------------------------------------------------")



        # get the number of peers per country 
        countriesList = getItemsFromColumn(rumorMetricsPanda, 'Country') 
        auxxarray, auxyarray = getDataFromPanda(rumorMetricsPanda, None, "Country", countriesList, 'counter') 
        print("Number of different Countries hosting Eth2 clients:", len(auxxarray))
        # Remove the Countries with less than X peers
        countryLimit = 10
        xarray = []
        yarray = []
        for idx, item in enumerate(auxyarray):
            if auxyarray[idx] >= countryLimit:
                yarray.append(item)
                xarray.append(auxxarray[idx])

        xarray, yarray = sortArrayMaxtoMin(xarray, yarray)
        # Get Color Grid
        barColor = GetColorGridFromArray(yarray)


        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': (12,7),                                                          
            'figTitle': 'PeersPerCountries.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'logy': False,
            'barColor': barColor,
            'textSize': textSize+2,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Number of Peers Connected from each Country",                             
            'xlabel': None,                                   
            'ylabel': 'Number of Connections',                                                
            'xticks': xarray, 
            'hGrids': False,                                                           
            'titleSize': titleSize+2,                                                        
            'labelSize': labelSize+2,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+2,                                                       
            'xticksSize': ticksSize-2,                                                       
            'yticksSize': ticksSize+2,                                                            
            'tickRotation': 90,
            'show': False})   

        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Connections", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'AverageOfConnectionsPerClientType.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True, #True
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize+3,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Connections per Client Type",                             
            'xlabel': None,                                   
            'ylabel': 'Number of Connections',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], #xarray,
            'hGrids': True,                                                      
            'titleSize': titleSize+3,                                                        
            'labelSize': labelSize+3,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+3,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize+1,                                                           
            'tickRotation': 0,
            'show': False}) 

        # get the average of disconnections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Disconnections", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'AverageOfDisconnectionsPerClientType.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True, #True
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize+3,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Disconnections per Client Type",                             
            'xlabel': None,                                   
            'ylabel': 'Number of Disconnections',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], #xarray,
            'hGrids': True,                                                      
            'titleSize': titleSize+3,                                                        
            'labelSize': labelSize+3,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+3,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize+1,                                                           
            'tickRotation': 0,
            'show': False})  

        # get the average of ConnectedTime per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Connected Time", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'AverageOfConnectedTimePerClientType.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True, #True
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize+3,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Connected Time per Client Type",                             
            'xlabel': None,                                   
            'ylabel': 'Time (Minutes)',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], #xarray,
            'hGrids': True,                                                      
            'titleSize': titleSize+3,                                                        
            'labelSize': labelSize+3,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+3,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize+1,                                                           
            'tickRotation': 0,
            'show': False})  

        # get the average RTT per client
        # since few of the clients dont have RTT measures
        # the calculus are made by hand
        xarray = clientList
        yarray = []
        latAuxArray = []
        xmetrics = 'Client'
        ymetrics = 'Latency'
        contador = 0
        prysmCnt = 0
        prysmTCnt = 0
        for _, item in enumerate(xarray):            
            auxCnt = 0
            for index, row in rumorMetricsPanda.iterrows():
                if item.lower() in str(row[xmetrics]).lower():
                    if row[ymetrics] != 0:
                        auxCnt = auxCnt + float(row[ymetrics])
                        if row[ymetrics] > 1:
                            latAuxArray.append(row[ymetrics])
                            if "prysm" in item.lower():
                                prysmTCnt = prysmTCnt + 1 
                    else:
                        if "prysm" in item.lower():
                            prysmCnt = prysmCnt + 1    
                        contador = contador + 1
                        # print("client type:", row[xmetrics], "has RTT", row[ymetrics] )
            auxAmount = rumorMetricsPanda.apply(lambda x: True if item.lower() in str(x[xmetrics]).lower() else False, axis=1)
            if auxCnt != 0:
                yarray.append(round((auxCnt/(len(auxAmount[auxAmount == True].index))),1))
            else:
                yarray.append(0)
        """
        print("len of the connected peers:", len(rumorMetricsPanda))
        print("Number of peers with 0 RTT:", contador)
        print("Number of Prysm clients with 0 RTT:", prysmCnt )
        print("Number of Prysm clients with lat != 0:", prysmTCnt)
        print("number of peers with more than 1 sec of RTT", latAuxArray)
        """

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'AverageRTTPerClientType.png',
            'pdf': pdfFile,                                 
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True, 
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize+3,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average RTT per Client Type",                             
            'xlabel': None,                                   
            'ylabel': 'RTT (seconds)',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],  
            'hGrids': True,                                                          
            'titleSize': titleSize+3,                                                        
            'labelSize': labelSize+3,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+3,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize+3,                                                           
            'tickRotation': 0,
            'show': False}) 


        # GossipSub
        # BeaconBlock
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Beacon Blocks", "Client", clientList, 'sum') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessagesFromBeaconBlock.png',
            'pdf': pdfFile,                                 
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Number of Received BeaconBlock Msgs",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                            
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                            
            'tickRotation': 0,
            'show': False})   


        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Beacon Blocks", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessageAverageFromBeaconBlock.png',
            'pdf': pdfFile,                                 
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'logy': True,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Received BeaconBlock Msgs",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], 
            'hGrids': True,                                                           
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                         
            'tickRotation': 0,
            'show': False}) 

        # BeaconAggregateAndProof
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Beacon Aggregations", "Client", clientList, 'sum') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessagesFromBeaconAggregateProof.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Number of Received BeaconAggregateAndProof Msgs",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received (10^6)',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                            
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                          
            'tickRotation': 0,
            'show': False})   


        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Beacon Aggregations", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessageAverageFromBeaconAggregateProof.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': True,
            'logy': True,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Received BeaconAggregateAndProof Msgs",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                           
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                        
            'tickRotation': 0,
            'show': False}) 

        # VoluntaryExit
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Voluntary Exits", "Client", clientList, 'sum') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessagesFromVoluntaryExit.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Number of Received VoluntaryExit Messages from Clients",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                            
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                         
            'tickRotation': 0,
            'show': False})   


        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Voluntary Exits", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessageAverageFromVoluntaryExit.png',
            'pdf': pdfFile,                                 
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Received VoluntaryExit Messages from Clients",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                            
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                       
            'tickRotation': 0,
            'show': False}) 

        # AttesterSlashing
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Attester Slashings", "Client", clientList, 'sum') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessagesFromAttesterSlashing.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Number of Received AttesterSlashing Messages from Clients",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], 
            'hGrids': True,                                                           
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                     
            'tickRotation': 0,
            'show': False})   


        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Attester Slashings", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessageAverageFromAttesterSlashing.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Received AttesterSlashing Messages from Clients",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'],
            'hGrids': True,                                                            
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                           
            'tickRotation': 0,
            'show': False}) 

        # ProposerSlashing
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Proposer Slashings", "Client", clientList, 'sum') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessagesFromProposerSlashing.png', 
            'pdf': pdfFile,                                
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None, 
            'logy': False,                                                       
            'title': "Number of Received ProposerSlashing Messages from Clients",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], 
            'hGrids': True,                                                           
            'titleSize': titleSize,                                                        
            'labelSize': labelSize,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize,                                                       
            'xticksSize': ticksSize,                                                       
            'yticksSize': ticksSize,                                                           
            'tickRotation': 0,
            'show': False})   


        # get the average of connections per client
        xarray, yarray = getDataFromPanda(rumorMetricsPanda, "Proposer Slashings", "Client", clientList, 'avg') 

        plotBarsFromArrays(xarray, yarray, pdf, opts={                                            
            'figSize': figSize,                                                          
            'figTitle': 'MessageAverageFromProposerSlashing.png',
            'pdf': pdfFile,                                 
            'outputPath': outputFigsFolder,                                                    
            'align': 'center', 
            'barValues': None,
            'logy': False,
            'barColor': clientColors,
            'textSize': textSize,                                                         
            'yLowLimit': 0,                                                             
            'yUpperLimit': None,                                                        
            'title': "Average of Received ProposerSlashing Messages",                             
            'xlabel': None,                                   
            'ylabel': 'Messages Received',                                                
            'xticks': ['Lig','Tek','Nim','Pry','Lod','Unk'], 
            'hGrids': True,                                                           
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
        plotFromPandas(rumorMetricsPanda, pdf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'ConnectionsWithPeers.png',
            'pdf': pdfFile,                                     
            'outputPath': outputFigsFolder,
            'kind': None,
            'legend': False,                                         
            'align': 'center',
            'ylog': False,
            'xmetrics': ['Connections'],                                                      
            'barValues': None,    
            'barColor': barColor,                                              
            'yLowLimit': 0,                                                         
            'yUpperLimit': None,  
            'grid': 'y',                                                   
            'title': "Number of Connections with each Peer",                  
            'xlabel': "Peers Connected",                                                         
            'ylabel': 'Number of Connections',                                      
            'xticks': None,     
            'lw': 3,                                                  
            'titleSize': titleSize+4,                                                        
            'labelSize': labelSize+4,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+4,                                                       
            'xticksSize': ticksSize+4,                                                       
            'yticksSize': ticksSize+4,                                                      
            'tickRotation': 0,                                                     
            'show': False}) 

        barColor = 'black'
        plotFromPandas(rumorMetricsPanda, pdf, opts={                                   
            'figSize': figSize,                                                   
            'figTitle': 'DisconnectionsWithPeers.png',
            'pdf': pdfFile,                                     
            'outputPath': outputFigsFolder,
            'kind': None,
            'legend': False,                                         
            'align': 'center',
            'ylog': False,
            'xmetrics': ['Disconnections'],                                                      
            'barValues': None,
            'barColor': barColor,                                                  
            'yLowLimit': 0,                                                         
            'yUpperLimit': None,     
            'grid': 'y',                                                
            'title': "Number of Disconnections with each Peer",                  
            'xlabel': "Peers Connected",                                                         
            'ylabel': 'Number of Disconnections',                                      
            'xticks': None,     
            'lw': 3,                                                  
            'titleSize': titleSize+4,                                                        
            'labelSize': labelSize+4,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+4,                                                       
            'xticksSize': ticksSize+4,                                                       
            'yticksSize': ticksSize+4,                                                        
            'tickRotation': 0,                                                     
            'show': False}) 

        barColor = 'black'
        plotFromPandas(rumorMetricsPanda, pdf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'TimeConnectedWithPeers.png', 
            'pdf': pdfFile,                                    
            'outputPath': outputFigsFolder,
            'kind': None,
            'legend': False,                                         
            'align': 'center',
            'ylog': True,
            'xmetrics': ['Connected Time'],                                                      
            'barValues': None,   
            'barColor': barColor,                                               
            'yLowLimit': 0,                                                         
            'yUpperLimit': None,
            'grid': 'y',                                                    
            'title': "Total of Time Connected with each Peer",                  
            'xlabel': "Peers Connected",                                                         
            'ylabel': 'Time (in Minutes)',                                      
            'xticks': None,     
            'lw': 3,                                                  
            'titleSize': titleSize+4,                                                        
            'labelSize': labelSize+4,                                                        
            'lengendPosition': 1,                                                   
            'legendSize': labelSize+4,                                                       
            'xticksSize': ticksSize+4,                                                       
            'yticksSize': ticksSize+4,                                                     
            'tickRotation': 0,                                                     
            'show': False}) 

        barColor = 'black'

        print("Peer with highest RTT", rumorMetricsPanda.loc[rumorMetricsPanda['Latency'].idxmax()])

        plotFromPandas(rumorMetricsPanda, pdf, opts={                                   
            'figSize': figSize,                                                      
            'figTitle': 'RTTWithPeers.png', 
            'pdf': pdfFile,                                    
            'outputPath': outputFigsFolder,
            'kind': None,
            'legend': False,                                         
            'align': 'center',
            'ylog': True,
            'xmetrics': ['Latency'],                                                      
            'barValues': None,   
            'barColor': barColor,                                               
            'yLowLimit': 0,                                                         
            'yUpperLimit': None,
            'grid': 'y',                                                    
            'title': "RTT with each Peer",                  
            'xlabel': "Peers Connected",                                                         
            'ylabel': 'Seconds',                                      
            'xticks': None,   
            'lw': 3,                                                    
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
            if row['Beacon Blocks'] in messagesDics:
                messagesDics[row['Beacon Blocks']] = messagesDics[row['Beacon Blocks']] + 1
            else:
                messagesDics[row['Beacon Blocks']] = 1

        sortedDict = collections.OrderedDict(sorted(messagesDics.items()))

        xarray = []
        yarray = []
        for item in sortedDict:
            xarray.append(item)
            yarray.append(sortedDict[item])

        #print(rumorMetricsPanda.loc[rumorMetricsPanda['Beacon Blocks'].idxmax()])

        plotColumn(rumorMetricsPanda, pdf, opts={
            'figSize': figSize, 
            'figTitle': 'BeaconBlockMessagePerClient.png',
            'pdf': pdfFile, 
            'outputPath': outputFigsFolder,
            'xlog': False,
            'ylog': True,
            'xMetrics': None,
            'yMetrics': ['Beacon Blocks'],
            'sortmetrics': 'Beacon Blocks',
            'xticks': None,
            'xLowLimit': 0,
            'xUpperLimit': len(rumorMetricsPanda),
            'xRange': 1,
            'yLowLimit': 10**0,
            'yRange': None,
            'yUpperLimit': None,
            'title': "Received Beacon Blocks from each Peer",
            'xLabel': "Peers Connected",
            'yLabel': 'Number of Messages Received',
            'legendLabel': None,
            'titleSize': titleSize + 4,
            'labelSize': labelSize + 4,
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
            'tickSize': 20})


        barColor = 'black'

        auxPanda = rumorMetricsPanda.sort_values(by='Beacon Blocks', ascending=True)
        cont = 0

        auxrow = rumorMetricsPanda.loc[rumorMetricsPanda['Total Messages'].idxmax()]
        print(auxrow)
        
        
        plotColumn(rumorMetricsPanda, pdf, opts={
            'figSize': figSize, 
            'figTitle': 'TotalMesagesPerTimeConnected.png',
            'pdf': pdfFile, 
            'outputPath': outputFigsFolder,
            'xlog': True,
            'ylog': True,
            'xMetrics': 'Connected Time',
            'yMetrics': ['Total Messages'],
            'sortmetrics': None,
            'xticks': True,
            'xLowLimit': None,
            'xUpperLimit': None, # maxX
            'xRange': None, # minX
            'yLowLimit': 10**0,
            'yRange': None,
            'yUpperLimit': None,
            'title': "Total of Messages for Connected Time",
            'xLabel': "Connected Time (Minutes)",
            'yLabel': 'Number of Messages Received',
            'legendLabel': None,
            'titleSize': titleSize + 4,
            'labelSize': labelSize + 4,
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
            'tickSize': 20})

    # ------ End of Get data -------

    print("Succesfully Analyzed!")

if __name__ == '__main__':
    main()
