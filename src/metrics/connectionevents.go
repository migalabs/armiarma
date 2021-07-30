package metrics

import (
	"time"
)

// Connection event model
type ConnectionEvents struct {
	ConnectionType string
	TimeMili       int64
}

func GetTimeMiliseconds() int64 {
	now := time.Now()
	//secs := now.Unix()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return millis
}

// Connection Utils

// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnectionEvents(eventList []ConnectionEvents, currentTime int64) (int64, int64, float64) {
	var startingTime int64 = 0
	var finishingTime int64 = 0
	// aux variables
	var prevEvent string = "Disconnection"
	var prevTime int64 = 0
	var contConn int64 = 0
	var contDisc int64 = 0
	var ttime int64 = 0 // total connected time
	var ctime int64 = 0 // aux time counter
	// flag that will be used as watchdog to see if at the moment of exporting the peer was connected
	var connFlag bool = false

	for _, event := range eventList {
		if prevEvent != event.ConnectionType {
			if event.ConnectionType == "Connection" {
				contConn = contConn + 1
				ctime = event.TimeMili // in milliseconds
				prevEvent = event.ConnectionType
				connFlag = true
			} else if event.ConnectionType == "Disconnection" {
				contDisc = contDisc + 1
				ttime = ttime + (event.TimeMili - ctime) // millis
				ctime = event.TimeMili
				prevEvent = event.ConnectionType
				connFlag = false
			}
		}
		if startingTime == 0 {
			startingTime = event.TimeMili
			finishingTime = event.TimeMili
		} else {
			if startingTime > event.TimeMili {
				startingTime = event.TimeMili
			}
			if finishingTime < event.TimeMili {
				finishingTime = event.TimeMili
			}
		}
	}
	// Check if at the moment of exporting the peer was connected
	if connFlag {
		ttime = ttime + (currentTime - prevTime)
	}
	return contConn, contDisc, float64(ttime) / 60000 // return the connection time in minutes ( / 60*1000)
}

// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnDisconnTime(pm *PeerMetrics, currentTime int64) (int64, int64, float64) {
	var connTime int64
	// Use the counters in the PeerMetrics to check if the peer is still connected
	if pm.ConnFlag {
		connTime = pm.TotConnTime + (currentTime - pm.LastConn)
	} else {
		connTime = pm.TotConnTime
	}
	pm.LastExport = currentTime
	return pm.TotConnections, pm.TotDisconnections, float64(connTime) / 60000 // return the connection time in minutes ( / 60)
}
