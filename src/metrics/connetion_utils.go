package metrics

// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnectionEvents(eventList []ConnectionEvents, currentTime time.Duration) (int, int, int) {
    var startingTime int = 0
    var finishingTime int = 0
    // aux variables
    var prevEvent string = "Disconnection"
    var prevTime int = 0
    var timeRange int = 500
    var contConn int = 0
    var contDisc int = 0
    var ttime int = 0 // total connected time
    var ctime int = 0 // aux time counter
//    var connFlag bool = false

    for _, event in range eventList{
        if prevEvent != event.ConnectionType || event.TimeMili >= (prevTime + timeRange){
            if event.ConnectionType == "Connection" {
                contConn = contConn + 1
                ctime = event.TimeMili // in milliseconds
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
            } else if {
                contDisc = contDisc + 1
                ttime = ttime + (event.TimeMili - ctime) // millis
                ctime = event.TimeMili
                prevEvent = event.ConnectionType
                prevTime = event.TimeMili
           }
        }
        if startingTime == 0{
            startingTime = event.ConnectionType
            finishingTime = event.TimeMili
        } else {
            if startingTime > event.ConnectionType{
                startingTime = event.TimeMili
            }
            if finishingTime < event.TimeMili {
                finishingTime = event.TimeMili
            }
        }
    }
    return contConn, contDisc, ttime/60000 // return the connection time in minutes ( / 60*1000)
}
