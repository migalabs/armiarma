from django.shortcuts import render
from . import settings

def index(request):
    return render(request, 'index.html')

def graphs(request):
    return render(request, 'graphs.html')

def peerstoreVsConnected(request):
    return render(request, 'plots/peerstoreVsConnected.html')

def peersPerCountry(request):
    return render(request, 'plots/peersPerCountry.html')

def peersPerClient(request):
    return render(request, 'plots/peersPerClient.html')

def connectionsWithPeers(request):
    return render(request, 'plots/connectionsWithPeers.html')
    
def disconnectionsWithPeers(request):
    return render(request, 'plots/disconnectionsWithPeers.html')
    
def timeConnectedWithPeers(request):
    return render(request, 'plots/timeConnectedWithPeers.html')
    
def latencyWithPeers(request):
    return render(request, 'plots/latencyWithPeers.html')
    
def averageLatency(request):
    return render(request, 'plots/averageLatency.html')
    
def averageConnections(request):
    return render(request, 'plots/averageConnections.html')
    
def averageDisconnections(request):
    return render(request, 'plots/averageDisconnections.html')
    
def averageTime(request):
    return render(request, 'plots/averageTime.html')
    
def beaconBlocks(request):
    return render(request, 'plots/beaconBlocks.html')
    
def beaconAggregations(request):
    return render(request, 'plots/beaconAggregations.html')
    
def voluntaryExits(request):
    return render(request, 'plots/voluntaryExits.html')
    
def attesterSlashing(request):
    return render(request, 'plots/attesterSlashing.html')

def proposerSlashing(request):
    return render(request, 'plots/proposerSlashing.html')

def averageBeaconBlock(request):
    return render(request, 'plots/averageBeaconBlock.html')

def averageBeaconAggregations(request):
    return render(request, 'plots/averageBeaconAggregations.html')

def averageVoluntaryExits(request):
    return render(request, 'plots/averageVoluntaryExits.html')

def averageAttesterSlashings(request):
    return render(request, 'plots/averageAttesterSlashings.html')

def averageProposerSlashings(request):
    return render(request, 'plots/averageProposerSlashings.html')

def blocksPerClient(request):
    return render(request, 'plots/blocksPerClient.html')

def totalMessages(request):
    return render(request, 'plots/totalMessages.html')

    