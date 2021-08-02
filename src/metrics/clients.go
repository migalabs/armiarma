package metrics

import (
	//"strconv"
)

type Version struct {
	Version string
	Count   int //needed?
}

type Client struct {
	ClientName string
	Versions   []Version
	Count      int //needed?
}

type Clients struct {
	Clients []Client
}

func NewClients() Clients {
	clients := Clients {
		Clients: make([]Client, 0),
	}
	return clients
}

func (c *Clients) AddClientVersion(clientName, clientVersion string) {
	for _, cl := range c.Clients {
		if clientName == cl.ClientName {
			for _, vr := range cl.Versions {
				if clientVersion == vr.Version {
					vr.Count++
				} else {
					newVersion := Version{Version: clientVersion, Count: 1}
					cl.Versions = append(cl.Versions, newVersion)
				}
			}
		} else {
			newVersion := Version{
				Version: clientVersion,
				Count: 1}
			newClient := Client{
				ClientName: clientName,
				Versions: make([]Version, 1),
				Count: 1,
			}
			newClient.Versions[0] = newVersion
			c.Clients = append(c.Clients, newClient)
		}
	}
}
