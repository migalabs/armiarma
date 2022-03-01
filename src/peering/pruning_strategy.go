package peering

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/hosts"
	db_utils "github.com/migalabs/armiarma/src/utils"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
)

var (
	PeerStratModuleName = "PRUNING"
	slog                = logrus.WithField(
		"module", PeerStratModuleName,
	)
	// Default Delays
	DeprecationTime       = 1024 * time.Minute // mMinutes after first negative connection that has to pass to deprecate a peer.
	DefaultNegDelay       = 12 * time.Hour     // Default delay that will be applied for those deprecated peers.
	DefaultPossitiveDelay = 6 * time.Hour      // Default delay after each positive severe negative attempts.
	StartExpD             = 2 * time.Minute    // Starting delay that will serve for the Exponential Delay.
	// Control variables
	MinIterTime = 15 * time.Second // Minimum time that has to pass before iterating again.

)

// Pruning Strategy is a Peering Strategy that applies penalties to peers that haven't shown activity when attempting to connect them.
// Combined with the Deprecated flag in the models.Peer struct, it produces more accurate metrics when exporting pruning peers that are no longer active.
type PruningStrategy struct {
	ctx context.Context

	network      string
	strategyType string
	PeerStore    *db.PeerStore

	// Peer Stream and Return ConnectionStatus channels (communication between modules)
	// both empty by default (need for initialization)
	peerStreamChan chan models.Peer
	nextPeerChan   chan struct{}
	connAttemptNot chan ConnectionAttemptStatus
	connEventNot   chan hosts.ConnectionEvent
	identEventNot  chan hosts.IdentificationEvent

	// List of peers sorted by the amount of time thatwe have to wait
	PeerQueue PeerQueue

	// Prometheus Control Variables
	lastIterTime             time.Duration
	iterForcingNextConnTime  time.Time
	attemptedPeers           int64
	queueErroDistribution    map[string]*int64
	PeerQueueIterations      int
	ErrorAttemptDistribution map[string]*int64
}

// NewPruningStrategy:
// Pruning strategy constructor, that will offer a models.Peer stream for the
// peering service. The provided models.Peer stream are ready to connect.
// @param ctx: parent context.
// @param peerstore: db.PeerStore.
// @param opts: base and logging option.
// @return peering strategy interface with the prunning service.
// @return error.
func NewPruningStrategy(ctx context.Context, network string, peerstore *db.PeerStore) (PruningStrategy, error) {
	// Generate the ConnStatus notification channel
	// TODO: consider making the ConnStatus channel larger
	pr := PruningStrategy{
		ctx:            ctx,
		network:        network,
		strategyType:   PeerStratModuleName,
		PeerStore:      peerstore,
		PeerQueue:      NewPeerQueue(),
		peerStreamChan: make(chan models.Peer, DefaultWorkers),
		nextPeerChan:   make(chan struct{}, DefaultWorkers),
		connAttemptNot: make(chan ConnectionAttemptStatus),
		connEventNot:   make(chan hosts.ConnectionEvent),
		identEventNot:  make(chan hosts.IdentificationEvent),
		// Metrics Variables
		queueErroDistribution:    make(map[string]*int64),
		ErrorAttemptDistribution: make(map[string]*int64),
	}
	return pr, nil
}

// Type:
// Returns the strategy type that has been set.
// @return string with the name of the pruning strategy.
func (c PruningStrategy) Type() string {
	return c.strategyType
}

// Run:
// Initializes the models.Peer stream on the returning models.Peer chan
// stores locally an auxiliary map wuth an array that will keep
// track of the next connection time.
// @return models.Peer channel with the next peer to connect.
func (c *PruningStrategy) Run() chan models.Peer {
	// start go routine that will notify of the full peerstore iteration and notifies it to the main strategy loop
	go c.peerstoreIteratorRoutine()
	go c.eventRecorderRoutine()

	return c.peerStreamChan
}

