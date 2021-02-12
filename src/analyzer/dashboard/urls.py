"""dashboard URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/3.1/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""
from django.contrib import admin
from django.urls import path

from . import views
from django.contrib.staticfiles.urls import staticfiles_urlpatterns

urlpatterns = [
    path('admin/', admin.site.urls),
    path('', views.index, name="index"),
    path('graphs', views.graphs, name="graphs"),
    path('PeerstoreVsConnectedPeers', views.peerstoreVsConnected, name="peerstoreVsConnected"),
    path('PeersPerCountries', views.peersPerCountry, name="peersPerCountry"),
    path('PeersPerClient', views.peersPerClient, name="peersPerClient"),
    path('ConnectionsWithPeers', views.connectionsWithPeers, name="connectionsWithPeers"),
    path('DisconnectionsWithPeers', views.disconnectionsWithPeers, name="disconnectionsWithPeers"),
    path('TimeConnectedWithPeers', views.timeConnectedWithPeers, name="timeConnectedWithPeers"),
    path('LatencyWithPeers', views.latencyWithPeers, name="latencyWithPeers"),
    path('AverageLatencyPerClientType', views.averageLatency, name="averageLatency"),
    path('AverageOfConnectionsPerClientType', views.averageConnections, name="averageConnections"),
    path('AverageOfDisconnectionsPerClientType', views.averageDisconnections, name="averageDisconnections"),
    path('AverageOfConnectedTimePerClientType', views.averageTime, name="averageTime"),
    path('MessagesFromBeaconBlock', views.beaconBlocks, name="beaconBlocks"),
    path('MessagesFromBeaconAggregateProof', views.beaconAggregations, name="beaconAggregations"),
    path('MessagesFromVoluntaryExit', views.voluntaryExits, name="voluntaryExits"),
    path('MessagesFromAttesterSlashing', views.attesterSlashing, name="attesterSlashing"),
    path('MessagesFromProposerSlashing', views.proposerSlashing, name="proposerSlashing"),
    path('MessageAverageFromBeaconBlock', views.averageBeaconBlock, name="averageBeaconBlock"),
    path('MessageAverageFromBeaconAggregateProof', views.averageBeaconAggregations, name="averageBeaconAggregations"),
    path('MessageAverageFromVoluntaryExit', views.averageVoluntaryExits, name="averageVoluntaryExits"),
    path('MessageAverageFromAttesterSlashing', views.averageAttesterSlashings, name="averageAttesterSlashings"),
    path('MessageAverageFromProposerSlashing', views.averageProposerSlashings, name="averageProposerSlashings"),
    path('BeaconBlockMessagePerClient', views.blocksPerClient, name="blocksPerClient"),
    path('TotalMesagesPerTimeConnected', views.totalMessages, name="totalMessages")

]

urlpatterns += staticfiles_urlpatterns()
