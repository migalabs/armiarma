package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultIpTTL        = 30 * 24 * time.Hour // 30 days
	ipReqBuffSize       = 1024                // number of ip queries that can be buffered in the channel
	ipApiEndpoint       = "http://ip-api.com/json/?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query&query={__ip__}"
	TooManyRequestError = "error HTTP 429"
)

// PEER LOCALIZER

type IpLocator struct {
	ctx context.Context
	// Request channels
	locationRequest chan string
	// dbClient
	dbClient *postgresql.DBClient

	// control variables for IP-API request
	// Control flags from prometheus
	apiCalls *int32
}

func NewIpLocator(ctx context.Context, dbCli *postgresql.DBClient) IpLocator {
	calls := int32(0)
	return IpLocator{
		ctx:             ctx,
		locationRequest: make(chan string, ipReqBuffSize),
		dbClient:        dbCli,
		apiCalls:        &calls,
	}
}

// Run the necessary routines to locate the IPs
func (c *IpLocator) Run() {
	//l.SetLevel(Logrus.DebugLevel)
	c.locatorRoutine()
}

// locatorRoutine is the main routine that will wait until an request to identify an IP arrives
// or if the routine gets canceled
func (c *IpLocator) locatorRoutine() {
	log.Info("IP locator routine started")
	apiCallChan := make(chan string, ipRequestSize) // Give it a size of 20 just in case there are many inbound requests at the same time
	go func() {
		for {
			select {
			// New request to identify an IP
			case reqIp := <-c.locationRequest:
				log.Debug("new request has been received for ip:", reqIp)
				// Check if the IP is already in the cache
				exists, expired, err := c.dbClient.CheckIpRecords(reqIp)
				if err != nil {
					log.Error("unable to check if IP already exists -", err.Error()) // Should it be a Panic?
				}
				// if exists and it didn't expired, don't do anything
				if exists && !expired {
					continue
				}

				// old
				if ok { // the response was alreadi in the cache.
					log.Debugf("ip %s was in cache", reqIp)
					response := *cacheResp
					// TODO: might be interesting to check if the error is due to an invalid IP
					// 		or if the reported error is due to a connection with the server (too many requests in the last minute)
					if response.err != nil {
						log.Warn("readed response from cache includes an error:", response.err.Error())
					}
					// put the received response in the channel to reply the external request

					request.respChan <- response
				} else {
					// Finally there is space in the channel
					apiCallChan <- request
				}

			// the context has been deleted, end go routine

			case <-c.ctx.Done():
				// close the channels
				close(c.locationRequest)
				return
			}
		}
	}()
	go func() {
		for {
			select {
			// request to Identify IP through api call
			case request := <-apiCallChan:
				// control flag to see if we have to wait untill we get the next API call
				// if nextDelayRequest is 0, we can go for the next one,
				var nextDelayRequest time.Duration
				var breakCallLoop bool = false
				var response ipResponse

				call := atomic.LoadInt32(c.apiCalls)
				atomic.AddInt32(c.apiCalls, 1)

				response.ip = request.ip
				// new API call needs to be done
				log.Debugf("call %d-> ip %s not in cache, making API call", call, response.ip)
				for !breakCallLoop {
					// if req delay is setted to true, make new request
					// make the API call, and receive the apiResponse, the nextDelayRequest and the error from the connection
					response.resp, nextDelayRequest, response.err = c.getLocationFromIp(request.ip)
					if response.err != nil {
						if response.err.Error() == TooManyRequestError {
							// if the error reports that we tried too many calls on the API, sleep given time and try again
							log.Debug("call", call, "-> error received:", response.err.Error(), "\nwaiting ", nextDelayRequest+(5*time.Second))
							// set req delay to true, noone can make requests
							// TODO: Change all the sleeps for
							/*
									select {
									case <- time.After(DURATION):
									case <- ctx.Done()
								}
							*/
							time.Sleep(nextDelayRequest + (5 * time.Second))
							continue
						} else {
							log.Debug("call", call, "-> diff error received:", response.err.Error())
							break
						}
					} else {
						log.Debugf("call %d-> api req success", call)
						// if the error is different from TooManyRequestError break loop and store the request
						break
					}

				}
				// check if there is any waiting time that we have to respect before next connection
				if nextDelayRequest != time.Duration(0) {
					log.Debugf("call %d-> number of allowed requests has been exceed, waiting %#v", call, nextDelayRequest+(5*time.Second))
					// set req delay to true, noone can make requests
					time.Sleep(nextDelayRequest + (5 * time.Second))
				}

				log.Debugf("call %d-> saving new request and return it")
				// add the response into the responseCache
				c.reqCache.addRequest(&response)

				// put the received response in the channel to reply the external request
				request.respChan <- response

			// the context has been deleted, end go routine
			case <-c.ctx.Done():
				// close the channels
				close(apiCallChan)
				return
			}
		}
	}()
}

// LocateIP is an externa request that any module could do to identify an IP
func (c *IpLocator) LocateIP(ip string) {
	// put the request in the Queue
	c.locationRequest <- ip
}

//
func (c *IpLocator) Close() {
	log.Info("closing IP-API service")
	// close the context for ending up the routine

}

type ipResponse struct {
	ip   string
	err  error
	resp IpInfo
}

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

type IpInfo struct {
	IpApiMsg
	ExpirationTime time.Time
}

// get location country and City from the multiaddress of the peer on the peerstore
func (c *IpLocator) getLocationFromIp(ip string) (IpInfo, time.Duration, error) {
	url := strings.Replace(ipApiEndpoint, "{__ip__}", ip, 1)

	var ipInfo IpInfo
	var delayTime time.Duration
	var retErr error

	// Make the IP-APi request
	resp, err := http.Get(url)
	// increment api calls counter
	atomic.AddInt32(c.apiCalls, 1)
	if err != nil {
		retErr = errors.Wrap(err, "unable to locate IP"+ip)
		return ipInfo, delayTime, retErr
	}
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])
	// check if the error that we are receiving means that we exeeded the request limit
	if resp.StatusCode == 429 {
		log.Warnf("limit of requests per minute has been exeeded, wait for next call %s secs", resp.Header["X-Ttl"][0])
		retErr = errors.New(TooManyRequestError)
		delayTime = time.Duration(timeLeft) * time.Second
		return ipInfo, delayTime, retErr
	}

	// Check the attempts left that we have to call the api
	attemptsLeft, _ := strconv.Atoi(resp.Header["X-Rl"][0])
	if attemptsLeft <= 0 {
		// if there are no more attempts left against the api, check how much time do we have to wait
		// until we can call it again
		// set the delayTime that we return to the given seconds to wait
		delayTime = time.Duration(timeLeft) * time.Second
	}

	// check if the response was success or not
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		retErr = errors.Wrap(err, "could not read response body")
		return ipInfo, delayTime, retErr
	}

	var apiMsg IpApiMsg
	// Convert response body to struct
	err = json.Unmarshal(bodyBytes, &apiMsg)
	if err != nil {
		retErr = errors.Wrap(err, "could not unmarshall response")
		return ipInfo, delayTime, retErr
	}
	// Check if the status of the request has been succesful
	if apiMsg.Status != "success" {
		retErr = errors.New(fmt.Sprintf("status from ip different than success, resp header:\n %#v", *resp))
		return ipInfo, delayTime, retErr
	}
	ipInfo.ExpirationTime = time.Now().UTC().Add(defaultIpTTL)
	ipInfo.IpApiMsg = apiMsg
	return ipInfo, delayTime, retErr
}
