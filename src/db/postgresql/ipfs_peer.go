package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
)

// Postgres intregration variables
var (
	createIpfsPeerTable = `
	CREATE TABLE IF NOT EXISTS t_ipfs_summary(
		f_peer_id TEXT,

		f_user_agent TEXT,
		f_client_name TEXT,
		f_client_os TEXT,
		f_client_version TEXT,

		f_ip TEXT,
		f_country TEXT,
		f_country_code TEXT,
		f_city TEXT,
		f_latency FLOAT8,
		f_multi_addrs TEXT[],
		f_protocols TEXT[],
		f_protocol_version TEXT,

		f_connected_direction TEXT[],
		f_is_connected BOOL,
		f_attempted BOOL,
		f_succeed BOOL,
		f_attempts BIGINT,
		f_error TEXT[],
		f_last_error_timestamp TIMESTAMP,
		f_deprecated BOOL,
		f_last_identify_timestamp TIMESTAMP,

		f_negative_conn_attempts TIMESTAMP[],
		f_connection_times TIMESTAMP[],
		f_disconnection_times TIMESTAMP[],
		f_metadata_request BOOL,
		f_metadata_succeed BOOL,

		f_last_export BIGINT,

		PRIMARY KEY (f_peer_id)
	);
	`
	insertIpfsPeer = `
	INSERT INTO t_ipfs_summary(
		f_peer_id,
		f_user_agent,
		f_client_name,
		f_client_os,
		f_client_version,
		f_ip,
		f_country,
		f_country_code,
		f_city,
		f_latency,
		f_multi_addrs,
		f_protocols,
		f_protocol_version,
		f_connected_direction,
		f_is_connected,
		f_attempted,
		f_succeed,
		f_attempts,
		f_error,
		f_last_error_timestamp,
		f_deprecated,
		f_last_identify_timestamp,
		f_negative_conn_attempts,
		f_connection_times,
		f_disconnection_times,
		f_metadata_request,
		f_metadata_succeed,
		f_last_export)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
	ON CONFLICT (f_peer_id)
	DO UPDATE SET
		f_user_agent=EXCLUDED.f_user_agent,
		f_client_name=EXCLUDED.f_client_name,
		f_client_os=EXCLUDED.f_client_os,
		f_client_version=EXCLUDED.f_client_version,
		f_ip=EXCLUDED.f_ip,
		f_country=EXCLUDED.f_country,
		f_country_code=EXCLUDED.f_country_code,
		f_city=EXCLUDED.f_city,
		f_latency=EXCLUDED.f_latency,
		f_multi_addrs=EXCLUDED.f_multi_addrs,
		f_protocols=EXCLUDED.f_protocols,
		f_protocol_version=EXCLUDED.f_protocol_version,
		f_connected_direction=EXCLUDED.f_connected_direction,
		f_is_connected=EXCLUDED.f_is_connected,
		f_attempted=EXCLUDED.f_attempted,
		f_succeed=EXCLUDED.f_succeed,
		f_attempts=EXCLUDED.f_attempts,
		f_error=EXCLUDED.f_error,
		f_last_error_timestamp=EXCLUDED.f_last_error_timestamp,
		f_deprecated=EXCLUDED.f_deprecated,
		f_last_identify_timestamp=EXCLUDED.f_last_identify_timestamp,
		f_negative_conn_attempts=EXCLUDED.f_negative_conn_attempts,
		f_connection_times=EXCLUDED.f_connection_times,
		f_disconnection_times=EXCLUDED.f_disconnection_times,
		f_metadata_request=EXCLUDED.f_metadata_request,
		f_metadata_succeed=EXCLUDED.f_metadata_succeed,
		f_last_export=EXCLUDED.f_last_export
		`
)

//
type IpfsModel struct {
	Network string
}

//
func NewIpfsPeerModel(network string) IpfsModel {
	return IpfsModel{
		Network: network,
	}
}

