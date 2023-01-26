package peering

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/db/peerstore"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/hosts"
	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
)

var (
	// Default Delays
	DeprecationTime       = 1024 * time.Minute // mMinutes after first negative connection that has to pass to deprecate a peer.
	DefaultNegDelay       = 12 * time.Hour     // Default delay that will be applied for those deprecated peers.
	DefaultPossitiveDelay = 2 * time.Hour      // Default delay after each positive severe negative attempts.
	StartExpD             = 2 * time.Minute    // Starting delay that will serve for the Exponential Delay.
	// Control variables
	MinIterTime = 5 * time.Second // Minimum time that has to pass before iterating again.

)

// Pruning Strategy is a Peering Strategy that applies penalties to peers that haven't shown activity when attempting to connect them.
// Combined with the Deprecated flag in the models.Peer struct, it produces more accurate metrics when exporting pruning peers that are no longer active.
type PruningStrategy struct {
	ctx context.Context

	network      utils.NetworkType
	strategyType string
	Peerstore    *peerstore.Peerstore
	DBClient     *psql.DBClient

	// Peer Stream and Return ConnectionStatus channels (communication between modules)
	// both empty by default (need for initialization)
	peerStreamChan chan *models.HostInfo
	nextPeerChan   chan struct{}
	connAttemptNot chan *models.ConnectionAttempt
	connEventNot   chan *models.EventTrace
	identEventNot  chan hosts.IdentificationEvent

	// List of peers sorted by the amount of time thatwe have to wait
	PeerQueue PeerQueue

	// Prometheus Control Variables
	lastIterTime             time.Duration
	iterForcingNextConnTime  time.Time
	attemptedPeers           int64
	queueErroDistribution    sync.Map
	PeerQueueIterations      int
	ErrorAttemptDistribution sync.Map
}

// NewPruningStrategy is a constructor that will offer a models.Peer stream for the
// peering service. The provided models.Peer stream are ready to connect.d
func NewPruningStrategy(
	ctx context.Context,
	network utils.NetworkType,
	localPeerstorePath string,
	dbClient *psql.DBClient) (*PruningStrategy, error) {

	// generate a local peerstore for addrBook
	// TODO: move this to the crawler? -> makes more logic to me here
	pstore := peerstore.NewPeerstore(localPeerstorePath)

	// TODO: consider making the ConnStatus channel larger
	return &PruningStrategy{
		ctx:            ctx,
		network:        network,
		strategyType:   "pruning",
		Peerstore:      pstore,
		DBClient:       dbClient,
		PeerQueue:      NewPeerQueue(),
		peerStreamChan: make(chan *models.HostInfo, DefaultWorkers),
		nextPeerChan:   make(chan struct{}, DefaultWorkers),
		connAttemptNot: make(chan *models.ConnectionAttempt),
		connEventNot:   make(chan *models.EventTrace),
		identEventNot:  make(chan hosts.IdentificationEvent),
		// Metrics Variables
	}, nil
}

// Type returns the strategy type that has been set.
func (c PruningStrategy) Type() string {
	return c.strategyType
}

// Run initializes the models.Peer stream on the returning models.Peer chan
// stores locally an auxiliary map wuth an array that will keep
// track of the next connection time.
func (c *PruningStrategy) Run() chan *models.HostInfo {
	// start go routine that will notify of the full peerstore iteration and notifies it to the main strategy loop
	go c.peerstoreIteratorRoutine()
	go c.eventRecorderRoutine()

	return c.peerStreamChan
}

