package postgresql

import (
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
)

var (
	// Create Peer table
	createFilecoinPeerTable = `
	CREATE TABLE IF NOT EXISTS t_filecoin_peers(
		f_peer_id TEXT,
		f_user_agent TEXT,
		f_ip TEXT,
		f_country TEXT,
		f_country_code TEXT,
		f_city TEXT,
		f_multi_addrs TEXT[],
		f_protocols TEXT[],
		f_protocol_version TEXT,
	
		PRIMARY KEY (f_peer_id)
	);
	`
	insertFilecoinPeer = `
	INSERT INTO t_filecoin_peers(
		f_peer_id,
		f_user_agent,
		f_ip,
		f_country,
		f_country_code,
		f_city,
		f_multi_addrs,
		f_protocols,
		f_protocol_version)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (f_peer_id)
	DO UPDATE SET
		f_peer_id=EXCLUDED.f_peer_id,
		f_user_agent=EXCLUDED.f_user_agent,
		f_ip=EXCLUDED.f_ip,
		f_country=EXCLUDED.f_country,
		f_country_code=EXCLUDED.f_country_code,
		f_city=EXCLUDED.f_city,
		f_multi_addrs=EXCLUDED.f_multi_addrs,
		f_protocols=EXCLUDED.f_protocols,
		f_protocol_version=EXCLUDED.f_protocol_version
	`
)

// creates Peer table in the Postgres DB
func (p *PostgresDBService) createFilecoinPeerTable() error {
	_, err := p.psqlPool.Exec(p.ctx, createFilecoinPeerTable)
	if err != nil {
		return errors.Wrap(err, "error creating peers summary table")
	}
	return nil
}

func (p *PostgresDBService) StoreFilecoinPeer(peerID string, peer models.FilecoinPeer) {
	_, err := p.psqlPool.Exec(
		p.ctx,
		insertFilecoinPeer,
		peerID,
		peer.UserAgent,
		peer.Ip,
		peer.Country,
		peer.CountryCode,
		peer.City,
		peer.MAddrs,
		peer.Protocols,
		peer.ProtocolVersion,
	)

	if err != nil {
		// TODO: Add error return value? will need memmory and boltdb update
		log.Errorf("error inserting peer in the psqldb %s", err.Error())
	}
}

func (p *PostgresDBService) LoadFilecoinPeer(peerID string) (models.FilecoinPeer, bool) {
	log.Debugf("loading peer %s", peerID)
	row := p.psqlPool.QueryRow(
		p.ctx,
		"SELECT *FROM t_filecoin_peers WHERE f_peer_id=$1",
		peerID,
	)
	peer := models.NewFilecoinPeer("")

	var multiAddrs []string
	err := row.Scan(
		&peer.PeerId,
		&peer.UserAgent,
		&peer.Ip,
		&peer.Country,
		&peer.CountryCode,
		&peer.City,
		&multiAddrs,
		&peer.Protocols,
		&peer.ProtocolVersion,
	)

	if err != nil {
		log.Debugf("error loading info from peer %s. %s", peerID, err.Error())
		return peer, false
	}

	// recompose the multiaddresses from []string
	for _, ma := range multiAddrs {
		maddres, err := utils.UnmarshalMaddr(ma)
		if err != nil {
			log.Warnf("unable to generate multiaddres from %s. %s", ma, err.Error())
			continue
		}
		peer.MAddrs = append(peer.MAddrs, maddres)
	}
	log.Debugf("load success of peer %s", peerID)
	return peer, true
}

func (p *PostgresDBService) DeleteFilecoinPeer(peerID string) {
	log.Debugf("deleting item")
	// Delete peer item from the table
	_, _ = p.psqlPool.Exec(
		p.ctx,
		"DELETE FROM t_filecoin_peers WHERE f_peer_id=$1",
		peerID,
	)
}

func (p *PostgresDBService) GetFilecoinPeers() []peer.ID {
	rows, err := p.psqlPool.Query(p.ctx,
		"SELECT f_peer_id FROM t_filecoin_peers")
	var peerList []peer.ID
	if err != nil {
		return peerList
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return peerList
		}
		for _, pID := range values {
			peerID_obj, err := peer.Decode(pID.(string))
			if err != nil {
				continue
			}
			peerList = append(peerList, peerID_obj)
		}
	}
	return peerList
}

func (p *PostgresDBService) GetNumberOfFilecoinPeers() (int, error) {
	rows, err := p.psqlPool.Query(p.ctx,
		"SELECT f_peer_id FROM t_filecoin_peers")
	if err != nil {
		return -1, errors.Wrap(err, "unable to query the total amount of peers in the table")
	}
	var peerCounter int = 0
	for rows.Next() {
		peerCounter++
	}
	return peerCounter, nil
}
