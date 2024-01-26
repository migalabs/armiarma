package apis

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	DatabaseDir           = "./database/"
	IP2LocationToken      = "IP2LOCATION_TOKEN"
	DBDownloadApiEndpoint = "https://www.ip2location.com/download/?token=%s&file=%s"
	UpdateThreshold       = 24 * time.Hour
	IPDbName              = "PX11LITEBIN"
	UncompressedFileName  = "IP2PROXY-LITE-PX11.BIN"
	TimestampFormat       = "20060102-150405"
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

func downloadAndSaveZippedDB() error {
	dbToken := os.Getenv(IP2LocationToken)
	if dbToken == "" {
		return errors.New("IP2LOCATION_TOKEN environment variable not set")
	}

	url := fmt.Sprintf(DBDownloadApiEndpoint, dbToken, IPDbName)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contentDisposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return err
	}

	filename := params["filename"]
	if filename == "" {
		filename = "PX11LITEBIN.zip"
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("Starting download of IP2Proxy DB to %s\n", filename)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Error while downloading IP2Proxy DB to %s\n", filename)
		return err
	}
	fmt.Printf("Download completed for IP2Proxy DB to %s\n", filename)

	return nil
}

// checks if the database needs to be updated
func needsUpdate() bool {
	fmt.Println("Ip2Location DB: checking time since last update...")
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}

	latest := time.Time{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "IP2PROXY-LITE-PX11") && strings.HasSuffix(f.Name(), ".BIN") {
			timestampPart := strings.TrimSuffix(f.Name(), ".BIN")
			timestampPart = strings.TrimPrefix(timestampPart, "IP2PROXY-LITE-PX11")
			fileTime, err := time.Parse(TimestampFormat, timestampPart)
			if err == nil && fileTime.After(latest) {
				latest = fileTime
			}
		}
	}
	fmt.Println("Finished checking directory...")

	return latest.IsZero() || time.Since(latest) > UpdateThreshold
}

// finds the latest database file in the directory to see if it's older than 24 hours
func findLatestDbFile() string {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}

	latestFile := ""
	latest := time.Time{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), UncompressedFileName) && strings.HasSuffix(f.Name(), ".BIN") {
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

func cleanupOldDatabases() error {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		return err
	}

	if len(files) <= 1 {
		return nil
	}

	timestampToFile := make(map[time.Time]string)
	var latestTime time.Time

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "IP2PROXY-LITE-PX11-") && strings.HasSuffix(file.Name(), ".BIN") {
			nameParts := strings.Split(file.Name(), "-")
			if len(nameParts) >= 3 {
				timestampPart := strings.TrimSuffix(nameParts[len(nameParts)-1], ".BIN")
				fileTime, err := time.Parse("20060102150405", timestampPart)
				if err != nil {
					log.Printf("Failed to parse time from filename '%s': %s\n", file.Name(), err)
					continue
				}

				timestampToFile[fileTime] = file.Name()
				if fileTime.After(latestTime) {
					latestTime = fileTime
				}
			}
		}
	}

	for fileTime, fileName := range timestampToFile {
		if fileTime.Before(latestTime) {
			err := os.Remove(filepath.Join(DatabaseDir, fileName))
			if err != nil {
				log.Printf("Failed to remove old database file: %s\n", fileName)
			} else {
				log.Printf("Removed old database file: %s\n", fileName)
			}
		}
	}

	return nil
}

func cleanupFolder() error {
	targetFiles := []string{"LICENSE", "README", ".zip"}

	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		shouldRemove := false
		for _, target := range targetFiles {
			if strings.Contains(file.Name(), target) {
				if file.Name() == "IP2PROXY-LITE-PX11.BIN" || strings.Contains(file.Name(), "IP2PROXY") {
					shouldRemove = false
				} else {
					shouldRemove = true
				}
				break
			}
		}

		if shouldRemove {
			err := os.Remove(filepath.Join(DatabaseDir, file.Name()))
			if err != nil {
				log.Printf("Failed to remove file: %s\n", file.Name())
			} else {
				log.Printf("Removed file: %s\n", file.Name())
			}
		}
	}

	return nil
}

func isNumeric(s string) bool {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

func verifyDbFile() (bool, error) {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		return false, err
	}
	if len(files) == 0 {
		return false, errors.New("No database file found")
	}

	// Check if "PX11LITEBIN.zip" file exists
	found := false
	for _, file := range files {
		if file.Name() == "PX11LITEBIN.zip" {
			found = true
			break
		}
	}

	if !found {
		return false, errors.New("File PX11LITEBIN.zip not found")
	}

	// unzip file and check validity of db
	err = unzip("PX11LITEBIN.zip", DatabaseDir)
	if err != nil {
		return false, err
	}
	dbPath := filepath.Join(DatabaseDir, "IP2PROXY-LITE-PX11.BIN") // name of the file after unzipping
	db, err := ip2proxy.OpenDB(dbPath)
	version := db.DatabaseVersion()
	if version == "" || !isNumeric(version) {
		return false, errors.New("Invalid database version")
	}
	return true, nil
}

func renameDbFile() error {
	files, err := ioutil.ReadDir(DatabaseDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("No database file found")
	}
	timeStamp := time.Now().Format(TimestampFormat)
	for _, file := range files {
		if file.Name() == "IP2PROXY-LITE-PX11.BIN" {
			err := os.Rename(filepath.Join(DatabaseDir, file.Name()), filepath.Join(DatabaseDir, "IP2PROXY-LITE-PX11-"+timeStamp+".BIN"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateDb() error {
	dbToken := os.Getenv(IP2LocationToken)
	if dbToken == "" {
		return errors.New("IP2LOCATION_TOKEN environment variable not set")
	}
	if err := downloadAndSaveZippedDB(); err != nil {
		log.Printf("Failed to download DB: %v\n", err)
		return err
	}
	valid, err := verifyDbFile()
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("Invalid database file")
	}

	err = renameDbFile()
	if err != nil {
		return err
	}

	cleanupOldDatabases() // removes all the old versions of the db file and leaves the one with the latest timestamp
	cleanupFolder()       // cleans the folder from all the other files that are not needed (zip, readme, license)

	return nil
}

// NEW VERSION
func getDatabaseFile() string {
	latestFile := findLatestDbFile()

	if latestFile == "" || needsUpdate() {
		updateDb()
	}
	return findLatestDbFile()

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

// some fields are commented because we don't have data for them
func mapTempIpInfoToApiMsg(data ip2proxy.IP2ProxyRecord, ip string) models.IpApiMsg {
	return models.IpApiMsg{
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

	dbFile := getDatabaseFile()
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
				resp := <-respC
				if resp.Err != nil {
					log.Error("error while locating IP -", resp.Err.Error())
					continue
				}
				if resp.IpInfo.IsEmpty() {
					log.Error("empty response from IP-API")
					continue
				}
				c.dbClient.PersistToDB(resp.IpInfo)
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
	apiMsg = mapTempIpInfoToApiMsg(tempInfo, ip)

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
