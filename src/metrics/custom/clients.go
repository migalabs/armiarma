package custom

import (
	"strconv"
)

type Client struct {
	Total int
	Versions []Version
}

func NewClient() Client {
	c := Client {
		Total: 0,
		Versions: make([]Version, 0),
	}
	return c
}

func (c *Client) SetTotal(t int) {
	c.Total = t
}

func (c *Client) AddVersion(v string, t int) {
	vers := NewVersion(v, t)
	c.Versions = append(c.Versions, vers)	
}

type Version [2]string

func NewVersion(v string, t int) Version {
	var vers Version
	vers[0] = v
	vers[1] = strconv.Itoa(t)
	return vers
}