func (p *IpfsModel) init(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	err := p.createPeerTable(ctx, pool)
	if err != nil {
		return err
	}
	// get status of the peer summary table
	ok := p.CheckPeersSummaryTableStatus(ctx, pool)
	if !ok {
		return errors.New("unable to check existing connected peers in the postgres db")
	}

	// ---- Client Diversity Table ----
	// TODO: not implemented yet
	/*
		err = p.createClientDiversityTable(ctx, pool)
		if err != nil {
			return err
		}
	*/
	return nil
}

// creates Peer table in the Postgres DB
func (p *IpfsModel) createPeerTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createIpfsPeerTable)
	if err != nil {
		return errors.Wrap(err, "error creating peers summary table")
	}
	return nil
}

func (p *IpfsModel) StorePeer(ctx context.Context, pool *pgxpool.Pool, peerID string, peer models.Peer) {

	// store the peer
	_, err := pool.Exec(
		ctx,
		insertIpfsPeer,
		peerID,
		peer.UserAgent,
		peer.ClientName,
		peer.ClientOS,
		peer.ClientVersion,
		peer.Ip,
		peer.Country,
		peer.CountryCode,
		peer.City,
		peer.Latency,
		peer.MAddrs,
		peer.Protocols,
		peer.ProtocolVersion,
		peer.ConnectedDirection,
		peer.IsConnected,
		peer.Attempted,
		peer.Succeed,
		peer.Attempts,
		peer.Error,
		peer.LastErrorTimestamp,
		peer.Deprecated,
		peer.LastIdentifyTimestamp,
		peer.NegativeConnAttempts,
		peer.ConnectionTimes,
		peer.DisconnectionTimes,
		peer.MetadataRequest,
		peer.MetadataSucceed,
		peer.LastExport,
	)

	if err != nil {
		// TODO: Add error return value? will need memmory and boltdb update
		log.Errorf("error inserting peer in the psqldb %s", err.Error())
	}

}

func (p *IpfsModel) LoadPeer(ctx context.Context, pool *pgxpool.Pool, peerID string) (models.Peer, bool) {
	log.Debugf("loading peer %s", peerID)
	row := pool.QueryRow(
		ctx,
		"SELECT *FROM t_ipfs_summary WHERE f_peer_id=$1",
		peerID,
	)
	peer := models.NewPeer("")

	var multiAddrs []string

	err := row.Scan(
		&peer.PeerId,
		&peer.UserAgent,
		&peer.ClientName,
		&peer.ClientOS,
		&peer.ClientVersion,
		&peer.Ip,
		&peer.Country,
		&peer.CountryCode,
		&peer.City,
		&peer.Latency,
		&multiAddrs,
		&peer.Protocols,
		&peer.ProtocolVersion,
		&peer.ConnectedDirection,
		&peer.IsConnected,
		&peer.Attempted,
		&peer.Succeed,
		&peer.Attempts,
		&peer.Error,
		&peer.LastErrorTimestamp,
		&peer.Deprecated,
		&peer.LastIdentifyTimestamp,
		&peer.NegativeConnAttempts,
		&peer.ConnectionTimes,
		&peer.DisconnectionTimes,
		&peer.MetadataRequest,
		&peer.MetadataSucceed,
		&peer.LastExport,
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

func (p *IpfsModel) DeletePeer(ctx context.Context, pool *pgxpool.Pool, peerID string) {
	log.Debugf("deleting item")
	// Delete peer item from the table
	_, _ = pool.Exec(
		ctx,
		"DELETE FROM t_ipfs_summary WHERE f_peer_id=$1",
		peerID,
	)
}

func (p *IpfsModel) GetPeers(ctx context.Context, pool *pgxpool.Pool) []peer.ID {
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_ipfs_summary")
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

func (p *IpfsModel) CheckPeersSummaryTableStatus(ctx context.Context, pool *pgxpool.Pool) bool {
	// check the last activity of the tool
	lastActivity, err := p.GetLastActivityTime(ctx, pool)
	if err != nil {
		log.Error(err, "unable to get last activity of the tool")
		return false
	}
	log.Info("Last Activity of the tool has been recorded to be", lastActivity)
	connectedPeers, err := p.GetConnectedPeers(ctx, pool)
	if err != nil {
		log.Errorf("unable to get connected peers. %s", err.Error())
		return false
	}

	var cnt int = 0
	for _, peerID := range connectedPeers {
		peer, ok := p.LoadPeer(ctx, pool, peerID)
		if !ok {
			log.Errorf("unable to load info from peer %s", peerID)
			return false
		}
		peer.DisconnectionEvent(lastActivity)
		p.StorePeer(ctx, pool, peerID, peer)
		if !ok {
			log.Errorf("unable to store info from peer %s", peerID)
			return false
		}
		cnt++
	}
	nPeers, err := p.GetNumberOfPeers(ctx, pool)
	if err != nil {
		log.Errorf("unable to get number of peers in db. %s", err.Error())
		return false
	}
	log.Infof("loaded psql db with %d peers on it (%d connected)", nPeers, cnt)
	return true
}

func (p *IpfsModel) GetConnectedPeers(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	var connPeers []string
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_ipfs_summary WHERE f_is_connected='true'",
	)
	if err != nil {
		return connPeers, errors.Wrap(err, "unable to query peers with IsConnected flag = true")
	}
	// Iterate through the peers whos flag IsConnecte=True
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return connPeers, errors.Wrap(err, "unable to parse PeerID of peer with IsConnected=true")
		}
		// aggregate the disconnection for the peer that was connected
		for _, peerID := range values {
			connPeers = append(connPeers, peerID.(string))
		}
	}

	return connPeers, nil
}

