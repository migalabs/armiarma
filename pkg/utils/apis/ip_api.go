package apis

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ip2location/ip2proxy-go/v4"
	"github.com/migalabs/armiarma/pkg/db/models"
	log "github.com/sirupsen/logrus"
)

var (
	ErrorQueueFull  = errors.New("queue is full")
	ErrorQueueEmpty = errors.New("queue is emtpy")
)

const (
	defaultIpTTL   = 30 * 24 * time.Hour // 30 days
	ipChanBuffSize = 45                  // number of ips that can be buffered unto the channel
	ipBuffSize     = 8192                // number of ip queries that can be queued in the ipQueue
	inApiEndpoint  = "https://www.ip2location.com/download/?token=%s&file=%s"
	minIterTime    = 100 * time.Millisecond
)

// DB Interface for DBWriter
type DBWriter interface {
	PersistToDB(interface{})
	ReadIpInfo(string) (models.IpInfo, error)
	CheckIpRecords(string) (bool, bool, error)
	GetExpiredIpInfo() ([]string, error)
}

type ipQueue struct {
	sync.RWMutex
	queueSize int
	ipList    []string
}

// PEER LOCALIZER
type IpLocator struct {
	ctx context.Context
	// Request channel	s
	locationRequest chan string

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

func newIpQueue(queueSize int) *ipQueue {
	return &ipQueue{
		queueSize: queueSize,
		ipList:    make([]string, 0, queueSize),
	}
}

// ----------------------------------------------------------- //
// ------------------ DB UPDATE UTILITIES -------------------- //
// ----------------------------------------------------------- //

const (
	DatabaseDir          = "./database/"
	IP2LocationToken     = "IP2LOCATION_TOKEN"
	InApiEndpoint        = "https://www.ip2location.com/download/?token=%s&file=%s"
	UpdateThreshold      = 24 * time.Hour
	IPv4DbName           = "PX11LITEBIN"
	IPv6DbName           = "PX11LITEBINIPV6"
	UncompressedFileName = "IP2LOCATION-LITE-DB11.BIN"
)

func unzip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadAndSave(url, baseFilename string) error {
	version := func() string {
		if strings.Contains(baseFilename, "IPV6") {
			return "IPv6"
		}
		return "IPv4"
	}()

	timestamp := time.Now().Format("20060102-150405") // Format: YYYYMMDD-HHMMSS

	filename := fmt.Sprintf("%s-%s.BIN", baseFilename, timestamp)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("Starting download of IP2Location DB for " + version)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error while downloading IP2Location DB for " + version)
		return err
	}
	fmt.Println("Download completed for IP2Location DB for " + version)

	return nil
}

func updateSpecificDb(dbName, dbToken string) {
	dbLink := fmt.Sprintf(InApiEndpoint, dbToken, dbName)
	if needsUpdate(dbName) {
		if err := downloadAndSave(dbLink, dbName); err != nil {
			log.Printf("Failed to update database %s: %v\n", dbName, err)
		}
	}
	cleanupOldDatabases(dbName)
}

// checks if the database needs to be updated
func needsUpdate(baseFilename string) bool {
	fmt.Println("Ip2Location DB: checking time since last update...")
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Finished checking directory...")

	latest := time.Time{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), baseFilename) && strings.HasSuffix(f.Name(), ".BIN") {
			nameParts := strings.Split(f.Name(), "-")
			if len(nameParts) >= 2 {
				timestampPart := nameParts[len(nameParts)-1]
				timestampPart = strings.TrimSuffix(timestampPart, ".BIN")
				fileTime, err := time.Parse("20060102-150405", timestampPart)
				if err == nil && fileTime.After(latest) {
					latest = fileTime
				}
			}
		}
	}

	return time.Since(latest) > UpdateThreshold
}

// finds the latest database file in the directory to see if it's older than 24 hours
func findLatestDbFile(baseFilename string) string {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}

	latestFile := ""
	latest := time.Time{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), baseFilename) && strings.HasSuffix(f.Name(), ".BIN") {
			nameParts := strings.Split(f.Name(), "-")
			if len(nameParts) >= 2 {
				timestampPart := nameParts[len(nameParts)-1]
				timestampPart = strings.TrimSuffix(timestampPart, ".BIN")
				fileTime, err := time.Parse("20060102-150405", timestampPart)
				if err == nil && fileTime.After(latest) {
					latest = fileTime
					latestFile = f.Name()
				}
			}
		}
	}

	return latestFile
}

func cleanupDatabases(dbNames ...string) {
	for _, dbName := range dbNames {
		cleanupOldDatabases(dbName)
	}
}

