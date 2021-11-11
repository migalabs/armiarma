package db

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// Version: this struct stores information
// about a single Client Version
type Version struct {
	Name  string
	Count int
}

// Client: stores the information about all versions of the client
type Client struct {
	Versions []Version
}

func NewClient() *Client {
	return &Client{
		Versions: make([]Version, 0),
	}
}

// ReturnTotalCount
// * This method returns the total number of nodes in any version
func (c Client) ReturnTotalCount() int {
	result := 0

	for _, versionTmp := range c.Versions {
		result += versionTmp.Count
	}
	return result
}

// AddVersion
// * This method adds one or creates a new version under the client
// @param inputVersion: the version to add
func (c *Client) AddVersion(inputVersion string) {

	for i, versionTmp := range c.Versions {
		if versionTmp.Name == inputVersion {
			c.Versions[i].Count += 1 // add one if found
			return
		}
	}

	// new version
	newVersion := Version{
		Name:  inputVersion,
		Count: 1,
	}
	c.Versions = append(c.Versions, newVersion)
	return
}

// ClientDist: stores the distribution of all watched clients
type ClientDist struct {
	Clients map[string]*Client
}

func NewClientDist() ClientDist {
	clientDist := ClientDist{
		Clients: make(map[string]*Client, 0),
	}
	return clientDist
}

// AddClientVersion
// * This method will add one to the count, or create a new entry
// * for the given version in the given client
func (c *ClientDist) AddClientVersion(clientName, clientVersion string) {

	client, ok := c.Clients[clientName]

	if !ok {
		// the client does not exist, add it
		c.Clients[clientName] = NewClient()
		client, ok = c.Clients[clientName]
	}

	client.AddVersion(clientVersion)
}

// GetClientNames
// * This method returns the names of the wacthed clients
// @return the string array containing the names
func (c ClientDist) GetClientNames() []string {
	clientNames := make([]string, 0)
	for k := range c.Clients {
		clientNames = append(clientNames, k)
	}
	return clientNames
}

// GetCountOfClient
// * This method returns the number of watched nodes under
// * the given client
// @param clientName: the client to count on
// @return the number of nodes
func (c ClientDist) GetCountOfClient(clientName string) int {
	return c.Clients[clientName].ReturnTotalCount()
}

// GetClientDistribution
// * This method returns the number of nodes for each watched client
// @return a map where key: clientName, value: totalCount
func (c ClientDist) GetClientDistribution() map[string]int {
	distributionResult := make(map[string]int)
	for clientName, client := range c.Clients {
		distributionResult[clientName] += client.ReturnTotalCount()
	}
	return distributionResult
}

// GetTotalCount
// * This method returns the total number of watched nodes
// @return the total number
func (c ClientDist) GetTotalCount() int {
	total := 0
	for _, client := range c.Clients {
		total += client.ReturnTotalCount()
	}
	return total
}

// GetClientVersionDistribution
// * This method calculates the distribution of clients and versions
// @return a map where key: client_version, value:count
func (c ClientDist) GetClientVersionDistribution() map[string]int {
	result := make(map[string]int, 0)
	for clientName, clientTmp := range c.Clients {
		for _, versionTmp := range clientTmp.Versions {
			clientVersion := clientName + "_" + versionTmp.Name
			result[clientVersion] = versionTmp.Count
		}
	}
	return result
}

// StringMapMetric: stores a standard map metric for prometheus
type StringMapMetric struct {
	data map[string]float64
}

func NewStringMapMetric() StringMapMetric {
	return StringMapMetric{
		data: make(map[string]float64),
	}
}

// AddOneorCreate
// * In case the key does not exist: create.
// * In case it exists, add one.
// @param inputKey: the key to add or create
func (m *StringMapMetric) AddOneorCreate(inputKey string) {
	_, ok := m.data[inputKey]
	if ok {
		m.data[inputKey] += 1
	} else {
		m.data[inputKey] = 1
	}
}

// SetValues
// * Given a prometheus registered variable, dump the metrics into it
// @param registered_name: the prometheus variable where to dump data
func (m StringMapMetric) SetValues(registered_name *prometheus.GaugeVec) {
	registered_name.Reset() // reset it first of all
	for k, v := range m.data {
		registered_name.WithLabelValues(k).Set(v)
	}
}

// ObtainDistribution
// * This method will calculate the distribution of the curent metric
// @return the distribution in a StringMapMetric object
func (m StringMapMetric) ObtainDistribution() StringMapMetric {
	result := NewStringMapMetric()

	for _, v := range m.data {
		result.AddOneorCreate(fmt.Sprintf("%.0f", v))
	}

	return result

}

// Traspose
// * This method will traspose the current metric.
// * This is: keys are now values and values are now keys
func (m *StringMapMetric) Traspose() {
	result := make(map[string]float64)
	for k, v := range m.data {
		new_value, err := strconv.ParseFloat(k, 64)
		if err != nil {
			// log error
		} else {
			result[fmt.Sprintf("%.0f", v)] = new_value
		}
	}
	m.data = result
}