// peerstoreIterator private function that is in charge of iterating through the peerstore,
// receive connections/disconnections, and fetch info comming from the peering service into the db.
// Main interaction of the Peering Service with the DB.
func (c *PruningStrategy) peerstoreIteratorRoutine() {
	logEntry := log.WithFields(log.Fields{
		"mod": "prun-strgy-itr",
	})
	logEntry.Debug("init")
	c.PeerQueueIterations = 0

	// get the peer list from the peerstore
	err := c.PeerQueue.UpdatePeerListFromRemoteDB(c.DBClient, c.Peerstore)
	if err != nil {
		log.Error(err)
	}

	c.queueErroDistribution.Store(PositiveDelayType, 0)
	c.queueErroDistribution.Store(NegativeWithHopeDelayType, 0)
	c.queueErroDistribution.Store(NegativeWithNoHopeDelayType, 0)
	c.queueErroDistribution.Store(Minus1DelayType, 0)
	c.queueErroDistribution.Store(ZeroDelayType, 0)
	c.queueErroDistribution.Store(TimeoutDelayType, 0)

	peerCounter := 0
	peerListLen := c.PeerQueue.Len()
	validIterTimer := time.NewTimer(MinIterTime)
	iterStartTime := time.Now()
	nextIterFlag := false
	for {
		select {
		// Receive the notification of sending the next peer
		case <-c.nextPeerChan:
			if peerListLen > 0 {
				logEntry.Trace("prepare next peer for pushing it into peer stream")
				// read info about next peer
				nextPeer := c.PeerQueue.PeerList[peerCounter]

				// check if the node is ready for connection
				// or we are in the first iteration, then we always try all of them
				if nextPeer.IsReadyForConnection() || c.PeerQueueIterations == 0 {
					pinfo, ok := c.Peerstore.LoadPeer(nextPeer.PeerID)
					if !ok {
						// Really unlikely to happen but try to retrieve the peer info from the DB
						pinfo, err = c.DBClient.GetPersistable(peer.ID(nextPeer.PeerID))
						if err != nil {
							log.Warn(err)
							continue
						}
					}

					// Send next peer to the peering service
					logEntry.Tracef("pushing next peer %s into peer stream", pinfo.ID)

					v, _ := c.queueErroDistribution.Load(nextPeer.DelayObj.GetType())
					val := v.(int)
					c.queueErroDistribution.Store(nextPeer.DelayObj.GetType(), val+1)
					// we need to send the hInfo of the peer - compose it from the persistable peer
					hInfo := models.NewHostInfo(
						pinfo.ID,
						pinfo.Network,
						models.WithMultiaddress(pinfo.Addrs),
					)
					c.peerStreamChan <- hInfo

					// increment peerCounter to see if we finished iterating the peerstore
					peerCounter++

				} else {
					logEntry.Trace("next peers has to wait to be connected")
					c.iterForcingNextConnTime = nextPeer.NextConnection()

					c.NextPeer()
					nextIterFlag = true
				}
			} else {
				logEntry.Warn("empty peerstore")
				// Recreate the call of the nextPeer that the iterator just used
				c.NextPeer()

			}
			if nextIterFlag || peerCounter >= peerListLen {
				// time to update the PeerList
				c.lastIterTime = time.Since(iterStartTime)
				atomic.StoreInt64(&c.attemptedPeers, int64(peerCounter))
				logEntry.Debug("peerstore iteration of ", peerCounter, " peers, done in ", c.lastIterTime)
				logEntry.Debug("missing ", c.PeerQueue.Len()-peerCounter, " peers waiting for next try")

				// check if the minIterTime has been
				<-validIterTimer.C

				// reset values
				c.queueErroDistribution = ResetMapValues(c.queueErroDistribution)

				// get the peer list from the peerstore
				err := c.PeerQueue.UpdatePeerListFromRemoteDB(c.DBClient, c.Peerstore)
				c.PeerQueueIterations++ // another iteration
				if err != nil {
					log.Error(err)
				}

				peerListLen = c.PeerQueue.Len()
				logEntry.Debugf("got new peer list with %d", peerListLen)
				validIterTimer = time.NewTimer(MinIterTime)
				iterStartTime = time.Now()
				peerCounter = 0
				nextIterFlag = false
			}

		// detect if the context has been shut down to end the go routine
		case <-c.ctx.Done():
			logEntry.Debug("closing")
			close(c.peerStreamChan)
			close(c.nextPeerChan)
			close(c.connEventNot)
			return
		}
	}
}

