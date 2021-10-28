package peering

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/utils"
	"github.com/migalabs/armiarma/src/hosts"

	log "github.com/sirupsen/logrus"
)

var (
	PruningStrategyName = "PRUNING"
	DefaultDelay        = 24 * time.Hour   // hours of dealy after each negative attempt with delay
	MinIterTime         = 10 * time.Second // Minimum time that has to pass before iterating again
	ConnEventBuffSize   = 10
)

type PruningOpts struct {
	AggregatedDelay time.Duration
	LogOpts         base.LogOpts
}

// Pruning Strategy is a Peering Strategy that applies penalties to peers that haven't show activity when attempting to connect them.
// Combined with the Deprecated flag in the db.Peer struct, it produces more accourated metrics when exporting, pruning peers that are no longer active.
type PruningStrategy struct {
	*base.Base
	strategyType string
	PeerStore    *db.PeerStore
	// Delay unit time that gets applied to those slashed peers when reporting inactivity errors when activly connecting
	AggregatedDelay time.Duration
	// Peer Stream and Return ConnectionStatus channels (communication between modules)
	// both empty by default (need for initialization)

	peerStreamChan chan db.Peer
	nextPeerChan   chan struct{}
	connAttemptNot chan ConnectionAttemptStatus
	connNot        chan hosts.ConnectionStatus
	disconnNot     chan hosts.DisconnectionStatus
	/*
		// TODO: Choose the necessary parameters for the pruning
		FilterDigest beacon.ForkDigest `ask:"--filter-digest" help:"Only connect when the peer is known to have the given fork digest in ENR. Or connect to any if not specified."`
		FilterPort   int               `ask:"--filter-port" help:"Only connect to peers that has the given port advertised on the ENR."`
		Filtering    bool              `changed:"filter-digest"`
	*/
}

func NewPruningStrategy(ctx context.Context, peerstore *db.PeerStore, opts PruningOpts) (PruningStrategy, error) {
	// TODO: cancel is still not implemented in the BaseCreation
	pruningCtx, _ := context.WithCancel(ctx)
	opts.LogOpts.ModName = PruningStrategyName
	b, err := base.NewBase(
		base.WithContext(pruningCtx),
		base.WithLogger(opts.LogOpts),
	)
	if err != nil {
		return PruningStrategy{}, err
	}
	// Generate the ConnStatus notification channel
	// TODO: consider making the ConnStatus channel larger
	pr := PruningStrategy{
		Base:           b,
		strategyType:   PruningStrategyName,
		PeerStore:      peerstore,
		peerStreamChan: make(chan db.Peer, ConnEventBuffSize),
		nextPeerChan:   make(chan struct{}, ConnEventBuffSize),
		connAttemptNot: make(chan ConnectionAttemptStatus, ConnEventBuffSize),
		connNot:        make(chan hosts.ConnectionStatus, ConnEventBuffSize),
		disconnNot:     make(chan hosts.DisconnectionStatus, ConnEventBuffSize),
	}
	return pr, nil
}

func (c PruningStrategy) Type() string {
	return c.strategyType
}

// NextPeer
// * Is a function that returns an iterator of Peers received from the PeerStore
// @return function that gives another peer to connect from the PeerStore (Filtered by the given strategy)
func (c *PruningStrategy) Run() chan db.Peer {
	// start go routine that will notify of the full peerstore iteration and notifies it to the main strategy loop
	go c.peerstoreIterator()
	return c.peerStreamChan
}

