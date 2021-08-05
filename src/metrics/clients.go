package metrics

type Version struct {
	Version string
	Count   int
}

type Clients struct {
	Clients map[string][]Version
}

func NewClients() Clients {
	clients := Clients{
		Clients: make(map[string][]Version, 0),
	}
	return clients
}

func (c *Clients) AddClientVersion(clientName, clientVersion string) {
	if versions, ok := c.Clients[clientName]; ok {
		versionFound := false
		for i, v := range versions {
			if v.Version == clientVersion {
				c.Clients[clientName][i].Count++
				versionFound = true
			}
		}
		if !versionFound {
			c.Clients[clientName] = append(c.Clients[clientName],
				Version{Version: clientVersion, Count: 1})
		}
	} else {
		c.Clients[clientName] = make([]Version, 1)
		c.Clients[clientName][0] = Version{Version: clientVersion, Count: 1}
	}
}

func (c *Clients) GetClientNames() []string {
	clientNames := make([]string, 0)
	for k := range c.Clients {
		clientNames = append(clientNames, k)
	}
	return clientNames
}

func (c *Clients) GetPeersOfClient(clientName string) int {
	total := 0
	for _, v := range c.Clients[clientName] {
		total += v.Count
	}
	return total
}
