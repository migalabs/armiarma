package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultIpTTL   = 30 * 24 * time.Hour // 30 days
	ipChanBuffSize = 45                  // number of ips that can be buffered unto the channel
	ipBuffSize     = 8192                // number of ip queries that can be queued in the ipQueue
	ipApiEndpoint  = "http://ip-api.com/json/{__ip__}?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query"
	minIterTime    = 100 * time.Millisecond
)

var TooManyRequestError error = fmt.Errorf("error HTTP 429")

// DB Interface for DBWriter
type DBWriter interface {
	PersistToDB(interface{})
	ReadIpInfo(string) (models.IpInfo, error)
	CheckIpRecords(string) (bool, bool, error)
	GetExpiredIpInfo() ([]string, error)
}

// PEER LOCALIZER
type IpLocator struct {
	ctx context.Context
	// Request channels
	locationRequest chan string

	// dbClient
	dbClient DBWriter

	ipQueue *ipQueue
	// control variables for IP-API request
	// Control flags from prometheus
	apiCalls *int32
}

func NewIpLocator(ctx context.Context, dbCli DBWriter) *IpLocator {
	calls := int32(0)
	return &IpLocator{
		ctx:             ctx,
		locationRequest: make(chan string, ipChanBuffSize),
		dbClient:        dbCli,
		apiCalls:        &calls,
		ipQueue:         newIpQueue(ipBuffSize),
	}
}

// Run the necessary routines to locate the IPs
func (c *IpLocator) Run() {
	//l.SetLevel(Logrus.TraceLevel)
	c.locatorRoutine()
}

// locatorRoutine is the main routine that will wait until an request to identify an IP arrives
// or if the routine gets canceled
func (c *IpLocator) locatorRoutine() {
	log.Info("IP locator routine started")
	// ip queue reading routine
	go func() {
		ticker := time.NewTicker(minIterTime)
		for {
			ip, err := c.ipQueue.readItem()
			if err == nil {
				// put the request in the Queue
				c.locationRequest <- ip
			}
			select {
			case <-ticker.C:
				ticker.Reset(minIterTime)

			case <-c.ctx.Done():
				return
			}
		}
	}()

	// ip locating routien
	go func() {
		var nextDelayRequest time.Duration
		for {
			select {
			// New request to identify an IP
			case reqIp := <-c.locationRequest:
				log.Trace("new request has been received for ip:", reqIp)
			reqLoop:
				for {
					// since it didn't exist or did expire, request the ip
					// new API call needs to be done
					log.Tracef(" making API call for %s", reqIp)
					atomic.AddInt32(c.apiCalls, 1)
					respC := c.locateIp(reqIp)
					select {
					case apiResp := <-respC:
						nextDelayRequest = apiResp.DelayTime
						log.WithFields(log.Fields{
							"delay":         nextDelayRequest,
							"attempts left": apiResp.AttemptsLeft,
						}).Debug("got response from IP-API request ")
						// check if there is an error
						switch apiResp.Err {
						case TooManyRequestError:
							// if the error reports that we tried too many calls on the API, sleep given time and try again
							log.Debug("call ", reqIp, " -> error received: ", apiResp.Err.Error(), "\nwaiting ", nextDelayRequest+(5*time.Second))
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
							// Upsert the IP into the db
							c.dbClient.PersistToDB(apiResp.IpInfo)
							break reqLoop

						default:
							log.Debug("call ", reqIp, " -> diff error received: ", apiResp.Err.Error())
							break reqLoop

						}

					case <-c.ctx.Done():
						log.Info("context closure has been detecting, closing IpApi caller")
						return
					}
				}
				// check if there is any waiting time that we have to respect before next connection
				if nextDelayRequest != time.Duration(0) {
					log.Debug("number of allowed requests has been exceed, waiting ", nextDelayRequest+(2*time.Second))
					// set req delay to true, noone can make requests
					ticker := time.NewTicker(nextDelayRequest + (2 * time.Second))
					select {
					case <-ticker.C:
						continue
					case <-c.ctx.Done():
						log.Info("context closure has been detecting, closing IpApi caller")
						return
					}
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
	// check first if IP is already in queue (to queue same ip)
	if c.ipQueue.ipExists(ip) {
		return
	}

	// Check if the IP is already in the cache
	exists, expired, err := c.dbClient.CheckIpRecords(ip)
	if err != nil {
		log.Error("unable to check if IP already exists -", err.Error()) // Should it be a Panic?
	}
	// if exists and it didn't expired, don't do anything
	if exists && !expired {
		return
	}

	// since it didn't exist or it is expired, locate it again
	ticker := time.NewTicker(1 * time.Second)
	// wait 1 sec because is the normal time to wait untill we can start querying again
	for {
		err := c.ipQueue.addItem(ip)
		if err == nil {
			break
		}
		<-ticker.C
		ticker.Reset(1 * time.Second)
		log.Debug("waiting to alocate a new IP request")
	}
	ticker.Stop()
}

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
	apiResponse.IpInfo, apiResponse.DelayTime, apiResponse.AttemptsLeft, apiResponse.Err = CallIpApi(ip)
	respC <- apiResponse
	// defer ^
}

func CallIpApi(ip string) (ipInfo models.IpInfo, delay time.Duration, attemptsLeft int, err error) {

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
		log.Debugf("limit of requests per minute has been exeeded, wait for next call %s secs", resp.Header["X-Ttl"][0])
		err = TooManyRequestError
		delay = time.Duration(timeLeft) * time.Second
		return
	}

	// Check the attempts left that we have to call the api
	attemptsLeft, _ = strconv.Atoi(resp.Header["X-Rl"][0])
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
		err = errors.New(fmt.Sprintf("status from ip different than success, resp header:\n %#v \n %+v", resp, apiMsg))
		return
	}

	ipInfo.ExpirationTime = time.Now().UTC().Add(defaultIpTTL)
	ipInfo.IpApiMsg = apiMsg
	return
}

func newIpQueue(queueSize int) *ipQueue {
	return &ipQueue{
		queueSize: queueSize,
		ipList:    make([]string, 0, queueSize),
	}
}

var (
	ErrorQueueFull  = errors.New("queue is full")
	ErrorQueueEmpty = errors.New("queue is emtpy")
)

type ipQueue struct {
	sync.RWMutex
	queueSize int
	ipList    []string
}

func (q *ipQueue) addItem(newItem string) error {
	q.Lock()
	defer q.Unlock()

	if q.len() >= q.queueSize {
		return ErrorQueueFull
	}

	// check if the ip is already in the list
	for _, ip := range q.ipList {
		if newItem == ip {
			return nil
		}
	}

	q.ipList = append(q.ipList, newItem)

	return nil
}

func (q *ipQueue) readItem() (string, error) {
	q.Lock()
	defer q.Unlock()

	var item string
	if q.len() <= 0 {
		return item, ErrorQueueEmpty
	}

	item = q.ipList[0]

	// remove after the item from the list
	q.ipList = append(q.ipList[:0], q.ipList[0+1:]...)

	return item, nil
}

func (q *ipQueue) ipExists(target string) bool {
	for _, ip := range q.ipList {
		if ip == target {
			return true
		}
	}
	return false
}

func (q *ipQueue) len() int {
	return len(q.ipList)
}

func (q *ipQueue) Len() int {
	q.RLock()
	defer q.RUnlock()

	return q.len()
}