func (p *IpfsModel) GetNumberOfPeers(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_ipfs_summary")
	if err != nil {
		return -1, errors.Wrap(err, "unable to query the total amount of peers in the table")
	}
	var peerCounter int = 0
	for rows.Next() {
		peerCounter++
	}
	return peerCounter, nil
}

func (p *IpfsModel) GetLastActivityTime(ctx context.Context, pool *pgxpool.Pool) (time.Time, error) {
	var lastActivity time.Time
	query := `
		WITH foo as (
			SELECT UNNEST(f_negative_conn_attempts) AS a_last_activity
			FROM t_ipfs_summary
		)
		SELECT MAX(a_last_activity) as a_last_activity
		FROM foo;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return lastActivity, errors.Wrap(err, "unable to get Last Activity of the tool from postgresql db, empty DB or query failed")
	}
	for rows.Next() {
		var t time.Time
		err := rows.Scan(&t)
		if err != nil {
			// check if DB is empty or if there are actual values
			peers := p.GetPeers(ctx, pool)
			if len(peers) == 0 {
				log.Info("Empty loaded DB")
				return time.Time{}, nil
			} else {
				continue
				//return time.Time{}, errors.Wrap(err, "unable to parse Last Activity of the tool from postgresql db, just didn't work, dunno why")
			}
		}
		if t.After(lastActivity) {
			lastActivity = t
		}
	}
	return lastActivity, nil
}

// Client diversity
// TODO: fill it

func (p *IpfsModel) StoreClientDiversitySnapshot(ctx context.Context, pool *pgxpool.Pool, cliDiver models.ClientDiversity) error {
	log.Debugf("ipfs - adding new client diversity item in psql")
	return nil
}

// So far not used since it's just for exporting
// Doesn't make much sense to add it to the crawler (no idea why would we need to access the snapshot of a given time)
func (p *IpfsModel) LoadClientDiversitySnapshot(ctx context.Context, pool *pgxpool.Pool, qTime time.Time) (models.ClientDiversity, error) {
	log.Debugf("ipfs - Loading client diversity of time %s", qTime)
	return models.ClientDiversity{}, nil
}
