package metrics

import (
    "strings"
)

// Connection Utils

// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnectionEvents(eventList []ConnectionEvents, currentTime int64) (int64, int64, float64) {
    var startingTime int64 = 0
    var finishingTime int64 = 0
    // aux variables
    var prevEvent string = "Disconnection"
    var prevTime int64 = 0
    var timeRange int64 = 500
    var contConn int64 = 0
    var contDisc int64 = 0
    var ttime int64 = 0 // total connected time
    var ctime int64 = 0 // aux time counter
    // flag that will be used as watchdog to see if at the moment of exporting the peer was connected
    var connFlag bool = false

    for _, event := range eventList{
        if prevEvent != event.ConnectionType || event.TimeMili >= (prevTime + timeRange){
            if event.ConnectionType == "Connection" {
                contConn = contConn + 1
                ctime = event.TimeMili // in milliseconds
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
                connFlag = true
            } else if event.ConnectionType == "Disconnection"{
                contDisc = contDisc + 1
                ttime = ttime + (event.TimeMili - ctime) // millis
                ctime = event.TimeMili
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
                connFlag = false
           }
        }
        if startingTime == 0{
            startingTime = event.TimeMili
            finishingTime = event.TimeMili
        } else {
            if startingTime > event.TimeMili{
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
    return contConn, contDisc, float64(ttime)/60000 // return the connection time in minutes ( / 60*1000)
}


// Client utils
// Main function that will analyze the client type and verion out of the Peer UserAgent
// return the Client Type and it's verison (if determined)
func FilterClientType(fullName string) (string, string) {
	var client string
	var version string
	// get the UserAgent in lowercases
	fullName = strings.ToLower(fullName)
	// check the client type
	if strings.Contains(fullName, "lighthouse") { // the client is from Lighthouse
		// Lighthouse UserAgent Example: "Lighthouse/v1.0.3-65dcdc3/x86_64-linux"
		client = "Lighthouse"
		// Extract version
		s := strings.Split(fullName, "/")
		aux := strings.Split(s[1], "-")
		version = aux[0]
	} else if strings.Contains(fullName, "prysm") { // the client is from Prysm
		// Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
		client = "Prysm"
		// Extract version
		s := strings.Split(fullName, "/")
		version = s[1]
	} else if strings.Contains(fullName, "teku") { // the client is from Prysm
		// Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
		client = "Teku"
		// Extract version
		s := strings.Split(fullName, "/")
		aux := strings.Split(s[2], "+")
		version = aux[0]
	} else if strings.Contains(fullName, "nimbus") {
		client = "Nimbus"
		version = "Unknown"
	} else if strings.Contains(fullName, "lodestar") {
		client = "Lodestar"
		version = "Unknown"
	} else if strings.Contains(fullName, "unknown") {
		client = "Unknown"
		version = "Unknown"
	} else {
		client = "Unknown"
		version = "Unknown"
	}
	return client, version
}
