# Armiarma Database Schema

This document describes the schema of the database used by Armiarma. It covers all tables, their columns and indexes.

## Overview

The Armiarma database tracks discovered peers, connection attempts, Ethereum-specific metadata (ENRs, status, blocks, attestations), IP geolocation and periodic snapshots of active peers. Key tables include:

* **peer\_info**: Basic libp2p peer identification and capabilities
* **eth\_nodes**: All ENR‐discovered nodes (both successful and failed connection attempts)
* **eth\_status**: Ethereum status for peers successfully connected
* **eth\_blocks**: Persisted beacon block messages
* **eth\_attestations**: Persisted beacon attestation messages
* **conn\_events**: Detailed connection and disconnection events
* **active\_peers**: Snapshots of active peer IDs over time
* **ips**: IP geolocation lookup cache

## peer\_info

Detailed information discovered via the libp2p Identify protocol. Updated whenever a peer is successfully identified.

| Column              | Type    | Nullable | Description                               |
| ------------------- | ------- | -------- | ----------------------------------------- |
| id                  | integer | not null | Internal sequence ID                      |
| peer\_id            | text    | not null | libp2p PeerID (primary key)               |
| network             | text    | not null | Network name (e.g. "eth2")                |
| multi\_addrs        | text\[] | not null | Multiaddresses announced by the peer      |
| ip                  | text    | not null | Peer IP address                           |
| port                | integer |          | Peer TCP port                             |
| user\_agent         | text    |          | libp2p user agent string                  |
| client\_name        | text    |          | Extracted client name (e.g. "Lighthouse") |
| client\_version     | text    |          | Client version string                     |
| client\_os          | text    |          | Operating system reported                 |
| client\_arch        | text    |          | CPU architecture reported                 |
| protocol\_version   | text    |          | libp2p protocol version                   |
| sup\_protocols      | text\[] |          | Supported sub‐protocols                   |
| latency             | integer |          | Round‐trip time to peer (ms)              |
| deprecated          | boolean |          | Flag for long‐inactive peers              |
| attempted           | boolean |          | Whether we've attempted connection        |
| last\_activity      | bigint  |          | Timestamp of last disconnection event     |
| last\_conn\_attempt | bigint  |          | Timestamp of last dial attempt            |
| last\_error         | text    |          | Error code of last failed dial            |

**Indexes:**

* `peer_info_pkey` PRIMARY KEY on (`peer_id`)

**Notes:**

* `last_activity` and `last_conn_attempt` are stored as Unix‐epoch bigint values.
* `last_error` records the categorized dial error returned by libp2p.

## eth\_nodes

All ENR entries discovered via discv5, regardless of connection outcome.

| Column              | Type    | Nullable | Description                               |
| ------------------- | ------- | -------- | ----------------------------------------- |
| id                  | integer | not null | Internal sequence ID                      |
| timestamp           | bigint  | not null | Discovery timestamp (Unix‐epoch)          |
| peer\_id            | text    |          | libp2p PeerID if ENR pubkey convertible   |
| node\_id            | text    | not null | ENR node ID (public key fingerprint) (PK) |
| seq                 | bigint  | not null | ENR sequence number                       |
| ip                  | text    | not null | ENR‐advertised IP address                 |
| tcp                 | integer |          | ENR‐advertised TCP port                   |
| udp                 | integer |          | ENR‐advertised UDP port                   |
| pubkey              | text    | not null | ENR‐advertised secp256k1 pubkey           |
| fork\_digest        | text    |          | `eth2` fork digest from ENR               |
| next\_fork\_version | text    |          | Next fork version                         |
| next\_fork\_epoch   | bigint  |          | Next fork epoch                           |
| attnets             | text    |          | Bitmask of attestation subnets (hex)      |
| attnets\_number     | integer |          | Number of set bits in `attnets`           |

**Indexes:**

* `eth_nodes_pkey` PRIMARY KEY on (`node_id`)
* `eth_nodes_peer_id_pubkey_key` UNIQUE on (`peer_id`, `pubkey`)

**Notes:**

* Even failed connection attempts yield an ENR entry here.
* `attnets_number` is computed by counting bits in the raw `attnets` byte array.

## eth\_status

Latest Ethereum status retrieved via the beacon RPC from peers that were successfully connected.

