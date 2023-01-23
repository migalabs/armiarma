package postgresql

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// InitPeerInfoTable compiles all the data needed and extractable from each peer
// it includes: HostInfo, PeerInfo, and ControlInfo
func (c *DBClient) InitPeerInfoTable() error {
	log.Debug("initializing peer_info table in db")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS peer_info(
			id SERIAL,
			peer_id TEXT NOT NULL,
			network TEXST NOT NULL,
			multi_addrs TEXT[] NOT NULL,
			ip TEXT NOT NULL,
			tcp INT,
			udp INT,

			user_agent TEXT,
			client_name TEXT,
			client_version TEXT, 
			client_os TEXT,
			client_arch TEXT,
			protocol_version TEXT,
			sup_protocols []TEXT,
			
			deprecated BOOL,
			left_network BOOL,
			attempted INT,
			last_activity BIGINT, 
			last_conn_attempt BIGINT,
			last_error TEXT,

			PRIMARY KEY (peer_id)
		);
		`)

	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	return nil
}

// InsertNewPeerInfo
func (c *DBClient) UpsertHostInfo(hInfo *models.HostInfo) (q string, args []interface{}) {
	// compose the query
	q = `INSERT INTO peer_info (
			peer_id,
			network,
			multi_addrs,
			ip,
			tcp,
			udp)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT ON CONSTRAINT peer_id
			UPDATE SET
			multi_addrs = excluded.multi_addrs,
			ip = excluded.ip
			tcp = excluded.tcp
			udp = excluded udp;
		`

	args = append(args, hInfo.ID.String())
	args = append(args, string(hInfo.Network))
	args = append(args, hInfo.MAddrs)
	args = append(args, hInfo.IP)
	args = append(args, hInfo.TCP)
	args = append(args, hInfo.UDP)

	return q, args
}

// InsertNewPeerInfo
func (c *DBClient) UpdatePeerInfo(pInfo *models.PeerInfo) (q string, args []interface{}) {
	// compose the query
	q = `
		UPDATE peer_info
		SET
			user_agent=$2,
			client_name=$3,
			client_version=$4,
			client_os=$5,
			client_arch=$6,
			protocol_version=$7,
			sup_protocols=$8,
			latency=$9
		WHERE peer_id=$1;
		`

	// filter UserAgent to get client name, version, os, and arch
	cliName, cliVers, cliOS, cliArch := utils.ParseClientType(c.Network, pInfo.UserAgent)

	args = append(args, pInfo.RemotePeer.String())
	args = append(args, pInfo.UserAgent)
	args = append(args, cliName)
	args = append(args, cliVers)
	args = append(args, cliOS)
	args = append(args, cliArch)
	args = append(args, pInfo.ProtocolVersion)
	args = append(args, pInfo.Protocols)
	args = append(args, pInfo.Latency.Milliseconds())

	return q, args
}

func (c *DBClient) UpdateConnAttempt(connAttempt *models.ConnectionAttempt) (query string, args []interface{}) {
	// logic to determine how to update the table
	if connAttempt.Status == models.PossitiveAttempt {
		// we have the chance to un-deprecate/un-left_network the peer
		query = `
				UPDATE peer_info
				SET 
					deprecated=$2,
					left_network=$3,
					attempted=$4,
					last_activity=$5, 
					last_conn_attempt=$5,
					last_error=$6,
				WHERE peer_id=$1;
			`

		args = append(args, connAttempt.RemotePeer.String())
		args = append(args, false)                        // Un-Deprecate peer
		args = append(args, false)                        // Un-LeftNetwork peer
		args = append(args, true)                         // Connection was attempted
		args = append(args, connAttempt.Timestamp.Unix()) // attempt timestamp (same as our new last activity)
		args = append(args, connAttempt.Error)            //
	} else {
		query = `
			UPDATE peer_info
			SET 
				deprecated=$2,
				left_network=$3,
				attempted=$4,
				last_conn_attempt=$5,
				last_error=$6,
			WHERE peer_id=$1;
		`
		args = append(args, true) // connection attempted
		args = append(args, connAttempt.Deprecable)
		args = append(args, connAttempt.LeftNetwork)
		args = append(args, connAttempt.Timestamp.Unix())
		args = append(args, connAttempt.Error)
	}

	return query, args
}

func (c *DBClient) GetFullHostInfo(pID peer.ID) (*models.HostInfo, error) {

	log.Debugf("reading info for peer %s", pID.String())

	hInfo := models.NewHostInfo(pID, utils.EthereumNetwork)
	pInfo := models.NewEmptyPeerInfo()
	cInfo := models.NewControlInfo()
	pInfo.RemotePeer = pID

	var maddresses []string
	var lastActivity int64
	var lastConnAttempt int64
	var latencyMillis int64

	// read the Peer from the SQL database
	err := c.psqlPool.QueryRow(c.ctx, `
		SELECT
			network,
			multiaddrs,
			ip,
			tcp,
			udp,
			user_agent,
			procol_version,
			sup_protocols,
			latency,
			deprecated,
			left_network,
			attempted,
			last_activity,
			last_conn_attempt,
			last_error,
		FROM peer_info
		WHERE peer_id=$1;
	`, pID.String()).Scan(
		&hInfo.Network,
		&maddresses,
		&hInfo.IP,
		&hInfo.TCP,
		&hInfo.UDP,
		&pInfo.UserAgent,
		&pInfo.ProtocolVersion,
		&pInfo.Protocols,
		&latencyMillis,
		&cInfo.Deprecated,
		&cInfo.LeftNetwork,
		&cInfo.Attempted,
		&lastActivity,
		&lastConnAttempt,
		&cInfo.LastError,
	)
	// Check if there was any error reading the peer from the SQL table
	if err != nil {
		return &models.HostInfo{}, errors.Wrap(err, "unable to retrieve full peer_info")
	}

	// parse the multiaddresses from the []string
	var mAddrs []ma.Multiaddr
	// translate []string to maddress
	for _, maStr := range maddresses {
		mAddr, err := ma.NewMultiaddr(maStr)
		if err != nil {
			return &models.HostInfo{}, errors.Wrap(err, "unable to parse mAddrs reading full peer_info")
		}
		mAddrs = append(mAddrs, mAddr)
	}

	// parse times from received Unix() timestamps
	cInfo.LastActivity = time.Unix(lastActivity, int64(0))
	cInfo.LastConnAttempt = time.Unix(lastConnAttempt, int64(0))
	// parse latency in millisecods
	pInfo.Latency = time.Duration(latencyMillis) * time.Millisecond

	hInfo.MAddrs = mAddrs
	hInfo.PeerInfo = *pInfo
	hInfo.ControlInfo = *cInfo

	return hInfo, nil
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

func (c *DBClient) UpdateLastActivityTimestamp(peerID peer.ID, t time.Time) (query string, args []interface{}) {

	query = `
		UPDATE peer_info
		SET
			last_activity=$2
		WHERE peer_id=$1;
	`

	args = append(args, peerID.String())
	args = append(args, t.Unix())

	return query, args
}

func (c *DBClient) GetPeersInNetwork() ([]peer.ID, error) {
	log.Tracef("retrieving the list of peer_ids from the DB that are not classified as left_network\n")

	var peerIDs []peer.ID

	rows, err := c.psqlPool.Query(c.ctx, `
		SELECT
			peer_id
		FROM peer_info
		WHERE left_network="false";`)
	if err != nil {
		return peerIDs, errors.Wrap(err, "unable to retrieve peers in the network")
	}

	for rows.Next() {
		pIDs, err := rows.Values()
		if err != nil {
			return peerIDs, err
		}
		for _, pIDStr := range pIDs {
			// decode peerID to have proper OBJ
			peerID, err := peer.Decode(pIDStr.(string))
			if err != nil {
				log.Errorf("unable to get peerID from DB %s\n", pIDStr)
				continue // if error, go for the next one
			}
			peerIDs = append(peerIDs, peerID)
		}
	}
	return peerIDs, nil
}

func (c *DBClient) GetNonDeprecatedPeers() ([]peer.ID, error) {
	log.Tracef("retrieving the list of peer_ids from the DB that are not deprecated\n")
	var peerIDs []peer.ID

	rows, err := c.psqlPool.Query(c.ctx, `
		SELECT
			peer_id
		FROM peer_info
		WHERE left_network="false" and deprecated="false";`)
	if err != nil {
		return peerIDs, errors.Wrap(err, "unable to retrieve peers in the network")
	}

	for rows.Next() {
		pIDs, err := rows.Values()
		if err != nil {
			return peerIDs, err
		}
		for _, pIDStr := range pIDs {
			// decode peerID to have proper OBJ
			peerID, err := peer.Decode(pIDStr.(string))
			if err != nil {
				log.Errorf("unable to get peerID from DB %s\n", pIDStr)
				continue // if error, go for the next one
			}
			peerIDs = append(peerIDs, peerID)
		}
	}
	return peerIDs, nil
}