// peerstoreIterator is a private function that is in charge of iterating through the peerstore,
// receive connections/disconnections, and fetch info comming from the peering service into the db.
// Main interaction of the Peering Service with the DB.
func (c *PruningStrategy) eventRecorderRoutine() {
	logEntry := log.WithFields(log.Fields{
		"mod": "prun-evnt-rec",
	})
	logEntry.Debugf("init")

	// Buffer of connection Events peer peer (map because of lack of concurrency)
	connEventBuffer := make(map[peer.ID]*models.ConnEvent, 0)

	for {
		select {
		// Receive the status of the peer that got connected to the crawler
		case connAttempt := <-c.connAttemptNot:
			logEntry.Tracef("new connection attempt has been received from peer %s", connAttempt.RemotePeer)
			// update the local info about the peer
			p, ok := c.PeerQueue.GetPeer(connAttempt.RemotePeer.String())
			if !ok {
				logEntry.Warnf("Could not find peer in peerqueue: %s", connAttempt.RemotePeer)
			}
			p.ConnEventHandler(connAttempt.Error)
			// Check if peer needs to be deprecated
			if p.Deprecable() {
				logEntry.Warnf("deprecating peer %s", connAttempt.RemotePeer.String())
				connAttempt.Deprecable = true
				// remove p from list of peers to ping (if it appears again in the discovery, it will be updated as undeprecated in the DB)
				c.PeerQueue.RemovePeer(connAttempt.RemotePeer.String())
			}

			c.DBClient.PersistToDB(connAttempt)

		// Receive a notification of a connection event
		case eventTrace := <-c.connEventNot:
			// check if we already have a connectionEvent for that Peer waiting for it pair to come
			bEvent, ok := connEventBuffer[eventTrace.PeerID]
			if !ok {
				// if there is no prev trace - create a new one
				bEvent = models.NewConnEvent(eventTrace.PeerID)
				connEventBuffer[eventTrace.PeerID] = bEvent
			}
			// check what event came in the trace and add it to the matching prev event
			switch eventTrace.Event.(type) {
			case (*models.ConnInfo):
				cInfo := eventTrace.Event.(*models.ConnInfo)
				bEvent.AddConnInfo(*cInfo)
			case (*models.EndConnInfo):
				endConnInfo := eventTrace.Event.(*models.EndConnInfo)
				bEvent.AddDisconn(*endConnInfo)
			default:
				logEntry.Warnf("invalid event trace for peer %s - %x\n", eventTrace.PeerID.String(), eventTrace.Event)
			}

			// check if the ConnEvent is ready to be persisted
			if bEvent.IsReadyToPersist() {
				logEntry.Debugf("persising full conn event for peer %s", bEvent.PeerID.String())
				c.DBClient.PersistToDB(bEvent)
			}

		case identEvent := <-c.identEventNot:
			// We received a new identification attempt
			// extract whether it was success or a failure
			// and track it in the PeerQueue and in the peerstore
			logEntry.Debugf("new identification from peer %s", identEvent.HostInfo.ID.String())

			// by default we think the identify was not successful, therefore negativewithhope and error
			var delayType string = NegativeWithHopeDelayType

			if identEvent.HostInfo.IsHostIdentified() {
				delayType = PositiveDelayType
			}

			var p *PrunedPeer
			p, ok := c.PeerQueue.GetPeer(identEvent.HostInfo.ID.String())
			if !ok {
				p = NewPrunedPeer(identEvent.HostInfo.ID.String(), delayType)
				c.PeerQueue.AddPeer(p)
			}
			// double-check when are we rewriting hInfo without IP, and port
			if identEvent.HostInfo.IP == "" {
				logEntry.Error("error trying to add host info without IP and ports", identEvent.HostInfo)
			}
			c.DBClient.PersistToDB(identEvent.HostInfo)

		// detect if the context has been shut down to end the go routine
		case <-c.ctx.Done():
			logEntry.Debug("closing event recorder routine")
			return
		}
	}
}

