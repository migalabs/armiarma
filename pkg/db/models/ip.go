package models

import (
	"time"

	"github.com/ip2location/ip2proxy-go/v4"
)

const (
	IpInfoTTL = 30 * 24 * time.Hour // 30 days
)

// IP-API message structure
type IpApiMsg struct {
	IP            string  `json:"query"`
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

// IsEmpty returns true if the struct reply that we got from the IP-API is empty
func (m *IpApiMsg) IsEmpty() bool {
	return m.Country == "" && m.City == ""
}

// some fields are commented because we don't have data for them
func mapTempIpInfoToApiMsg(data ip2proxy.IP2ProxyRecord, ip string) IpApiMsg {
	return IpApiMsg{
		IP:     ip,
		Status: "success",
		// Continent:    	"",
		// ContinentCode: 	"",
		Country:     data.CountryLong,
		CountryCode: data.CountryShort,
		Region:      data.Region,
		// RegionName:    	"",
		City: data.City,
		// Zip:           	"",
		// Lat: 			"",
		// Lon: 			"",
		Isp: data.Isp,
		// Org:           	"",
		// As:            	"",
		// AsName:       	"",
		Mobile: data.UsageType == "MOB",
		Proxy:  data.ProxyType != "",
		// Hosting:       	false,
	}
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
