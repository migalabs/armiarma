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

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultIpTTL  = 30 * 24 * time.Hour // 30 days
	ipReqBuffSize = 1024                // number of ip queries that can be buffered in the channel
	ipApiEndpoint = "http://ip-api.com/json/{__ip__}?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query"
)

var TooManyRequestError error = fmt.Errorf("error HTTP 429")

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
	go func() {
		var nextDelayRequest time.Duration
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

				// since it didn't exist or did expire, request the ip
				// new API call needs to be done
				log.Debugf(" making API call for %s", reqIp)

			reqLoop:
				for {
					atomic.AddInt32(c.apiCalls, 1)
					respC := c.locateIp(reqIp)
					select {
					case apiResp := <-respC:
						// check if there is an error
						switch apiResp.Err {
						case TooManyRequestError:
							nextDelayRequest = apiResp.DelayTime
							// if the error reports that we tried too many calls on the API, sleep given time and try again
							log.Debug("call", reqIp, "-> error received:", err.Error(), "\nwaiting ", nextDelayRequest+(5*time.Second))
							ticker := time.NewTicker(nextDelayRequest + (5 * time.Second))
							select {
							case <-ticker.C:
								continue
							case <-c.ctx.Done():
								log.Info("context closure has been detecting, closing IpApi caller")
								return
							}
						case nil:
							// if the error is different from TooManyRequestError break loop and store the request
							log.Debugf("call %s-> api req success", reqIp)
							// check what to do with the results INSERT / UPDATE
							if exists && expired {
								if err := c.dbClient.UpdateIP(apiResp.IpInfo); err != nil {
									log.Error(err)
								}
							} else {
								if err := c.dbClient.InsertNewIP(apiResp.IpInfo); err != nil {
									log.Error(err)
								}
							}
							break reqLoop

						default:
							log.Debug("call", reqIp, "-> diff error received:", err.Error())
							break reqLoop

						}

					case <-c.ctx.Done():
						log.Info("context closure has been detecting, closing IpApi caller")
						return
					}
				}
				// check if there is any waiting time that we have to respect before next connection
				if nextDelayRequest != time.Duration(0) {
					log.Debugf("number of allowed requests has been exceed, waiting %#v", nextDelayRequest+(5*time.Second))
					// set req delay to true, noone can make requests
					time.Sleep(nextDelayRequest + (5 * time.Second))
				}

			// the context has been deleted, end go routine
			case <-c.ctx.Done():
				// close the channels
				close(c.locationRequest)
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

func (c *IpLocator) locateIp(ip string) chan models.ApiResp {
	respC := make(chan models.ApiResp)
	go callIpApi(ip, respC)
	return respC
}

// get location country and City from the multiaddress of the peer on the peerstore
func callIpApi(ip string, respC chan models.ApiResp) {
	var apiResponse models.ApiResp
	apiResponse.IpInfo, apiResponse.DelayTime, apiResponse.Err = CallIpApi(ip)
	respC <- apiResponse
	// defer ^
}

func CallIpApi(ip string) (ipInfo models.IpInfo, delay time.Duration, err error) {

	url := strings.Replace(ipApiEndpoint, "{__ip__}", ip, 1)

	// Make the IP-APi request
	resp, err := http.Get(url)
	if err != nil {
		err = errors.Wrap(err, "unable to locate IP"+ip)
		return
	}
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])
	// check if the error that we are receiving means that we exeeded the request limit
	if resp.StatusCode == 429 {
		log.Warnf("limit of requests per minute has been exeeded, wait for next call %s secs", resp.Header["X-Ttl"][0])
		err = TooManyRequestError
		delay = time.Duration(timeLeft) * time.Second
		return
	}

	// Check the attempts left that we have to call the api
	attemptsLeft, _ := strconv.Atoi(resp.Header["X-Rl"][0])
	if attemptsLeft <= 0 {
		// if there are no more attempts left against the api, check how much time do we have to wait
		// until we can call it again
		// set the delayTime that we return to the given seconds to wait
		delay = time.Duration(timeLeft) * time.Second
	}

	// check if the response was success or not
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "could not read response body")
		return
	}

	var apiMsg models.IpApiMsg
	// Convert response body to struct
	err = json.Unmarshal(bodyBytes, &apiMsg)
	if err != nil {
		err = errors.Wrap(err, "could not unmarshall response")
		return
	}
	// Check if the status of the request has been succesful
	if apiMsg.Status != "success" {
		err = errors.New(fmt.Sprintf("status from ip different than success, resp header:\n %#v", *resp))
		return
	}

	ipInfo.ExpirationTime = time.Now().UTC().Add(defaultIpTTL)
	ipInfo.IpApiMsg = apiMsg
	return
}