// peerstoreIterator:
// Private function that is in charge of iterating through the peerstore,
// receive connections/disconnections, and fetch info comming from the peering service into the db.
// Main interaction of the Peering Service with the DB.
func (c *PruningStrategy) peerstoreIteratorRoutine() {
	slog.Debug("starting the peerstore iterator routine")
	c.PeerQueueIterations = 0

	// get the peer list from the peerstore
	err := c.PeerQueue.UpdatePeerListFromPeerStore(c.PeerStore)
	if err != nil {
		slog.Error(err)
	}

	posDelay := int64(0)
	negWHDelay := int64(0)
	negWHTDelay := int64(0)
	min1Delay := int64(0)
	timeoutType := int64(0)
	zerDelay := int64(0)

	errorAttemptCategories := map[string]*int64{
		PositiveDelayType:           &posDelay,
		NegativeWithHopeDelayType:   &negWHDelay,
		NegativeWithNoHopeDelayType: &negWHTDelay,
		Minus1DelayType:             &min1Delay,
		ZeroDelayType:               &zerDelay,
		TimeoutDelayType:            &timeoutType,
	}
	c.ErrorAttemptDistribution = errorAttemptCategories
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
				slog.Debug("prepare next peer for pushing it into peer stream")
				// read info about next peer
				nextPeer := c.PeerQueue.PeerList[peerCounter]
				// check if the node is ready for connection
				// or we are in the first iteration, then we always try all of them
				if nextPeer.IsReadyForConnection() || c.PeerQueueIterations == 0 {
					pinfo, err := c.PeerStore.GetPeerData(nextPeer.PeerID)
					if err != nil {
						slog.Warn(err)
						pinfo = models.NewPeer(nextPeer.PeerID)
					}
					// compose all the detailed info for the given peer
					// Generating New peer to fetch info
					npeer := models.NewPeer(nextPeer.PeerID)

					// get IP
					pID, _ := peer.Decode(nextPeer.PeerID)
					if err != nil {
						slog.Errorf("error decoding PeerID string into peer.ID %s", err.Error())
					}
					npeer.PeerId = pID.String()

					// Eth2 network
					if c.network == "eth2" {
						peerEnr, ok := pinfo.GetAtt("enr")
						enr := enode.MustParse(peerEnr.(string))
						if ok && peerEnr != nil {
							npeer.SetAtt("nodeid", enr.ID().String())
							// TODO:
							npeer.Ip = enr.IP().String()
						}
						k, _ := pID.ExtractPublicKey()
						pubk, _ := k.Raw()
						npeer.SetAtt("pubkey", hex.EncodeToString(pubk))
					}

					npeer.MAddrs = pinfo.MAddrs
					// Update metadata of peer
					c.PeerStore.StoreOrUpdatePeer(npeer)

					// Send next peer to the peering service
					slog.Debugf("pushing next peer %s into peer stream", pinfo.PeerId)
					atomic.AddInt64(errorAttemptCategories[nextPeer.DelayObj.GetType()], 1)
					c.peerStreamChan <- pinfo

					// increment peerCounter to see if we finished iterating the peerstore
					peerCounter++

				} else {
					slog.Debug("next peers has to wait to be connected")
					c.iterForcingNextConnTime = nextPeer.NextConnection()

					c.NextPeer()
					nextIterFlag = true
				}
			} else {
				slog.Warn("empty peerstore")
				// Recreate the call of the nextPeer that the iterator just used
				c.NextPeer()

			}
			if nextIterFlag || peerCounter >= peerListLen {
				// time to update the PeerList
				c.lastIterTime = time.Since(iterStartTime)
				atomic.StoreInt64(&c.attemptedPeers, int64(peerCounter))
				slog.Info("peerstore iteration of ", peerCounter, " peers, done in ", c.lastIterTime)
				slog.Info("missing ", c.PeerQueue.Len()-peerCounter, " peers waiting for next try")

				// check if the minIterTime has been
				<-validIterTimer.C

				// reset values

				c.ErrorAttemptDistribution = errorAttemptCategories
				errorAttemptCategories = ResetMapValues(errorAttemptCategories)
				// get the peer list from the peerstore
				err := c.PeerQueue.UpdatePeerListFromPeerStore(c.PeerStore)
				c.PeerQueueIterations++ // another iteration
				if err != nil {
					slog.Error(err)
				}
				peerListLen = c.PeerQueue.Len()
				slog.Infof("got new peer list with %d", peerListLen)
				validIterTimer = time.NewTimer(MinIterTime)
				iterStartTime = time.Now()
				peerCounter = 0
				nextIterFlag = false
			}

		// detect if the context has been shut down to end the go routine
		case <-c.ctx.Done():
			slog.Debug("closing peerstore iterator")
			close(c.peerStreamChan)
			close(c.nextPeerChan)
			close(c.connEventNot)
			return
		}
	}
}

