package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// IP-API message structure
type IpApiMessage struct {
	Query       string `json:"query"`
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Zip         string `json:"zip"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Timezone    string `json:"timezone"`
	Isp         string `json:"isp"`
	Org         string `json:"org"`
	As          string `json:"as"`
}

// TODO: temporal fix. put into a class
var attemptsLeft int = 10
var timeLeft int = 1

// get location country and City from the multiaddress of the peer on the peerstore
func GetLocationFromIp(ip string) (country string, city string, err error) {
	url := "http://ip-api.com/json/" + ip

	// When getting close to 0 attempts
	if attemptsLeft < 5 {
		time.Sleep(time.Duration(timeLeft) * time.Second)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", "", errors.Wrap(err, "could not get data from ip api")
	}

	attemptsLeft, _ = strconv.Atoi(resp.Header["X-Rl"][0])
	timeLeft, _ = strconv.Atoi(resp.Header["X-Ttl"][0])

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Todo struct
	var ipApiResp IpApiMessage
	json.Unmarshal(bodyBytes, &ipApiResp)

	// Check if the status of the request has been succesful
	if ipApiResp.Status != "success" {
		return "", "", errors.New("status from ip different than success")
	}

	country = ipApiResp.Country
	city = ipApiResp.City

	// check if country and city are correctly imported
	if len(country) == 0 || len(city) == 0 {
		return "", "", errors.New("country or city are empty")
	}

	// return the received values from the received message
	return country, city, nil
}