// NextPeer notifies the peerstore iterator that a new peer has been requested.
// After it, the peerstore iterator will put the new peer in the PeerStreamChan.
func (c *PruningStrategy) NextPeer() {
	c.nextPeerChan <- struct{}{}
}

// NewConnectionAttempt notifies the peerstore iterator that a new ConnStatus has been received.
// After it, the peerstore iterator will aggregate the extra info.
func (c *PruningStrategy) NewConnectionAttempt(connAttStat *models.ConnectionAttempt) {
	c.connAttemptNot <- connAttStat
}

// NewConnectionEvent notifies the peerstore iterator that a new Connection has been received.
// It puts the connection metadata in the connNot channel to let the select
// loop all the metadata of the received connection.
func (c *PruningStrategy) NewConnectionEvent(eventTrace *models.EventTrace) {
	c.connEventNot <- eventTrace
}

// NewIdentificationEvent will insert a new identification item in the identificationeventnorifier channel.
func (c *PruningStrategy) NewIdentificationEvent(newIdent hosts.IdentificationEvent) {
	c.identEventNot <- newIdent
}

// --------------------------------------------------
// Metrics Exporting Functions for Peering Prometheus
// --------------------------------------------------

// LastIterTime returns the lastiteration time of the peerqueue
func (c *PruningStrategy) LastIterTime() float64 {
	return float64(c.lastIterTime.Microseconds()) / 1000000
}

func (c *PruningStrategy) IterForcingNextConnTime() string {
	return c.iterForcingNextConnTime.String()
}

func (c *PruningStrategy) AttemptedPeersSinceLastIter() int64 {
	return atomic.LoadInt64(&c.attemptedPeers)
}

func (c *PruningStrategy) ControlDistribution() sync.Map {
	return c.PeerQueue.DelayDistribution()
}

func (c *PruningStrategy) GetErrorAttemptDistribution() sync.Map {
	return c.ErrorAttemptDistribution
}

// PeerQueue is an auxiliar peer array and map list to keep the list of peers sorted
// by connection time, and still able to modify in a short time the values of each peer.
type PeerQueue struct {
	PeerList []*PrunedPeer
	PeerMap  sync.Map
	// Metrics
	queueErroDistribution sync.Map
}

// DelayDistribution returns the distribution of the delays in a map.
func (c *PeerQueue) DelayDistribution() sync.Map {
	return c.queueErroDistribution
}

// NewPeerQueue is the constructor of a NewPeerQueue
func NewPeerQueue() PeerQueue {
	pq := PeerQueue{
		PeerList: make([]*PrunedPeer, 0),
	}
	return pq
}

// IsPeerAlready checks whether a peer is already in the Queue.
func (c *PeerQueue) IsPeerAlready(peerID string) bool {
	_, ok := c.PeerMap.Load(peerID)
	return ok
}

// AddPeer Adds a peer to the peerqueue.
func (c *PeerQueue) AddPeer(pPeer *PrunedPeer) {
	// append new item at the beginning of the array
	c.PeerList = append([]*PrunedPeer{pPeer}, c.PeerList...)
	c.PeerMap.Store(pPeer.PeerID, pPeer)
}

// RemovePeer()
func (c *PeerQueue) RemovePeer(pPeer string) {
	c.PeerMap.Delete(pPeer)
	var idx int = -1
	for index, pInfo := range c.PeerList {
		if pInfo.PeerID == pPeer {
			idx = index
			break
		}
	}
	if idx > -1 {
		c.PeerList = append(c.PeerList[:idx], c.PeerList[idx+1:]...)
	}
}

// GetPeer retrieves the info of the peer requested from args.
func (c *PeerQueue) GetPeer(peerID string) (*PrunedPeer, bool) {
	p, ok := c.PeerMap.Load(peerID)
	if !ok {
		return &PrunedPeer{}, ok
	}
	return p.(*PrunedPeer), ok
}