// peerstoreIterator
// * Private function that is in charge of iterating through the peerstore,
// * receive connections/disconnectios, and fetch info comming from the peering service into the db
// * Main interaction of the Peering Service with the DB
// @param
// @return
// TODO: 	Set this as a different module inside strategy
// 			Implement some kind of sorting over the peer list, to reduce iteration time
func (c *PruningStrategy) peerstoreIterator() {
	// get Ctx of the pruning module
	modCtx := c.Ctx()
	// get the peer list from the peerstore
	peerList := c.PeerStore.GetPeerList()
	peerCounter := 0
	peerListLen := len(peerList)
	validIterTimer := time.NewTimer(MinIterTime)
	iterStartTime := time.Now()
	for {
		select {
		// Receive the notification of sending the next peer
		case <-c.nextPeerChan:
			if peerListLen > 0 {
				c.Log.Debug("prepare next peer for pushing it into peer stream")
				// read info about next peer
				// TODO: So far just iterating the entire peerstore
				peerID := peerList[peerCounter]
				pinfo, err := c.PeerStore.GetPeerData(peerID.String())
				if err != nil {
					log.Warn(err)
					pinfo = db.NewPeer(peerID.String())
				}
				// check if peer has been already deprecated for being many hours without connected
				// TODO: resume this in a function
				wtime := pinfo.DaysToWait()
				if wtime != 0 {
					lconn, err := pinfo.LastAttempt()
					if err != nil {
						log.Warn("the peer should have a last connection attempt but list is empty")
					}
					lconnSecs := lconn.Add(time.Duration(wtime) * DefaultDelay).Unix()
					tnow := time.Now().Unix()
					// Compare time now with last connection plus waiting list
					if (tnow - lconnSecs) <= 0 { // TODO: should be replaced by a filtered list where the next peer is ready to be connected
						// If result is lower than 0, still have time to wait
						// continue to next peer
						// Recreate the call of the nextPeer that the iterator just used
						c.NextPeer()
						continue
					}
				}

				// compose all the detailed info for the given peer
				// Generating New peer to fetch info
				peer := db.NewPeer(pinfo.PeerId)
				peerEnr := pinfo.GetBlockchainNode()
				if peerEnr != nil {
					peer.NodeId = peerEnr.ID().String()
					// TODO:
					peer.Ip = peerEnr.IP().String()
				}

				peer.PeerId = peerID.String()
				k, _ := peerID.ExtractPublicKey()
				pubk, _ := k.Raw()
				peer.Pubkey = hex.EncodeToString(pubk)
				peer.MAddrs = pinfo.MAddrs
				// Update metadata of peer
				c.PeerStore.StoreOrUpdatePeer(peer)
				// Send next peer to the peering service
				c.Log.Debugf("pushing next peer %s into peer stream", pinfo.PeerId)
				c.peerStreamChan <- pinfo

				/* TODO: Deprecated for now
				// wait for the response of the ConnStatus (CAREFUL: I hope this doesn't block the strategy)
				<-c.ConnStatusNot
				*/
				// increment peerCounter to see if we finished iterating the peerstore
				peerCounter++
			} else {
				c.Log.Warn("empty peerstore")
				// Recreate the call of the nextPeer that the iterator just used
				c.NextPeer()
				/*
					// Just in case we dont enter the next if sentence
					c.Log.Debug("waiting for min iter time")
					<-validIterTimer.C
					// reset values
					peerList = c.PeerStore.GetPeerList()
					peerListLen = len(peerList)
					c.Log.Debug(" min iter time done, got new peer list with %d", len(peerList))
					validIterTimer = time.NewTimer(MinIterTime)
					peerCounter = 0
				*/
			}
			if peerCounter >= peerListLen {
				// time to update the PeerList
				iterTime := time.Since(iterStartTime)
				c.Log.Debug("peerstore iteration done in ", iterTime)
				c.PeerStore.NewPeerstoreIteration(iterTime)
				// check if the minIterTime has been
				<-validIterTimer.C
				// reset values
				peerList = c.PeerStore.GetPeerList()
				peerListLen = len(peerList)
				c.Log.Debugf("got new peer list with %d", len(peerList))
				validIterTimer = time.NewTimer(MinIterTime)
				peerCounter = 0
			}

		// Receive the status of the peer that got connected to the crawler
		case connAttemtpStatus := <-c.connAttemptNot:
			c.Log.Debugf("new connection attempt has been received from peer %s", connAttemtpStatus.Peer.PeerId)
			if connAttemtpStatus.Successful {
				c.Log.Debugf("adding success connection to peer %s", connAttemtpStatus.Peer.PeerId)
				c.PeerStore.StoreOrUpdatePeer(connAttemtpStatus.Peer)
				c.PeerStore.AddNewPosConnectionAttempt(connAttemtpStatus.Peer.PeerId)
			} else {
				c.Log.Debugf("adding negative connection to peer %s", connAttemtpStatus.Peer.PeerId)
				c.RecErrorHandler(connAttemtpStatus.Peer.PeerId, connAttemtpStatus.RecError.Error())
			}

		// Receive the notification of a that got disconnected from the crawler
		case connStat := <-c.connNot:
			c.Log.Debugf("new connection has been received from peer %s", connStat.Peer.PeerId)
			c.PeerStore.StoreOrUpdatePeer(connStat.Peer)

		// Receive the notification of a that got disconnected from the crawler
		case disconnStat := <-c.disconnNot:
			c.Log.Debugf("new disconnection has been received from peer %s", disconnStat.Peer.PeerId)
			c.PeerStore.StoreOrUpdatePeer(disconnStat.Peer)

		// detect if the context has been shut down to end the go routine
		case <-modCtx.Done():
			c.Log.Debug("closing peerstore iterator")
		}
	}
}

