package metrics

import(
    "time"
)

// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnectionEvents(eventList []ConnectionEvents, currentTime time.Time) (int64, int64, int64) {
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
//    var connFlag bool = false
    // temporary
    _ = currentTime

    for _, event := range eventList{
        if prevEvent != event.ConnectionType || event.TimeMili >= (prevTime + timeRange){
            if event.ConnectionType == "Connection" {
                contConn = contConn + 1
                ctime = event.TimeMili // in milliseconds
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
            } else if event.ConnectionType == "Disconnection"{
                contDisc = contDisc + 1
                ttime = ttime + (event.TimeMili - ctime) // millis
                ctime = event.TimeMili
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
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
    return contConn, contDisc, ttime/60000 // return the connection time in minutes ( / 60*1000)
}