// SortPeerList sorts the PeerQueue array leaving at the beginning the peers
// with the shorter next peer connection.
func (c *PeerQueue) SortPeerList() {
	sort.Sort(c)
}

// ---  SORTING METHODS FOR PeerQueue ----

// Swap is part of sort.Interface.
func (c *PeerQueue) Swap(i, j int) {
	c.PeerList[i], c.PeerList[j] = c.PeerList[j], c.PeerList[i]
}

// Less is part of sort.Interface. We use c.PeerList.NextConnection as the value to sort by.
func (c PeerQueue) Less(i, j int) bool {
	return c.PeerList[i].NextConnection().Before(c.PeerList[j].NextConnection())
}

// Len is part of sort.Interface. We use the peer list to get the length of the array.
func (c PeerQueue) Len() int {
	return len(c.PeerList)
}

// UpdatePeerListFromPeerStore will refresh the peerqueue with the peerstore.
// Basically we add those peers that did not exist before in the peerqueue.
func (c *PeerQueue) UpdatePeerListFromRemoteDB(dbClient *psql.DBClient, localPeerstore *peerstore.Peerstore) error {
	// Get the list of peers from the peerstore
	peerList, err := dbClient.GetNonDeprecatedPeers()
	if err != nil {
		return errors.Wrap(err, "fail to update pruning peer queue from DBClient")
	}
	totcnt := 0
	new := 0

	// metrics
	c.queueErroDistribution.Store(PositiveDelayType, 0)
	c.queueErroDistribution.Store(NegativeWithHopeDelayType, 0)
	c.queueErroDistribution.Store(NegativeWithNoHopeDelayType, 0)
	c.queueErroDistribution.Store(Minus1DelayType, 0)
	c.queueErroDistribution.Store(ZeroDelayType, 0)
	c.queueErroDistribution.Store(TimeoutDelayType, 0)

	// Fill the PeerQueue.PeerList with the missing peers from the
	for _, peerID := range peerList {
		totcnt++
		if !c.IsPeerAlready(peerID.String()) {
			// Peer was not in the list of peers
			pInfo, ok := localPeerstore.LoadPeer(peerID.String())
			if !ok {
				// if peer not found locally, fetch data
				pInfo, err = dbClient.GetPersistable(peerID)
				if err != nil {
					log.Errorf("unable import peer to PeerQueue. %s\n", err.Error())
					continue
				}
				// add it locally
				localPeerstore.StorePeer(pInfo)
			}
			new++

			// Whenever we find a new peer that we didn't have locally, add zero delay
			// even when we read all the peerstore from the DB Endpoint when restarting
			delayDegree := 0             // default
			delayType := Minus1DelayType // default

			newPrunnedPeer := NewPrunedPeer(pInfo.ID.String(), delayType)
			newPrunnedPeer.DelayObj.SetDegree(delayDegree)

			// add the new item to the list
			c.AddPeer(newPrunnedPeer)
			v, _ := c.queueErroDistribution.Load(delayType)
			val := v.(int)
			c.queueErroDistribution.Store(delayType, val+1)

		} else {
			prunnedPeer, _ := c.GetPeer(peerID.String())
			v, _ := c.queueErroDistribution.Load(prunnedPeer.DelayObj.GetType())
			val := v.(int)
			c.queueErroDistribution.Store(prunnedPeer.DelayObj.GetType(), val+1)
		}
	}
	// Sort the list of peers based on the next connection
	c.SortPeerList()
	log.Debugf("Num of peers in PeerQueue: %d\n", c.Len())
	return nil
}

// TODO: think about including a sync.RWMutex in case we upgrade to workers
type PrunedPeer struct {
	PeerID                   string
	DelayObj                 DelayObject // define the delay to connect based on error
	BaseConnectionTimestamp  time.Time   // define the first event. To calculate the next connection we sum this with delay.
	BaseDeprecationTimestamp time.Time   // this + DeprecationTime defines when we are ready to deprecate
}

