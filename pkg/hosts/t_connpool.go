package hosts

import (
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/db/postgresql"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

// TODO: - Still missing proper error handling from the persister
// 		 - Leaving the connection attempts for another table and model

// Aggregation of orphan connections that weren't persisted to the database
type ConnectionPool struct {
	// peer identifier of the host making the connections
	hostID peer.ID

	// Cache of open connections
	connPool map[string]*models.ConnEvent

	// Connection to the database
	dbClient *postgresql.DBClient
}

func NewConnectionPool(hostID peer.ID, dbCli *postgresql.DBClient) *ConnectionPool {
	return &ConnectionPool{
		hostID:   hostID,
		connPool: make(map[string]*models.ConnEvent, 0),
		dbClient: dbCli,
	}
}

func (p *ConnectionPool) AddNewEvent(conEv *models.ConnEvent) error {
	// check if the ConnEvent contains the min basic info
	if conEv.PeerID == "" {
		return errors.New("trying to add a new ConnEvent with an empty PeerID")
	}

	// Check if we already have an event for that peer
	_, ok := p.CheckIfOpenEventForPeer(conEv.PeerID)
	if !ok {
		// add the received new connEv to the pool
		p.connPool[conEv.PeerID.String()] = conEv
	}
	return nil
}

// Check and return if there is an already existing ConnEvent for a given peer
func (p *ConnectionPool) CheckIfOpenEventForPeer(pID peer.ID) (*models.ConnEvent, bool) {
	con, ok := p.connPool[pID.String()]
	return con, ok
}

func (p *ConnectionPool) AddConnInfo(pID peer.ID, connInfo models.ConnInfo) {

	// check if there is an existing pool connection to this peer
	conEv, ok := p.CheckIfOpenEventForPeer(pID)
	if !ok {
		log.Errorf("trying to insert conn-info for peer %s that was not registered", pID.String())
		// Create a new on otherwise
		conEv = models.NewConnEvent(pID)
		// keep it to the
		p.connPool[pID.String()] = conEv
	}
	// add the info gathered from the connection
	conEv.AddConnInfo(connInfo)

	// is it ready to persist?
	if conEv.IsReadyToPersist() {
		p.PersistToDB(pID)
		// remove the peer connecion from the pool
		p.removeConnEvent(conEv.PeerID)
	}
}

func (p *ConnectionPool) AddDisconn(pID peer.ID, discEv models.EndConnInfo) {
	// check if there is an existing pool connection to this peer
	conEv, ok := p.CheckIfOpenEventForPeer(pID)
	if !ok {
		log.Errorf("trying to insert conn-info for peer %s that was not registered", pID.String())
		// Create a new on otherwise
		conEv = models.NewConnEvent(pID)
		// keep it to the
		p.connPool[pID.String()] = conEv
	}
	// add the info gathered from the connection
	conEv.AddDisconn(discEv)

	// is it ready to persist?
	if conEv.IsReadyToPersist() {
		p.PersistToDB(pID)
		// remove the peer connecion from the pool
		p.removeConnEvent(conEv.PeerID)
	}
}

func (p *ConnectionPool) PersistToDB(pID peer.ID) {
	p.dbClient.PersistToDB(p.connPool[pID.String()])
}

func (p *ConnectionPool) removeConnEvent(pID peer.ID) {
	delete(p.connPool, pID.String())
}
