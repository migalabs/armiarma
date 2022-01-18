import postgres

import (
	pgx "github.com/jackc/pgx/v4"
)

// Postgres intregration variables
var (	
	createPeerTable = `
	CREATE TABLE IF NOT EXISTS t_
		f_peer_id TEXT,
		f_pubkey TEXT,
		f_node_id TEXT,
		f_user_agent TEXT,
		f_client_name TEXT,
		f_client_os TEXT,
		f_client_version TEXT,
		f_enr TEXT,
		f_ip TEXT,
		f_country TEXT,
		f_country_code TEXT,
		f_city TEXT,
		f_latency TEXT,
		f_multi_addrs TEXT[],
		
	`
)