func NewPrunedPeer(peerID string, inputType string) *PrunedPeer {
	t := time.Now()
	pp := PrunedPeer{
		PeerID:                   peerID,
		DelayObj:                 ReturnAccordingDelayObject(inputType),
		BaseConnectionTimestamp:  t,
		BaseDeprecationTimestamp: t, // by default we set it now, so if no positive connection it will be deprecated in 24 hours since creation of this prunned peer
	}

	return &pp
}

// IsReadyForConnection evaluates if the given peer is ready to be connected.
func (c *PrunedPeer) IsReadyForConnection() bool {
	now := time.Now()
	// if we are not before the time, then we are either equal or after the connection time
	return !now.Before(c.NextConnection())
}

// NextConnection returns the time where the pPeer needs to be connected (based on previous connAttempts)
func (c *PrunedPeer) NextConnection() time.Time {

	if c.DelayObj.GetType() == Minus1DelayType { // in case of Minus1, this is new peer and we want it to connect as soon as possible
		return time.Time{}
	}

	if c.DelayObj.CalculateDelay() > MaxDelayTime {
		return c.BaseConnectionTimestamp.Add(MaxDelayTime)
	}

	// nextConnection should be from first event + the applied delay
	return c.BaseConnectionTimestamp.Add(c.DelayObj.CalculateDelay())
}

// Deprecable evaluates if the peer is in time to be deprecated.
func (c *PrunedPeer) Deprecable() bool {
	// if the difference between now and the BaseDeprecationTimestampo is more than the DeprecationTime, true
	if time.Now().Sub(c.BaseDeprecationTimestamp) >= DeprecationTime {
		return true
	}

	return false
}

// RecErrorHandler selects actuation method for each of the possible errors while actively dialing peers.
func (c *PrunedPeer) ConnEventHandler(recErr string) {
	c.UpdateDelay(ErrorToDelayType(recErr))
}

// NewEvent will reevaluate the delay in case of a new Positive or NegativeDelay happens
func (c *PrunedPeer) UpdateDelay(newDelayType string) {
	// if the delaytype is different, always refresh the object
	c.BaseConnectionTimestamp = time.Now()

	if c.DelayObj.GetType() != newDelayType {
		c.DelayObj = ReturnAccordingDelayObject(newDelayType)
	}

	// if there is a positive delay (success identify), then we update the deprecation time
	// therefore, we start counting from now to deprecate
	if c.DelayObj.GetType() == PositiveDelayType {
		c.BaseDeprecationTimestamp = time.Now()
	}

	c.DelayObj.AddDegree()
}

// ErrorToDelayType transforms an error into a DelayType.
func ErrorToDelayType(errString string) string {
	switch errString {
	case hosts.NoConnError:
		return PositiveDelayType

	case hosts.DialErrorConnectionResetByPeer,
		hosts.DialErrorConnectionRefused,
		hosts.DialErrorContextDeadlineExceeded,
		hosts.DialErrorBackOff,
		hosts.ErrorRequestingMetadta,
		"unknown":
		return NegativeWithHopeDelayType

	case hosts.DialErrorNoRouteToHost,
		hosts.DialErrorNetworkUnreachable,
		hosts.DialErrorPeerIDMismatch,
		hosts.DialErrorSelfAttempt,
		hosts.DialErrorNoGoodAddresses:
		return NegativeWithNoHopeDelayType

	case hosts.DialErrorIoTimeout:
		return TimeoutDelayType

	default:
		log.Tracef("Default Delay applied, error: %s\n", errString)
		return NegativeWithHopeDelayType
	}
}

// ResetMapValues iterates over a string int map and resets all values to 0.
func ResetMapValues(inputMap sync.Map) sync.Map {
	var outMap sync.Map
	inputMap.Range(func(key, value interface{}) bool {
		outMap.Store(key, 0)
		return true
	})
	return outMap
}