| Column           | Type    | Nullable | Description                         |
| ---------------- | ------- | -------- | ----------------------------------- |
| id               | integer | not null | Internal sequence ID                |
| peer\_id         | text    | not null | libp2p PeerID (PK)                  |
| timestamp        | bigint  |          | Status timestamp (Unix‐epoch)       |
| fork\_digest     | text    |          | Current `eth2` fork digest          |
| finalized\_root  | text    |          | Finalized block root                |
| finalized\_epoch | bigint  |          | Finalized epoch                     |
| head\_root       | text    |          | Head block root                     |
| head\_slot       | bigint  |          | Head slot                           |
| seq\_number      | bigint  |          | Metadata sequence number            |
| attnets          | text    |          | Attestation subnet bitmask (hex)    |
| syncnets         | text    |          | Sync committee subnet bitmask (hex) |

**Indexes:**

* `eth_status_pkey` PRIMARY KEY on (`peer_id`)

**Notes:**

* Only peers with at least one successful dial are inserted/updated here.
* Mirrors fields defined by the `common.Status` SSZ struct.

## eth\_blocks

Beacon block messages observed on gossipsub.

| Column         | Type      | Nullable | Description                               |
| -------------- | --------- | -------- | ----------------------------------------- |
| id             | integer   |          | Internal sequence ID                      |
| msg\_id        | text      | not null | Unique message identifier (SSZ hash) (PK) |
| sender         | text      | not null | libp2p PeerID of publisher                |
| slot           | bigint    | not null | Beacon slot number                        |
| arrival\_time  | timestamp | not null | Local timestamp when block arrived        |
| time\_in\_slot | real      | not null | Seconds offset within slot                |
| val\_idx       | bigint    |          | Validator index that proposed block       |

**Indexes:**

* `eth_blocks_pkey` PRIMARY KEY on (`msg_id`)

## eth\_attestations

Beacon attestation messages observed on gossipsub.

| Column         | Type    | Nullable | Description                               |
| -------------- | ------- | -------- | ----------------------------------------- |
| id             | integer |          | Internal sequence ID                      |
| msg\_id        | text    | not null | Unique message identifier (SSZ hash) (PK) |
| sender         | text    | not null | libp2p PeerID of publisher                |
| subnet         | integer | not null | Attestation subnet ID                     |
| slot           | bigint  | not null | Beacon slot number                        |
| arrival\_time  | time    | not null | Local time when attestation arrived       |
| time\_in\_slot | real    | not null | Seconds offset within slot                |
| val\_pubkey    | text    |          | Validator public key                      |

**Indexes:**

* `eth_attestations_pkey` PRIMARY KEY on (`msg_id`)

## conn\_events

Detailed connection/disconnection events captured by the host.

| Column        | Type    | Nullable | Description                            |
| ------------- | ------- | -------- | -------------------------------------- |
| id            | integer | not null | Internal sequence ID (PK)              |
| peer\_id      | text    | not null | libp2p PeerID                          |
| direction     | text    | not null | "inbound" or "outbound"                |
| conn\_time    | bigint  | not null | Connect timestamp (Unix‐epoch)         |
| latency       | bigint  |          | Time between connect and identify (µs) |
| disconn\_time | bigint  | not null | Disconnect timestamp (Unix‐epoch)      |
| identified    | boolean |          | Whether Identify succeeded             |
| error         | text    | not null | Dial error code or empty on success    |

**Indexes:**

* `conn_events_pkey` PRIMARY KEY on (`id`)

## active\_peers

Snapshots of active peer IDs over time.

| Column    | Type      | Nullable | Description                    |
| --------- | --------- | -------- | ------------------------------ |
| timestamp | timestamp | not null | Backup time                    |
| peers     | bigint\[] |          | Array of `peer_info.id` values |

**Indexes:**

* `active_peers_pkey` PRIMARY KEY on (`timestamp`)

## ips

Geolocation cache for IP addresses, refreshed periodically.

| Column           | Type      | Nullable | Description           |
| ---------------- | --------- | -------- | --------------------- |
| ip               | text      | not null | IP address (PK)       |
| expiration\_time | timestamp | not null | Cache TTL expiration  |
| continent        | text      | not null | Continent name        |
| continent\_code  | text      | not null | Continent ISO code    |
| country          | text      | not null | Country name          |
| country\_code    | text      | not null | Country ISO code      |
| region           | text      | not null | Region code           |
| region\_name     | text      | not null | Region name           |
| city             | text      | not null | City name             |
| zip              | text      | not null | Postal code           |
| lat              | real      | not null | Latitude              |
| lon              | real      | not null | Longitude             |
| isp              | text      | not null | ISP name              |
| org              | text      | not null | Organization name     |
| as\_raw          | text      | not null | Autonomous System     |
| asname           | text      | not null | AS name               |
| mobile           | boolean   | not null | Mobile network flag   |
| proxy            | boolean   | not null | Proxy flag            |
| hosting          | boolean   | not null | Hosting provider flag |

**Indexes:**

* `ips_pkey` PRIMARY KEY on (`ip`)