// peerstoreIterator:
// Private function that is in charge of iterating through the peerstore,
// receive connections/disconnections, and fetch info comming from the peering service into the db.
// Main interaction of the Peering Service with the DB.
func (c *PruningStrategy) eventRecorderRoutine() {
	slog.Debugf("Running the Event RecorderRoutine")
	for {
		select {
		// Receive the status of the peer that got connected to the crawler
		case connAttemtpStatus := <-c.connAttemptNot:
			slog.Debugf("new connection attempt has been received from peer %s", connAttemtpStatus.Peer.PeerId)

			if connAttemtpStatus.Successful { // in case of success we do not register in any pruned peer
				// pruned per only receives positive events in the identify case down below
				slog.Debugf("adding success connection to peer %s", connAttemtpStatus.Peer.PeerId)
				connAttemtpStatus.Peer.AddPositiveConnAttempt()

			} else {
				// negative event, register in pruned peer
				slog.Debugf("adding negative connection to peer %s", connAttemtpStatus.Peer.PeerId)
				// Update Pruning Metadata
				var p *PrunedPeer
				p, ok := c.PeerQueue.GetPeer(connAttemtpStatus.Peer.PeerId)
				if !ok {
					slog.Warnf("Could not find peer in peerqueue: %s", connAttemtpStatus.Peer.PeerId)
				}
				errString := p.ConnEventHandler(connAttemtpStatus.RecError.Error())
				connAttemtpStatus.Peer.AddNegConnAtt(p.Deprecable(), errString)

			}
			c.PeerStore.StoreOrUpdatePeer(connAttemtpStatus.Peer)

		// Receive a notification of a connection event
		case connEvent := <-c.connEventNot:
			switch connEvent.ConnType {
			case int8(1):
				slog.Debugf("new conection from %s", connEvent.Peer.PeerId)
			case int8(2):
				slog.Debugf("new disconnection has been received from peer %s", connEvent.Peer.PeerId)
			default:
				slog.Warnf("unrecognized event from peer %s", connEvent.Peer.PeerId)
				// since its an unexpected event, dont fetch the incoming peer and reset de select case
				continue
			}
			c.PeerStore.StoreOrUpdatePeer(connEvent.Peer)

		case identEvent := <-c.identEventNot:
			// TODO: We received a new identification attempt
			// extract whether it was success or a failure
			// and track it in the PeerQueue and in the peerstore
			slog.Debugf("new identification %s from peer %s", strconv.FormatBool(identEvent.Peer.IsConnected), identEvent.Peer.PeerId)

			// by default we think the identify was not successful, therefore negativewithhope and error
			delayType := NegativeWithHopeDelayType
			errorType := "Error requesting metadata"
			if identEvent.Peer.MetadataSucceed {
				// positive identify
				delayType = PositiveDelayType
				errorType = "None"
			}
			var p *PrunedPeer
			p, ok := c.PeerQueue.GetPeer(identEvent.Peer.PeerId)
			if !ok {
				p = NewPrunedPeer(identEvent.Peer.PeerId, delayType)
				c.PeerQueue.AddPeer(p)
			}

			p.ConnEventHandler(errorType)

			// peerstore save data
			c.PeerStore.StoreOrUpdatePeer(identEvent.Peer)

		// detect if the context has been shut down to end the go routine
		case <-c.ctx.Done():
			slog.Debug("closing event recorder routine")
			return
		}
	}
}

// NextPeer:
// Notifies the peerstore iterator that a new peer has been requested.
// After it, the peerstore iterator will put the new peer in the PeerStreamChan.
func (c *PruningStrategy) NextPeer() {
	slog.Debug("next peer has been requested")
	c.nextPeerChan <- struct{}{}
}

// NewConnectionAttempt:
// Notifies the peerstore iterator that a new ConnStatus has been received.
// After it, the peerstore iterator will aggregate the extra info.
// @param connAttStat: the object containing the data from the attempt
func (c *PruningStrategy) NewConnectionAttempt(connAttStat ConnectionAttemptStatus) {
	slog.Debug("new connection attempt has been received from peer", connAttStat.Peer.PeerId)
	c.connAttemptNot <- connAttStat
}

// NewConnectionEvent:
// Notifies the peerstore iterator that a new Connection has been received.
// It puts the connection metadata in the connNot channel to let the select
// loop all the metadata of the received connection.
func (c *PruningStrategy) NewConnectionEvent(connEvent hosts.ConnectionEvent) {
	slog.Debug("next connection event has been received from peer", connEvent.Peer.PeerId)
	c.connEventNot <- connEvent
}

// NewIdentificationEvent:
// This method will insert a new identification item in the identificationeventnorifier channel.
// @param newIdent: the object containing data about the event.
func (c *PruningStrategy) NewIdentificationEvent(newIdent hosts.IdentificationEvent) {
	slog.Debugf("new identification %s has been received from peer %s", strconv.FormatBool(newIdent.Peer.IsConnected), newIdent.Peer.PeerId)
	c.identEventNot <- newIdent
}

