package peering

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/hosts"
	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

var (
	// Default Delays
	DeprecationTime = 4 * time.Hour   // mMinutes after first negative connection that has to pass to deprecate a peer.
	StartExpD       = 2 * time.Minute // Starting delay that will serve for the Exponential Delay.
	// Control variables
	MinIterTime = 5 * time.Second // Minimum time that has to pass before iterating again.
	//
	PruneStrategy = "pruning"
)

// Pruning Strategy is a Peering Strategy that applies penalties to peers that haven't shown activity when attempting to connect them.
// Combined with the Deprecated flag in the models.Peer struct, it produces more accurate metrics when exporting pruning peers that are no longer active.
type PruningStrategy struct {
	ctx context.Context

	network  utils.NetworkType
	DBClient *psql.DBClient

	// Peer Stream and Return ConnectionStatus channels (communication between modules)
	// both empty by default (need for initialization)
	peerStreamChan chan *models.HostInfo
	nextPeerChan   chan struct{}
	connAttemptNot chan *models.ConnectionAttempt
	connEventNot   chan *models.EventTrace
	identEventNot  chan hosts.IdentificationEvent

	// List of peers sorted by the amount of time thatwe have to wait
	PeerQueue *PeerQueue

	// Prometheus Control Variables
	m              sync.RWMutex
	lastIterTime   time.Duration
	attemptedPeers map[Delay]int64
}

// NewPruningStrategy is a constructor that will offer a models.Peer stream for the
// peering service. The provided models.Peer stream are ready to connect.d
func NewPruningStrategy(
	ctx context.Context,
	network utils.NetworkType,
	dbClient *psql.DBClient) (*PruningStrategy, error) {

	return &PruningStrategy{
		ctx:            ctx,
		network:        network,
		DBClient:       dbClient,
		PeerQueue:      NewPeerQueue(dbClient),
		peerStreamChan: make(chan *models.HostInfo, DefaultWorkers),
		nextPeerChan:   make(chan struct{}, DefaultWorkers),
		connAttemptNot: make(chan *models.ConnectionAttempt),
		connEventNot:   make(chan *models.EventTrace),
		identEventNot:  make(chan hosts.IdentificationEvent),
		attemptedPeers: make(map[Delay]int64, 0),
	}, nil
}

