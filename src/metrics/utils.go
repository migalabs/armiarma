package metrics

import (
	"strings"
	//"github.com/protolambda/rumor/metrics/utils"
	"github.com/protolambda/rumor/metrics/custom"
)

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
	} else if strings.Contains(fullName, "js-libp2p") {
		client = "Lodestar"
		s := strings.Split(fullName, "/")
		version = s[1]
	} else if strings.Contains(fullName, "unknown") {
		client = "Unknown"
		version = "Unknown"
	} else {
		client = "Unknown"
		version = "Unknown"
	}
	return client, version
}

// Function that iterates through the peers keeping track of the client type, and versions
func AnalyzeClientType(gm *GossipMetrics, clientname string) custom.Client {
	client := custom.NewClient()
	/*

	clicnt := 0
	versions := make(map[string]int)

	c.GossipMetrics.Range(func(k, val interface{}) bool {
		v := val.(utils.PeerMetrics)
		item := v.ClientType
		_, err = csvFile.WriteString(v.ToCsvLine())

		if item == clientname { // peer with the same client type as the one we are searching for
			clicnt += 1
			// add the version to the map or increase the actual counter
			ver := df.ClientVersions.GetByIndex(idx)
			//x := versions[ver]
			versions[v.ClientType] += 1
		}
		return true
	})*/

	/*

	clicnt := 0
	versions := make(map[string]int)

	// iterate through the peer metrics with reading the client List
	for idx, item := range df.ClientTypes {
		if item == clientname { // peer with the same client type as the one we are searching for
			clicnt += 1
			// add the version to the map or increase the actual counter
			ver := df.ClientVersions.GetByIndex(idx)
			//x := versions[ver]
			versions[ver] += 1
		}
	}
	// after reading the entire metrics we can generate the custom.Client struct
	client.SetTotal(clicnt)
	v := make([]string, 0, len(versions))
	for val := range versions {
		v = append(v, val)
	}
	sort.Strings(v)
	for i, j := 0, len(v)-1; i < j; i, j = i+1, j-1 {
		v[i], v[j] = v[j], v[i]
	}
	for _, item := range v {
		client.AddVersion(item, versions[item])
	}*/
	return client
}

// Function that iterates through the peers keeping track of the client type, and versions if the peer was requested the metadata
//func (df MetricsDataFrame) AnalyzeClientTypeIfMetadataRequested(clientname string) custom.Client {
//	client := custom.NewClient()
	/*
	clicnt := 0
	versions := make(map[string]int)

	// iterate through the peer metrics with reading the client List
	for idx, item := range df.ClientTypes {
		if item == clientname { // peer with the same client type as the one we are searching for and metadata was requested
			// check if the metadata was requested from the peer
			i := df.RequestedMetadata.GetByIndex(idx)
			if i {
				clicnt += 1
				// add the version to the map or increase the actual counter
				ver := df.ClientVersions.GetByIndex(idx)
				//x := versions[ver]
				versions[ver] += 1
			}
		}
	}
	// after reading the entire metrics we can generate the custom.Client struct
	client.SetTotal(clicnt)
	v := make([]string, 0, len(versions))
	for val := range versions {
		v = append(v, val)
	}
	sort.Strings(v)
	for i, j := 0, len(v)-1; i < j; i, j = i+1, j-1 {
		v[i], v[j] = v[j], v[i]
	}
	for _, item := range v {
		client.AddVersion(item, versions[item])
	}*/
//	return client
//}