// --------------------------------------------------
// Metrics Exporting Functions for Peering Prometheus
// --------------------------------------------------

// LastIterTime
// @return the lastiteration time of the peerqueue
func (c *PruningStrategy) LastIterTime() float64 {
	return float64(c.lastIterTime.Microseconds()) / 1000000
}

func (c *PruningStrategy) IterForcingNextConnTime() string {
	return c.iterForcingNextConnTime.String()
}

func (c *PruningStrategy) AttemptedPeersSinceLastIter() int64 {
	return atomic.LoadInt64(&c.attemptedPeers)
}

func (c *PruningStrategy) ControlDistribution() map[string]*int64 {
	return c.PeerQueue.DelayDistribution()
}

func (c *PruningStrategy) GetErrorAttemptDistribution() map[string]*int64 {
	return c.ErrorAttemptDistribution
}

// PeerQueue:
// Auxiliar peer array and map list to keep the list of peers sorted
// by connection time, and still able to modify in a short time
// the values of each peer.
type PeerQueue struct {
	PeerList []*PrunedPeer
	PeerMap  map[string]*PrunedPeer
	// Metrics
	queueErroDistribution map[string]*int64
}

// DelayDistribution:
// @return the distribution of the delays in a map.
func (c *PeerQueue) DelayDistribution() map[string]*int64 {
	return c.queueErroDistribution
}

// NewPeerQueue:
// Constructor of a NewPeerQueue.
// @return new PeerQueue.
func NewPeerQueue() PeerQueue {
	pq := PeerQueue{
		PeerList:              make([]*PrunedPeer, 0),
		PeerMap:               make(map[string]*PrunedPeer),
		queueErroDistribution: make(map[string]*int64),
	}
	return pq
}

// IsPeerAlready:
// Check whether a peer is already in the Queue.
// @params peerID: string of the peerID that we want to find.
// @return true is peer is already, false if not.
func (c *PeerQueue) IsPeerAlready(peerID string) bool {
	_, ok := c.PeerMap[peerID]
	return ok
}

// AddPeer
// Add a peer to the peerqueue.
// @params pPeer: the pruned peer to add
func (c *PeerQueue) AddPeer(pPeer *PrunedPeer) {
	// append new item at the beginning of the array
	c.PeerList = append([]*PrunedPeer{pPeer}, c.PeerList...)
	c.PeerMap[pPeer.PeerID] = pPeer
}

// GetPeer:
// Retrieves the info of the peer requested from args.
// @params peerID: string of the peerID that we want to find.
// @return pointer to pruned peer.
// @return bool, true if exists, false if doesn't.
func (c *PeerQueue) GetPeer(peerID string) (*PrunedPeer, bool) {
	p, ok := c.PeerMap[peerID]
	return p, ok
}

// SortPeerList:
// Sort the PeerQueue array leaving at the beginning the peers
// with the shorter next peer connection.
func (c *PeerQueue) SortPeerList() {
	sort.Sort(c)
}

// SORTING METHODS FOR PeerQueue

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