// ClosePeerStream
// * Closes in a controled secuence the module related go routines and channels
// * Ending with the Base.Ctx cancelation
func (c *PruningStrategy) Close() {
	c.Log.Infof("closing pruning strategy")
	// close the involved channels
	close(c.peerStreamChan)
	close(c.nextPeerChan)
	close(c.connNot)
	close(c.disconnNot)
	// shutdown the base context of the pruning
	c.Cancel()
}

// NextPeer
// * Notifies the peerstore iterator that a new peer has been requested
// * After it, the peerstore iteratow will put the new peer in the PeerStreamChan
func (c *PruningStrategy) NextPeer() {
	c.Log.Debug("next peer has been requested")
	c.nextPeerChan <- struct{}{}
}

// NewConnectionAttemptStatus
// * Notifies the peerstore iterator that a new ConnStatus has been received
// * After it, the peerstore iteratow will aggregate the extra info
func (c *PruningStrategy) NewConnectionAttempt(connAttStat ConnectionAttemptStatus) {
	c.Log.Debug("next connection has been received")
	c.connAttemptNot <- connAttStat
}

// NewConnectionStatus
// * Notifies the peerstore iterator that a new ConnStatus has been received
// * After it, the peerstore iteratow will aggregate the extra info
func (c *PruningStrategy) NewConnection(connStat hosts.ConnectionStatus) {
	c.Log.Debug("next connection has been received")
	c.connNot <- connStat
}

// NewConnectionStatus
// * Notifies the peerstore iterator that a new ConnStatus has been received
// * After it, the peerstore iteratow will aggregate the extra info
func (c *PruningStrategy) NewDisconnection(disconnStat hosts.DisconnectionStatus) {
	c.Log.Debug("next connection has been received")
	c.disconnNot <- disconnStat
}

// peeringWorker
// *
// *
// @params
// @return
// TODO: Still not sure if we need workers for iterating the peerstore
func peeringWorker(ctx context.Context, ps *db.PeerStore, peerChan chan string) {

}

// RecErrorHandler
// * function that selects actuation method for each of the possible errors while actively dialing peers
// @params peerID in string format, recorded error in string format
func (c *PruningStrategy) RecErrorHandler(pe string, rec_err string) {
	var fn func(p *db.Peer)
	switch utils.FilterError(rec_err) {
	case "Connection reset by peer":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "i/o timeout":
		fn = func(p *db.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	case "dial to self attempted":
		// we tried to peer ourselfs! deprecate the peer
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "dial backoff":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "connection refused":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "no route to host":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "unreachable network":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "peer id mismatch, peer dissmissed":
		// deprecate old peer and generate a new one
		// trim new peerID from error message
		np := strings.Split(rec_err, "key matches ")[1]
		np = strings.Replace(np, ")", "", -1)
		//newPeerID := peer.ID(np)
		//f.WriteString(fmt.Sprintf("%s shifted to %s\n", pe.String(), newPeerID))
		// Generate a new Peer with the addrs of the previous one and the ID suggested at the
		log.Infof("deprecating peer %s, but adding possible new peer %s", pe, np)
		/*
			_, err := newPeerID.ExtractPublicKey()
			if err != nil {
				fmt.Println("error obtainign pubkey from peerid", err)
			} else {
				fmt.Println("pubkey success, obtained")
			}
			TODO: -Fix empty pubkey originated from adding the new PeerID to the Peerstore
					-Apparently the len of the new peer doesn't have the same one as the previous one
			// Generate new Addrs for the possible new discovered peer
			addrs := c.Store.Addrs(pe)
			enr := c.Store.LatestENR(pe)
			fmt.Println("len old", len(pe.String()), "len new", len(newPeerID.String()))
			fmt.Println(rec_err)
			// Info about the peer should be added to the metrics
			// *** Carefull - problems with reading the pubkey ***
			//newP := db.NewPeer(newPeerID.String())
			//c.PeerStore.Store(newPeerID.String(), newP)
			_, _ = c.Store.UpdateENRMaybe(newPeerID, enr)
			c.Store.AddAddrs(newPeerID, addrs, time.Duration(48)*time.Hour)
		*/
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	default:
		fn = func(p *db.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	}
	c.PeerStore.AddNewNegConnectionAttempt(pe, rec_err, fn)
}
