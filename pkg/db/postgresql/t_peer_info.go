package postgresql

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/pkg/errors"
)

func (c *DBClient) InitPeerInfoTable() error {
	log.Debug("initializing peer_info table in db")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS peer_info(
			id SERIAL,
			peer_id TEXT NOT NULL,
			user_agent TEXT,
			client_name TEXT,
			client_version TEXT, 
			client_os TEXT,
			client_arch TEXT,
			multi_addrs TEXT[] NOT NULL,
			ip TEXT NOT NULL,
			protocol_version TEXT,
			sup_protocols []TEXT

			PRIMARY KEY (peer_id)
		);
		`)

	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	return nil
}

// InsertNewPeerInfo
func (c *DBClient) InsertNewPeerInfo(pInfo *models.PeerInfo) error {

	log.Debugf("inserting new peer %s", pInfo.ID.String())

	// filter UserAgent to get client name, version, os, and arch
	cliName, cliVers, cliOS, cliArch := utils.ParseClientType(c.Network, pInfo.UserAgent)
	_, err := c.psqlPool.Exec(c.ctx, `
		INSERT INTO peer_info (
			peer_id,
			user_agent,
			client_name,
			client_version,
			client_os,
			client_arch,
			multi_addrs,
			ip,
			protocol_version,
			sup_protocols)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		pInfo.ID.String(),
		pInfo.UserAgent,
		cliName,
		cliVers,
		cliOS,
		cliArch,
		pInfo.MAddrs,
		pInfo.IpInfo,
		pInfo.ProtocolVersion,
		pInfo.Protocols,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create tx to db")
	}

	return nil
}

func (c *DBClient) UpdatePeerInfo(pInfo *models.PeerInfo) error {
	log.Debugf("updating info for peer %s", pInfo.ID.String())

	// filter UserAgent to get client name, version, os, and arch
	cliName, cliVers, cliOS, cliArch := utils.ParseClientType(c.Network, pInfo.UserAgent)
	_, err := c.psqlPool.Exec(c.ctx, `
		UPDATE peer_info SET	 
			user_agent = $2,
			client_name = $3,
			client_version = $4,
			client_os = $5,
			client_arch = $6,
			multi_addrs = $7,
			ip = $8,
			protocol_version = $9,
			sup_protocols = $10
		WHERE peer_id = $1;
		`,
		pInfo.ID.String(),
		pInfo.UserAgent,
		cliName,
		cliVers,
		cliOS,
		cliArch,
		pInfo.MAddrs,
		pInfo.IpInfo.IP,
		pInfo.ProtocolVersion,
		pInfo.Protocols,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create tx to db")
	}

	return nil
}

func (c *DBClient) ReadPeerInfo(pID peer.ID) (*models.PeerInfo, error) {

	log.Debugf("reading info for peer %s", pID.String())

	// TODO: Shall I concatenate all the different tables?
	// I should Still keep a local list of Peers in the local PeerStore
	return &models.PeerInfo{}, nil
}

func (c *DBClient) PeerInfoExists(pID peer.ID) bool {

	// get the ip
	var id string

	err := c.psqlPool.QueryRow(c.ctx, `
		SELECT 
			peer_id
		FROM peer_info
		WHERE peer_id=$1;
	`, pID).Scan(&id)

	if err != nil {
		log.Debugf("peer %d doesn't exist", id)
		return false
	}
	log.Debugf("peer %d exists", id)
	return true
}