// UpdatePeerListFromPeerStore
// This method will refresh the peerqueue with the peerstore.
// Basically we add those peers that did not exist before in the peerqueue.
// @param peerstore: db where to read from.
func (c *PeerQueue) UpdatePeerListFromPeerStore(peerstore *db.PeerStore) error {
	// Get the list of peers from the peerstore
	peerList := peerstore.GetPeerList()
	totcnt := 0
	new := 0

	posType := int64(0)
	negWHType := int64(0)
	negWHTType := int64(0)
	min1Type := int64(0)
	timeoutType := int64(0)
	zerType := int64(0)

	c.queueErroDistribution = map[string]*int64{
		PositiveDelayType:           &posType,
		NegativeWithHopeDelayType:   &negWHType,
		NegativeWithNoHopeDelayType: &negWHTType,
		Minus1DelayType:             &min1Type,
		ZeroDelayType:               &zerType,
		TimeoutDelayType:            &timeoutType,
	}
	// Fill the PeerQueue.PeerList with the missing peers from the
	for _, peerID := range peerList {
		totcnt++
		if !c.IsPeerAlready(peerID.String()) {
			new++
			// Peer was not in the list of peers
			pInfo, err := peerstore.GetPeerData(peerID.String())
			if err != nil {
				return fmt.Errorf("unable import peer to PeerQueue. %s", err.Error())
			}
			// check the last connAttempt of the peer, in case we are restoring the peerqueue
			// from the peerstore in case of restart

			errorList := pInfo.Error
			delayDegree := 0                       // default
			delayType := NegativeWithHopeDelayType // default

			if len(pInfo.ConnectionTimes) > 0 || pInfo.Attempted { // this peer has had activity
				// get last connection time
				if len(errorList) > 0 {
					// there are errors in the peer (either none or something)
					lastError := errorList[len(errorList)-1]

					if lastError == "None" {
						if !pInfo.MetadataSucceed {
							// it could happen that error is None and MetadataSuccess = false
							// connection successful, but no metadata
							lastError = "Error requesting metadata"
						}

					} else {
						for i, _ := range errorList { /// experimental
							// recreate the nuber of consecutive errors backwards
							if errorList[len(errorList)-1-i] == lastError {
								delayDegree++
							} else {
								break
							}
						}
					}

					// we now have the last error and the degree repeated
					delayType = ErrorToDelayType(lastError)

				} else {
					// there are no errors, but connections
					// all inbound
					if pInfo.MetadataSucceed {
						delayType = PositiveDelayType
					} else {
						delayType = ErrorToDelayType("Error requesting metadata")
					}
				}

			} else {
				// this peer is new
				delayType = Minus1DelayType
			}

			newPrunnedPeer := NewPrunedPeer(pInfo.PeerId, delayType)
			newPrunnedPeer.DelayObj.SetDegree(delayDegree)

			if pInfo.Deprecated {
				// set basedeprecationtime to the past so this keeps deprecated
				newPrunnedPeer.BaseDeprecationTimestamp = time.Now().Add(-DeprecationTime)
			}

			// add the new item to the list
			c.AddPeer(newPrunnedPeer)
			atomic.AddInt64(c.queueErroDistribution[delayType], 1)
		} else {
			prunnedPeer, _ := c.GetPeer(peerID.String())
			atomic.AddInt64(c.queueErroDistribution[prunnedPeer.DelayObj.GetType()], 1)
		}
	}
	// Sort the list of peers based on the next connection
	c.SortPeerList()
	slog.Infof("len PeerQueue: %d\n", c.Len())
	return nil
}

// TODO: think about includint a sync.RWMutex in case we upgrade to workers
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

// IsReadyForConnection:
// This method evaluates if the given peer is ready to be connected.
// @return True of False if we are in time to connect or not.
func (c *PrunedPeer) IsReadyForConnection() bool {
	now := time.Now()
	// if we are not before the time, then we are either equal or after the connection time
	return !now.Before(c.NextConnection())
}

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

// Deprecable:
// This method evaluates if the peer is in time to be deprecated.
// @return true (in time to be deprecated) / false (not ready to be deprecated).
func (c *PrunedPeer) Deprecable() bool {
	// if the difference between now and the BaseDeprecationTimestampo is more than the DeprecationTime, true
	if time.Now().Sub(c.BaseDeprecationTimestamp) >= DeprecationTime {
		return true
	}

	return false
}

// RecErrorHandler:
// Function that selects actuation method for each of the possible errors while actively dialing peers.
// @params peerID in string format, recorded error in string format.
func (c *PrunedPeer) ConnEventHandler(recErr string) string {

	c.UpdateDelay(ErrorToDelayType(recErr))

	return db_utils.FilterError(recErr)
}

// NewEvent:
// This method will reevaluate the delay in case of a new
// Positive or NegativeDelay happenned
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

// ErrorToDelayType:
// Transforms an error into a DelayType.
// @param errString: the string to analyze.
// @return the categroy type in string format.
func ErrorToDelayType(errString string) string {

	prettyErr := db_utils.FilterError(errString)
	switch prettyErr {
	case "none":
		return PositiveDelayType
	case "connection reset by peer", "connection refused", "context deadline exceeded", "dial backoff", "metadata error":
		return NegativeWithHopeDelayType
	case "no route to host", "unreachable network", "peer id mismatch", "dial to self attempted":
		return NegativeWithNoHopeDelayType
	case "i/o timeout":
		return TimeoutDelayType
	default:
		slog.Warnf("Default Delay applied, error: %s-\n", prettyErr)
		return NegativeWithHopeDelayType
	}
}

// ResetMapValues
// Iterates over a string int map and resets all values to 0.
// @return the reset map
func ResetMapValues(inputMap map[string]*int64) map[string]*int64 {
	for k := range inputMap {
		atomic.StoreInt64(inputMap[k], 0)
	}
	return inputMap
}
