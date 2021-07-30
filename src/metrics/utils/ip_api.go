package utils

import (
  "strings"
  "net/http"
  "fmt"
  "strconv"
  "time"
  "io/ioutil"
  "encoding/json"
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

// get IP, location country and City from the multiaddress of the peer on the peerstore
func GetIpAndLocationFromAddrs(multiAddrs string) (ip string, country string, city string) {
	ip = strings.TrimPrefix(multiAddrs, "/ip4/")
	ipSlices := strings.Split(ip, "/")
	ip = ipSlices[0]
	url := "http://ip-api.com/json/" + ip
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	attemptsLeft, _ := strconv.Atoi(resp.Header["X-Rl"][0])
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])

	if attemptsLeft == 0 { // We have exceeded the limit of requests 45req/min
		time.Sleep(time.Duration(timeLeft) * time.Second)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Println(err)
			country = "Unknown"
			city = "Unknown"
			return ip, country, city
		}
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Todo struct
	var ipApiResp IpApiMessage
	json.Unmarshal(bodyBytes, &ipApiResp)

	// Check if the status of the request has been succesful
	if ipApiResp.Status != "success" {
		/*
			fmt.Println("Error with the received response status,", ipApiResp.Status)
			if ipApiResp.Query == ip {
				fmt.Println("The given IP of the peer is private")
			}
		*/
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	country = ipApiResp.Country
	city = ipApiResp.City

	// check if country and city are correctly imported
	if len(country) == 0 || len(city) == 0 {
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	// return the received values from the received message
	return ip, country, city

}
