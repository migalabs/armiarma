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
func (c *DBClient) UpsertPeerInfo(pInfo *models.PeerInfo) (q string, args []interface{}) {
	// compose the query
	q = `INSERT INTO peer_info (
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
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT ON CONSTRAINT peer_id
			UPDATE SET
			user_agent = excluded.user_agent,
			client_name = excluded.client_name,
			client_version = excluded.client_version,
			client_os = excluded.client_os,
			client_arch = excluded.client_arch,
			multi_addrs = excluded.multi_addrs,
			ip = excluded.ip,
			protocol_version = excluded.protocol_version,
			sup_protocols = excluded.sup_protocols;
		`

	// filter UserAgent to get client name, version, os, and arch
	cliName, cliVers, cliOS, cliArch := utils.ParseClientType(c.Network, pInfo.UserAgent)

	args = append(args, pInfo.ID.String())
	args = append(args, pInfo.UserAgent)
	args = append(args, cliName)
	args = append(args, cliVers)
	args = append(args, cliOS)
	args = append(args, cliArch)
	args = append(args, pInfo.MAddrs)
	args = append(args, pInfo.IpInfo)
	args = append(args, pInfo.ProtocolVersion)
	args = append(args, pInfo.Protocols)

	return q, args
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
