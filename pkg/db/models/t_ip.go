package models

import "time"

const (
	IpInfoTTL = 30 * 24 * time.Hour // 30 days

)

// IP-API message structure
type IpApiMsg struct {
	Query         string  `json:"query"`
	Status        string  `json:"status"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"countryCode"`
	Region        string  `json:"region"`
	RegionName    string  `json:"regionName"`
	City          string  `json:"city"`
	Zip           string  `json:"zip"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Isp           string  `json:"isp"`
	Org           string  `json:"org"`
	As            string  `json:"as"`
	AsName        string  `json:"asname"`
	Mobile        bool    `json:"mobile"`
	Proxy         bool    `json:"proxy"`
	Hosting       bool    `json:"hosting"`
}

type ApiResp struct {
	IpInfo    IpInfo
	DelayTime time.Duration
	Err       error
}

type IpInfo struct {
	IpApiMsg
	ExpirationTime time.Time
}
