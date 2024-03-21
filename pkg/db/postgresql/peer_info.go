package postgresql

import (
	"time"

	pgx "github.com/jackc/pgx/v4"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// InitPeerInfoTable compiles all the data needed and extractable from each peer
// it includes: HostInfo, PeerInfo, and ControlInfo
func (c *DBClient) InitPeerInfoTable() error {
	log.Debug("initializing peer_info table in psql-db")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS peer_info(
			id SERIAL,
			peer_id TEXT NOT NULL,
			network TEXT NOT NULL,
			multi_addrs TEXT[] NOT NULL,
			ip TEXT NOT NULL,
			port INT,

			user_agent TEXT,
			client_name TEXT,
			client_version TEXT, 
			client_os TEXT,
			client_arch TEXT,
			protocol_version TEXT,
			sup_protocols TEXT[],
			latency INT,
			
			deprecated BOOL,
			attempted BOOL,
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
	log.Trace("upserting host in peer_info table")
	// compose the query
	q = `INSERT INTO peer_info (
			peer_id,
			network,
			multi_addrs,
			ip,
			port,
			deprecated)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (peer_id)
		DO UPDATE SET
			multi_addrs = excluded.multi_addrs,
			ip = excluded.ip,
			port = excluded.port,
			deprecated = excluded.deprecated;
		`

	args = append(args, hInfo.ID.String())
	args = append(args, string(hInfo.Network))
	args = append(args, hInfo.MAddrs)
	args = append(args, hInfo.IP)
	args = append(args, hInfo.Port)
	args = append(args, false)

	return q, args
}

// InsertNewPeerInfo
func (c *DBClient) UpdatePeerInfo(pInfo *models.PeerInfo) (q string, args []interface{}) {
	log.Trace("upserting peer in peer_info table")
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
	log.Tracef("updating peer_info because of new conn attempt %+v", connAttempt)
	// logic to determine how to update the table
	if connAttempt.Status == models.PossitiveAttempt {
		// we have the chance to un-deprecate the peer
		query = `
				UPDATE peer_info
				SET 
					deprecated=$2,
					attempted=$3,
					last_activity=$4, 
					last_conn_attempt=$5,
					last_error=$6
				WHERE peer_id=$1;
			`
		args = append(args, connAttempt.RemotePeer.String())
		args = append(args, false)                        // Un-Deprecate peer
		args = append(args, true)                         // Connection was attempted
		args = append(args, connAttempt.Timestamp.Unix()) // attempt timestamp (same as our new last activity)
		args = append(args, connAttempt.Timestamp.Unix()) // attempt timestamp (same as our new last activity)
		args = append(args, connAttempt.Error)
	} else {
		query = `
			UPDATE peer_info
			SET 
				deprecated=$2,
				attempted=$3,
				last_conn_attempt=$4,
				last_error=$5
			WHERE peer_id=$1;
		`
		args = append(args, connAttempt.RemotePeer.String())
		args = append(args, connAttempt.Deprecable)
		args = append(args, true) // connection attempted
		args = append(args, connAttempt.Timestamp.Unix())
		args = append(args, connAttempt.Error)
	}

	return query, args
}

func (c *DBClient) GetFullHostInfo(pID peer.ID) (*models.HostInfo, error) {
	log.Tracef("reading info for peer %s", pID.String())

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
			multi_addrs,
			ip,
			port,
			user_agent,
			protocol_version,
			sup_protocols,
			latency,
			deprecated,
			attempted,
			last_activity,
			last_conn_attempt,
			last_error
		FROM peer_info
		WHERE peer_id=$1;
	`, pID.String()).Scan(
		&hInfo.Network,
		&maddresses,
		&hInfo.IP,
		&hInfo.Port,
		&pInfo.UserAgent,
		&pInfo.ProtocolVersion,
		&pInfo.Protocols,
		&latencyMillis,
		&cInfo.Deprecated,
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

func (c *DBClient) GetPersistable(pID string) (models.RemoteConnectablePeer, error) {
	log.Tracef("reading persistable info for peer %s", pID)

	var maddresses []string
	var network utils.NetworkType

	// read the Peer from the SQL database
	err := c.psqlPool.QueryRow(c.ctx, `
		SELECT
			network,
			multi_addrs
		FROM peer_info
		WHERE peer_id=$1;
	`, pID).Scan(
		&network,
		&maddresses,
	)
	// Check if there was any error reading the peer from the SQL table
	if err != nil && err != pgx.ErrNoRows {
		return models.RemoteConnectablePeer{}, errors.Wrap(err, "unable to retrieve full peer_info")
	}

	// parse the multiaddresses from the []string
	var mAddrs []ma.Multiaddr
	// translate []string to maddress
	for _, maStr := range maddresses {
		mAddr, err := ma.NewMultiaddr(maStr)
		if err != nil {
			return models.RemoteConnectablePeer{}, errors.Wrap(err, "unable to parse mAddrs reading full peer_info")
		}
		mAddrs = append(mAddrs, mAddr)
	}

	peerID, err := peer.Decode(pID)
	if err != nil {
		return models.RemoteConnectablePeer{}, err
	}

	connectable := models.NewRemoteConnectablePeer(
		peerID,
		mAddrs,
		network,
	)
	return *connectable, nil
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

	if err != nil && err != pgx.ErrNoRows {
		log.Debugf("peer %s doesn't exist", id)
		return false
	}
	log.Debugf("peer %s exists", id)
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

func (c *DBClient) GetNonDeprecatedPeers() ([]*models.RemoteConnectablePeer, error) {
	log.Tracef("retrieving the list of peer_ids from the DB that are not deprecated\n")
	var connectPeers []*models.RemoteConnectablePeer

	rows, err := c.psqlPool.Query(c.ctx, `
		SELECT
			peer_id,
			network,
			multi_addrs
		FROM peer_info
		WHERE deprecated='false';`)

	// If there are no rows, don't panic
	if err != nil && err != pgx.ErrNoRows {
		return connectPeers, errors.Wrap(err, "unable to retrieve peers in the network")
	}
	defer rows.Close()

	for rows.Next() {
		var peerIDStr string
		var mAddrsStr []string
		var networkStr string

		err := rows.Scan(&peerIDStr, &networkStr, &mAddrsStr)
		if err != nil {
			return connectPeers, err
		}

		// persist peerID
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			log.Errorf("unable to get peerID from DB %s \n", peerIDStr)
			continue // if error, go for the next one
		}

		// parse the network type
		network := utils.NetworkType(networkStr)

		// parse the multiaddress
		maddrs := make([]ma.Multiaddr, 0, len(mAddrsStr))
		for _, element := range mAddrsStr {
			mAddr, err := ma.NewMultiaddr(element)
			if err != nil {
				log.Error(errors.Wrap(err, "unable to parse mAddrs reading full peer_info"))
				continue
			}
			maddrs = append(maddrs, mAddr)
		}
		// create the persistable instance
		connectable := models.NewRemoteConnectablePeer(
			peerID,
			maddrs,
			network,
		)

		// decode peerID to have proper OBJ
		connectPeers = append(connectPeers, connectable)
	}
	return connectPeers, nil
}