// Type returns the strategy type that has been set.
func (c PruningStrategy) Type() string {
	return PruneStrategy
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

// ResetMapValues iterates over a string int map and resets all values to 0.
func (c *PruningStrategy) composeDelayDistFromAttemptedPeers(prunedPeers map[peer.ID]*PrunedPeer) {
	c.m.Lock()
	defer c.m.Unlock()

	// reset the delay list
	c.attemptedPeers = make(map[Delay]int64, 0)

	for _, pPeer := range prunedPeers {
		_, ok := c.attemptedPeers[pPeer.delayObj.dtype]
		if !ok {
			c.attemptedPeers[pPeer.delayObj.dtype] = int64(0)
		}
		c.attemptedPeers[pPeer.delayObj.dtype]++
	}
}

// peerstoreIterator private function that is in charge of iterating through the peerstore,
// receive connections/disconnections, and fetch info comming from the peering service into the db.
// Main interaction of the Peering Service with the DB.
func (c *PruningStrategy) peerstoreIteratorRoutine() {
	logEntry := log.WithFields(log.Fields{
		"mod": "prun-strgy-itr",
	})
	logEntry.Debug("init")

	// get the peer list from the peerstore
	err := c.PeerQueue.UpdatePeerListFromRemoteDB()
	if err != nil {
		log.Error(err)
	}

	validIterTimer := time.NewTimer(MinIterTime)
	iterStartTime := time.Now()
	callForPeer := false
	attemptedPeers := make(map[peer.ID]*PrunedPeer)

	for {
		select {
		// Receive the notification of sending the next peer
		case <-c.nextPeerChan:
			// check is the PeerQueue is not empty and if the next peer is valid
			if !c.PeerQueue.IsEmpty() && c.PeerQueue.ValidNextPeer() {
				logEntry.Trace("prepare next peer for pushing it into peer stream")
				// read info about next peer
				nextPeer := c.PeerQueue.GetNextPeer()

				// double check that we didn't attempt the peer in the same iteration
				_, ok := attemptedPeers[nextPeer.iD]
				if ok {
					log.Warnf("we already attempted to connect peer %s in the same iteration", nextPeer.iD.String())
					callForPeer = true
					goto pointerCheck
				}

				// add peer to the list of peers attempted in the last iter
				attemptedPeers[nextPeer.iD] = nextPeer

				// we need to send the hInfo of the peer - compose it from the persistable peer
				hInfo := models.NewHostInfo(
					nextPeer.iD,
					nextPeer.network,
					models.WithMultiaddress(nextPeer.addr),
				)

				// Send next peer to the peering service
				logEntry.Tracef("pushing next peer %s into peer stream", nextPeer.iD.String())
				c.peerStreamChan <- hInfo

			} else {
				// check the cause of the failure:
				if c.PeerQueue.IsEmpty() {
					logEntry.Warn("forcing peerstore update because it's empty - len=", c.PeerQueue.Len())

				}
				// check the cause of the failure:
				if c.PeerQueue.ValidNextPeer() {
					logEntry.Warn("forcing peerstore update because next peer is not ready to be connected")
				}
				// time to update the PeerList
				c.lastIterTime = time.Since(iterStartTime)
				logEntry.Debug("peerstore iteration of ", len(c.attemptedPeers), " peers, done in ", c.lastIterTime)

				// check if the minIterTime has been
				<-validIterTimer.C

				// save attempted peers' values and reset the map
				c.composeDelayDistFromAttemptedPeers(attemptedPeers)
				attemptedPeers = make(map[peer.ID]*PrunedPeer)

				// reset the pointer
				c.PeerQueue.ResetPeerPointer()

				// get the peer list from the peerstore
				err := c.PeerQueue.UpdatePeerListFromRemoteDB()
				if err != nil {
					log.Error(err)
				}

				logEntry.Debugf("got new peer list with %d", c.PeerQueue.Len())
				validIterTimer = time.NewTimer(MinIterTime)
				iterStartTime = time.Now()

				// Recreate the call of the nextPeer that the iterator just used
				callForPeer = true
			}

		pointerCheck:
			// check if we need to call for a new Peer
			if callForPeer {
				c.NextPeer()
				callForPeer = false
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
			logEntry.Tracef("new connection attempt has been received from peer %s", connAttempt.RemotePeer.String())
			// update the local info about the peer
			p, ok := c.PeerQueue.GetPeer(connAttempt.RemotePeer)
			if !ok {
				// we shoould never receive a connection attempt of a peer that is not in the list
				// thus, raise a error log with the attempt status
				if connAttempt.Status == models.NegativeAttempt {
					log.Errorf("we received a negative attempt of connection to %s - that was probably deprecated", connAttempt.RemotePeer.String())
				} else {
					log.Errorf("we received a possitive attempt of connection to %s - but was probably deprecated", connAttempt.RemotePeer.String())
				}
			} else {
				p.ConnEventHandler(connAttempt.Error)
				// Check if peer needs to be deprecated
				if p.Deprecable() {
					logEntry.Warnf("deprecating peer %s", connAttempt.RemotePeer.String())
					connAttempt.Deprecable = true
					// remove p from list of peers to ping (if it appears again in the discovery, it will be updated as undeprecated in the DB)
					c.PeerQueue.RemovePeer(connAttempt.RemotePeer)
				}
				c.DBClient.PersistToDB(connAttempt)
			}

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
			logEntry.Debugf("new identification from peer %s", identEvent.HostInfo.ID.String())
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
	c.m.RLock()
	defer c.m.RUnlock()
	return float64(c.lastIterTime.Microseconds()) / 1000000
}

func (c *PruningStrategy) AttemptedPeersSinceLastIter() int64 {
	c.m.RLock()
	defer c.m.RUnlock()
	attempted := int64(0)
	for _, v := range c.attemptedPeers {
		attempted += v
	}
	return attempted
}

func (c *PruningStrategy) ControlDistribution() map[string]int64 {
	return c.PeerQueue.DelayDistribution()
}

func (c *PruningStrategy) GetErrorAttemptDistribution() map[string]int64 {
	c.m.RLock()
	defer c.m.RUnlock()
	attemptDist := make(map[string]int64)
	for k, v := range c.attemptedPeers {
		attemptDist[string(k)] = v
	}
	return attemptDist
}

// PeerQueue is an auxiliar peer array and map list to keep the list of peers sorted
// by connection time, and still able to modify in a short time the values of each peer.
type PeerQueue struct {
	sync.RWMutex

	// DBs
	dbClient *psql.DBClient

	// control variables
	peerPtr  int
	peerList []*PrunedPeer
	peerMap  map[peer.ID]*PrunedPeer
}

// NewPeerQueue is the constructor of a NewPeerQueue
func NewPeerQueue(dbClient *psql.DBClient) *PeerQueue {
	return &PeerQueue{
		dbClient: dbClient,
		peerPtr:  0,
		peerList: make([]*PrunedPeer, 0),
		peerMap:  make(map[peer.ID]*PrunedPeer),
	}
}

func (c *PeerQueue) IsEmpty() bool {
	c.RLock()
	defer c.RUnlock()
	return len(c.peerList) == 0
}

func (c *PeerQueue) ValidNextPeer() bool {
	c.RLock()
	defer c.RUnlock()

	// if the pointer is in a correct range and the next peer is ready for connection
	if c.peerPtr >= c.Len() {
		return false
	} else {
		if !c.peerList[c.peerPtr].IsReadyForConnection() {
			return false
		}
	}
	return true
}

func (c *PeerQueue) GetNextPeer() *PrunedPeer {
	c.Lock()
	defer c.Unlock()

	nextPeer := c.peerList[c.peerPtr]
	c.peerPtr++
	// check if the pointer went beyond the limits
	// if c.peerPtr >= c.Len() {
	// 	// reset counter
	// 	c.peerPtr = 0
	// 	err := c.UpdatePeerListFromRemoteDB()
	// 	if err != nil {
	// 		log.Error(errors.Wrap(err, "exceeded number of peers in list"))
	// 	}
	// }
	return nextPeer
}

func (c *PeerQueue) ResetPeerPointer() {
	c.Lock()
	defer c.Unlock()
	c.peerPtr = 0
}

// IsPeerAlready checks whether a peer is already in the Queue.
func (c *PeerQueue) IsPeerAlready(id peer.ID) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.peerMap[id]
	return ok
}

// AddPeer Adds a peer to the peerqueue.
func (c *PeerQueue) AddPeer(pPeer *PrunedPeer) {
	c.Lock()
	defer c.Unlock()

	// append new item at the beginning of the array
	c.peerList = append([]*PrunedPeer{pPeer}, c.peerList...)
	c.peerMap[pPeer.iD] = pPeer
}

// RemovePeer()
func (c *PeerQueue) RemovePeer(id peer.ID) {
	c.Lock()
	defer c.Unlock()

	// check if we have the peer in our local peerqueue
	_, ok := c.peerMap[id]
	if !ok {
		log.Debugf("peer %s not in local peerstore", id.String())
		return
	}

	// proceed to delete the peer from our queue
	log.Debugf("total len of queue %d - removing peer %s", c.Len(), id.String())

	delete(c.peerMap, id)

	var idx int = -1
	for index, pInfo := range c.peerList {
		if pInfo.iD == id {
			idx = index
			break
		}
	}

	if idx > -1 {
		c.peerList = append(c.peerList[:idx], c.peerList[idx+1:]...)
	} else {
		log.Debugf("unable to find peer %s inside queued peers", id.String())
		return
	}

	log.Debugf("total len of queue %d post removing peer", c.Len())
}

// GetPeer retrieves the info of the peer requested from args.
func (c *PeerQueue) GetPeer(id peer.ID) (*PrunedPeer, bool) {
	c.Lock()
	defer c.Unlock()

	p, ok := c.peerMap[id]
	if !ok {
		return &PrunedPeer{}, ok
	}
	return p, ok
}

// DelayDistribution returns the distribution of the delays in a map.
func (c *PeerQueue) DelayDistribution() map[string]int64 {
	c.RLock()
	defer c.RUnlock()
	// iter through the peers in the queue map getting the distribution
	distribution := make(map[string]int64)
	for _, val := range c.peerMap {
		_, ok := distribution[string(val.delayObj.dtype)]
		if !ok {
			distribution[string(val.delayObj.dtype)] = int64(0)
		}
		distribution[string(val.delayObj.dtype)]++
	}
	return distribution
}

// SortPeerList sorts the PeerQueue array leaving at the beginning the peers
// with the shorter next peer connection.
func (c *PeerQueue) SortPeerList() {
	c.Lock()
	defer c.Unlock()
	sort.Sort(c)
}

// ---  SORTING METHODS FOR PeerQueue ----

// Swap is part of sort.Interface.
func (c *PeerQueue) Swap(i, j int) {
	c.peerList[i], c.peerList[j] = c.peerList[j], c.peerList[i]
}

// Less is part of sort.Interface. We use c.PeerList.NextConnection as the value to sort by.
func (c *PeerQueue) Less(i, j int) bool {
	return c.peerList[i].NextConnection().Before(c.peerList[j].NextConnection())
}

// Len is part of sort.Interface. We use the peer list to get the length of the array.
func (c *PeerQueue) Len() int {
	return len(c.peerList)
}

// UpdatePeerListFromPeerStore will refresh the peerqueue with the peerstore.
// Basically we add those peers that did not exist before in the peerqueue.
func (c *PeerQueue) UpdatePeerListFromRemoteDB() error {
	// Get the list of peers from the peerstore
	peerList, err := c.dbClient.GetNonDeprecatedPeers()
	if err != nil {
		return errors.Wrap(err, "fail to update pruning peer queue from DBClient")
	}
	// metrics
	totcnt := 0
	new := 0

	// Fill the PeerQueue.PeerList with the missing peers from the
	for _, connectablePeer := range peerList {
		totcnt++
		if !c.IsPeerAlready(connectablePeer.ID) {
			new++
			log.Tracef("peer %s not locally, storing it", connectablePeer.ID.String())
			// Whenever we find a new peer that we didn't have locally, add zero delay
			// even when we read all the peerstore from the DB Endpoint when restarting
			newPrunnedPeer := NewPrunedPeer(connectablePeer.ID, connectablePeer.Addrs, connectablePeer.Network, Minus1Delay)
			// add the new item to the list
			c.AddPeer(newPrunnedPeer)
		}
	}
	// Sort the list of peers based on the next connection
	c.SortPeerList()
	log.Debugf("Num of peers in PeerQueue: %d\n", c.Len())
	return nil
}

type PrunedPeer struct {
	iD                       peer.ID
	addr                     []ma.Multiaddr
	network                  utils.NetworkType
	delayObj                 DelayObject // define the delay to connect based on error
	baseConnectionTimestamp  time.Time   // define the first event. To calculate the next connection we sum this with delay.
	baseDeprecationTimestamp time.Time   // this + DeprecationTime defines when we are ready to deprecate
}

func NewPrunedPeer(id peer.ID, maddrs []ma.Multiaddr, network utils.NetworkType, delay Delay) *PrunedPeer {
	t := time.Now()
	pp := PrunedPeer{
		iD:                       id,
		addr:                     maddrs,
		network:                  network,
		delayObj:                 NewDelayObject(delay),
		baseConnectionTimestamp:  t,
		baseDeprecationTimestamp: t, // by default we set it now, so if no positive connection it will be deprecated in 24 hours since creation of this prunned peer
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
	if c.delayObj.dtype == Minus1Delay { // in case of Minus1, this is new peer and we want it to connect as soon as possible
		return time.Time{}
	}
	if c.delayObj.CalculateDelay() > MaxDelayTime {
		return c.baseConnectionTimestamp.Add(MaxDelayTime)
	}
	// nextConnection should be from first event + the applied delay
	return c.baseConnectionTimestamp.Add(c.delayObj.CalculateDelay())
}

// Deprecable evaluates if the peer is in time to be deprecated.
func (c *PrunedPeer) Deprecable() bool {
	// if the difference between now and the BaseDeprecationTimestampo is more than the DeprecationTime, true
	if time.Now().Sub(c.baseDeprecationTimestamp) >= DeprecationTime {
		return true
	}
	return false
}

// RecErrorHandler selects actuation method for each of the possible errors while actively dialing peers.
func (c *PrunedPeer) ConnEventHandler(recErr string) {
	c.UpdateDelay(ErrorToDelayType(recErr))
}

// NewEvent will reevaluate the delay in case of a new Positive or NegativeDelay happens
func (c *PrunedPeer) UpdateDelay(newDelayType Delay) {
	// if the delaytype is different, always refresh the object
	c.baseConnectionTimestamp = time.Now()

	if c.delayObj.dtype != newDelayType {
		c.delayObj = NewDelayObject(newDelayType)
	}

	// if there is a positive delay (success identify), then we update the deprecation time
	// therefore, we start counting from now to deprecate
	if c.delayObj.dtype == PositiveDelay {
		c.baseDeprecationTimestamp = time.Now()
	}

	c.delayObj.IncreaseDegree()
}
