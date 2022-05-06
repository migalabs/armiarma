package info

var (
	// Crawler
	DefaultIP        string = "0.0.0.0"
	DefaultTcpPort   int    = 9020
	DefaultUdpPort   int    = 9020
	DefaultUserAgent string = "bsc_crawler"
	DefaultLogLevel  string = "info"

	// Bootnodes
	DefaultEth2BootNodesFile string = "./bootstrap_nodes/bootnodes_mainnet.json"
	DefaultIpfsBootNodesFile string = "./bootstrap_nodes/ipfs-bootnodes.json"

	// Config Files
	DefaultEth2ConfigFile string = "./config-files/eth2-config.json"
	DefaultIpfsConfigFile string = "./config-files/ipfs-config.json"

	// ETH2
	DefaultEth2Network string = "mainnet"

	// IPFS
	DefaultIpfsNetwork string = "ipfs"
	Ipfsprotocols             = []string{
		"/ipfs/kad/1.0.0",
		"/ipfs/kad/2.0.0",
	}
	Filecoinprotocols = []string{
		"/fil/kad/testnetnet/kad/1.0.0",
	}

	// Metrics
	DefaultOutputPath string = "./peerstore"

	// Control
	MinPort           int      = 0
	MaxPort           int      = 65000
	PossibleLogLevels []string = []string{"trace", "debug", "info", "warn", "error"}
)