func cleanupOldDatabases(baseFilename string) {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}

	timestampToFile := make(map[time.Time]string)

	var latest time.Time

	for _, f := range files {
		if strings.HasPrefix(f.Name(), baseFilename) && strings.HasSuffix(f.Name(), ".BIN") {
			nameParts := strings.Split(f.Name(), "-")
			if len(nameParts) >= 2 {
				timestampPart := nameParts[len(nameParts)-1]
				timestampPart = strings.TrimSuffix(timestampPart, ".BIN")
				fileTime, err := time.Parse("20060102-150405", timestampPart)
				if err != nil {
					log.Printf("Failed to parse time from filename '%s': %s\n", f.Name(), err)
					continue
				}

				timestampToFile[fileTime] = f.Name()

				if fileTime.After(latest) {
					latest = fileTime
				}
			}
		}
	}

	for t, name := range timestampToFile {
		if t.Before(latest) {
			err := os.Remove(filepath.Join(DatabaseDir, name))
			if err != nil {
				log.Printf("Failed to remove old database file: %s\n", name)
			} else {
				log.Printf("Removed old database file: %s\n", name)
			}
		}
	}
}

func getDatabaseFile(ip string) string {
	var baseFilename string
	if isIPv4(ip) {
		baseFilename = "PX11LITEBIN"
	} else {
		baseFilename = "PX11LITEBINIPV6"
	}

	latestFile := findLatestDbFile(baseFilename)
	if latestFile == "" || needsUpdate(latestFile) {
		updateDb()
		latestFile = findLatestDbFile(baseFilename)
	}

	return latestFile
}

func updateDb() error {
	dbToken := os.Getenv(IP2LocationToken)
	if dbToken == "" {
		return errors.New("IP2LOCATION_TOKEN environment variable not set")
	}

	IPv4DbLink := fmt.Sprintf(inApiEndpoint, dbToken, IPv4DbName)
	IPv6DbLink := fmt.Sprintf(inApiEndpoint, dbToken, IPv6DbName)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if needsUpdate(IPv4DbName) {
			if err := downloadAndSave(IPv4DbLink, IPv4DbName); err != nil {
				log.Println(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if needsUpdate(IPv6DbName) {
			if err := downloadAndSave(IPv6DbLink, IPv6DbName); err != nil {
				log.Println(err)
			}
		}
	}()

	wg.Wait()

	cleanupDatabases(IPv4DbName, IPv6DbName)

	return nil
}

// ------------------------------------------------- //

func isIPv4(ip string) bool {
	ipv4Pattern := `^(\d{1,3}\.){3}\d{1,3}$`
	match, _ := regexp.MatchString(ipv4Pattern, ip)
	return match
}

func isIPv6(ip string) bool {
	ipv6Pattern := `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`
	match, _ := regexp.MatchString(ipv6Pattern, ip)
	return match
}

// ------------------------------------------------- //

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

// add an item to the IP queue
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

// reads items from the IP queue
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

// this function replaces the API call in the old version of the script
func locate(ip string) (ip2proxy.IP2ProxyRecord, error) {
	if !isIPv4(ip) && !isIPv6(ip) {
		return ip2proxy.IP2ProxyRecord{}, fmt.Errorf("invalid IP address")
	}

	dbFile := getDatabaseFile(ip)
	//todo: unzip db file and clean up and use the new name of the db
	db, err := ip2proxy.OpenDB(DatabaseDir + dbFile)
	if err != nil {
		return ip2proxy.IP2ProxyRecord{}, err
	}
	defer db.Close()

	results, err := db.GetAll(ip)
	if err != nil {
		return ip2proxy.IP2ProxyRecord{}, err
	}

	return results, err
}

func (c *IpLocator) locatorRoutine() {
	go func() {
		ticker := time.NewTicker(minIterTime)
		for {
			ip, err := c.ipQueue.readItem()
			if err == nil {
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

	go func() {
		for {
			select {
			case ip := <-c.locationRequest:
				respC := c.locateIp(ip)
			case <-c.ctx.Done():
				return
			}
		}
	}()
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

	var tempInfo ip2proxy.IP2ProxyRecord
	tempInfo, err = locate(ip)
	if err != nil {
		return
	}

	var apiMsg models.IpApiMsg
	apiMsg = models.mapTempIpInfoToApiMsg(tempInfo, ip)

	ipInfo.ExpirationTime = time.Now().UTC().Add(defaultIpTTL)
	ipInfo.IpApiMsg = apiMsg
	return
}

// ------------------------------------------------- //

func (c *IpLocator) Run() {
	//l.SetLevel(Logrus.TraceLevel)
	c.locatorRoutine()
}

func (c *IpLocator) Close() {
	log.Info("closing IP-API service")
	// close the context for ending up the routine
	c.ctx.Done()

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
