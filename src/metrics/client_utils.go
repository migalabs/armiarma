package metrics

import (
    "fmt"
    "strings"
)

var MainClients []string = {"Lighthouse", "Teku", "Prysm", "Nimbus", "Lodestar", "Unknown"}

// Main function that will analyze the client type and verion out of the Peer UserAgent
// return the Client Type and it's verison (if determined)
func FilterClientType( fullName string ) ( string, string) {
    var client string
    var version string
    // get the UserAgent in lowercases
    fullName = strings.ToLower(fullName)
    // check the client type
    if strings.Contains(fullName, "lighthouse"){ // the client is from Lighthouse
        // Lighthouse UserAgent Example: "Lighthouse/v1.0.3-65dcdc3/x86_64-linux"
        client = "Lighthouse"
        // Extract version
        s := strings.Split(fullName, "/")
        version = s[1]
    } else if strings.Contains(fullName, "prysm"){ // the client is from Prysm
        // Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
        client = "Prysm"
        // Extract version
        s := strings.Split(fullName, "/")
        version = s[1]
    } else if strings.Contains(fullName, "teku"){ // the client is from Prysm
        // Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
        client = "Teku"
        // Extract version
        s := strings.Split(fullName, "/")
        version = s[2]
    } else if strings.Contains(fullName, "nimbus"){
        client = "Nimbus"
        version = "Unknown"
    } else if stings.Contains(fullName, "lodestar"){
        client = "Lodestar"
        version = "Unknown"
    } else if strings.Contains(fullName, "unknown"){
        client = "Unknown"
        version = "Unknown"
    } else {
        client = "Unknown"
        version = "Unknown"
    }
    return client, version
}
