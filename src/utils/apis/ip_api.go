package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ModuleName string = "PEER-LOCALIZER"

	Log = logrus.WithField(
		"module", ModuleName,
	)

	ipApiEndpoint       = "http://ip-api.com/json/"
	TooManyRequestError = "error HTTP 429"
)

// PEER LOCALIZER

type PeerLocalizer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	reqCache requestCache
	// Request channels
	locationRequest chan locReq
	// control variables for IP-API request
	// Control flags from prometheus
	apiCalls *int32
}

func NewPeerLocalizer(ctx context.Context, cacheSize int) PeerLocalizer {
	locContext, cancelFunc := context.WithCancel(ctx)
	// generate the cache list
	reqCache := newRequestCache(cacheSize)

	calls := int32(0)
	return PeerLocalizer{
		ctx:             locContext,
		cancel:          cancelFunc,
		reqCache:        reqCache,
		locationRequest: make(chan locReq),
		apiCalls:        &calls,
	}
}

// Run the necessary routines to locate the IPs
func (c *PeerLocalizer) Run() {
	//l.SetLevel(Log.DebugLevel)
	c.locatorRoutine()
}

// locatorRoutine is the main routine that will wait until an request to identify an IP arrives
// or if the routine gets canceled
func (c *PeerLocalizer) locatorRoutine() {
	Log.Info("IP locator routine started")
	apiCallChan := make(chan locReq) // Give it a size of 20 just in case there are many inbound requests at the same time
	go func() {
		for {
			select {
			// New request to identify an IP
			case request := <-c.locationRequest:
				Log.Debug("new request has been received for ip:", request.ip)
				// Check if the IP is already in the cache
				cacheResp, ok := c.reqCache.checkIpInCache(request.ip)
				if ok { // the response was alreadi in the cache.
					Log.Debugf("ip %s was in cache", request.ip)
					response := *cacheResp
					// TODO: might be interesting to check if the error is due to an invalid IP
					// 		or if the reported error is due to a connection with the server (too many requests in the last minute)
					if response.err != nil {
						Log.Warn("readed response from cache includes an error:", response.err.Error())
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
				Log.Debugf("call %d-> ip %s not in cache, making API call", call, response.ip)
				for !breakCallLoop {
					// if req delay is setted to true, make new request
					// make the API call, and receive the apiResponse, the nextDelayRequest and the error from the connection
					response.resp, nextDelayRequest, response.err = c.getLocationFromIp(request.ip)
					if response.err != nil {
						if response.err.Error() == TooManyRequestError {
							// if the error reports that we tried too many calls on the API, sleep given time and try again
							Log.Debug("call", call, "-> error received:", response.err.Error(), "\nwaiting ", nextDelayRequest+(5*time.Second))
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
							Log.Debug("call", call, "-> diff error received:", response.err.Error())
							break
						}
					} else {
						Log.Debugf("call %d-> api req success", call)
						// if the error is different from TooManyRequestError break loop and store the request
						break
					}

				}
				// check if there is any waiting time that we have to respect before next connection
				if nextDelayRequest != time.Duration(0) {
					Log.Debugf("call %d-> number of allowed requests has been exceed, waiting %#v", call, nextDelayRequest+(5*time.Second))
					// set req delay to true, noone can make requests
					time.Sleep(nextDelayRequest + (5 * time.Second))
				}

				Log.Debugf("call %d-> saving new request and return it")
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
func (c *PeerLocalizer) LocateIP(ip string) (IpApiMessage, error) {
	// generate a new locRequest
	req := newLocReq(ip)
	// close opened channel at the end of the function
	defer close(req.respChan)
	// put the request in the Queue
	c.locationRequest <- req

	// wait until the opened response channel receives the response
	response := <-req.respChan
	// check content of the response
	if response.err != nil {
		err := errors.Wrap(response.err, "unable to locate IP"+response.ip)
		return IpApiMessage{}, err
	}
	return response.resp, nil
}

//
func (c *PeerLocalizer) Close() {
	Log.Info("closing ", ModuleName)
	// close the context for ending up the routine
	c.cancel()
}

// internal request between the external request and the internal locator routine
// includes a channel where to put the IpMsg and the received error
type locReq struct {
	ip       string
	respChan chan ipResponse
}

func newLocReq(ip string) locReq {
	return locReq{
		ip:       ip,
		respChan: make(chan ipResponse, 1), // give depth of 1 response to the channel
	}
}

// REQUES CACHE LIST

type requestCache struct {
	maxCacheLen int
	reqList     []*ipResponse
	reqMap      map[string]*ipResponse
}

// Generate new requestCache defining the max len that this one can take
func newRequestCache(maxLen int) requestCache {
	return requestCache{
		maxCacheLen: maxLen,
		reqList:     make([]*ipResponse, 0),
		reqMap:      make(map[string]*ipResponse),
	}
}

// checkIpInCache returns the location for the given IP if it is in cache, !ok if the ip is not in the cache
func (c *requestCache) checkIpInCache(ip string) (*ipResponse, bool) {
	resp, ok := c.reqMap[ip]
	return resp, ok
}

// addRequest includes a new IP and its location into the requestCache.
// If the cache is its limit, remove first request from the list and
// store the new one
func (c *requestCache) addRequest(response *ipResponse) error {
	Log.Debug("adding new response to the cache from IP:" + response.ip)
	// check if the IP is empty
	if response.ip == "" {
		return errors.New("the given IP is empty")
	}
	// check if the cache is full
	if c.len() >= c.maxCacheLen {
		// if the cache is full, remove the first item from the list and from the map
		c.delFirstRequest()
	}
	// add the new item to the last position of the array and to the map
	c.reqList = append(c.reqList, response)
	c.reqMap[response.ip] = response
	return nil
}

// getFirstRequest returns pointer to the first ipResponse located in the cache, nil if list is empty
func (c *requestCache) getFirstRequest() *ipResponse {
	if c.len() <= 0 {
		return nil
	}
	return c.reqList[0]
}

// delFirstRequest deletes the first item from the array and from the map respectively
func (c *requestCache) delFirstRequest() {
	// if the cache is full, remove the first item from the list and from the map
	fReq := c.getFirstRequest()
	// del from the map
	delete(c.reqMap, fReq.ip)
	// copy last array without first item into c.reqList
	c.reqList = c.reqList[1:]
}

// len returns the len of the cache array
func (c *requestCache) len() int {
	return len(c.reqList)
}

type ipResponse struct {
	ip   string
	err  error
	resp IpApiMessage
}

// IP-API message structure
type IpApiMessage struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
}

// get location country and City from the multiaddress of the peer on the peerstore
func (c *PeerLocalizer) getLocationFromIp(ip string) (apiMsg IpApiMessage, delayTime time.Duration, retErr error) {
	url := ipApiEndpoint + ip

	// Make the IP-APi request
	resp, err := http.Get(url)
	// increment api calls counter
	atomic.AddInt32(c.apiCalls, 1)
	if err != nil {
		retErr = errors.Wrap(err, "unable to locate IP"+ip)
		return
	}
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])
	// check if the error that we are receiving means that we exeeded the request limit
	if resp.StatusCode == 429 {
		Log.Warnf("limit of requests per minute has been exeeded, wait for next call %s secs", resp.Header["X-Ttl"][0])
		retErr = errors.New(TooManyRequestError)
		delayTime = time.Duration(timeLeft) * time.Second
		return
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
		return
	}

	// Convert response body to struct
	err = json.Unmarshal(bodyBytes, &apiMsg)
	if err != nil {
		retErr = errors.Wrap(err, "could not unmarshall response")
		return
	}
	// Check if the status of the request has been succesful
	if apiMsg.Status != "success" {
		retErr = errors.New(fmt.Sprintf("status from ip different than success, resp header:\n %#v", *resp))
		return
	}
	return
}
