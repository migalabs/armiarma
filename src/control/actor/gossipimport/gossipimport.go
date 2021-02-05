package gossipimport

import (
    "io/ioutil"
    "context"
    "encoding/json"
    "fmt"
    "bytes"
    "net/http"
    "github.com/protolambda/ztyp/codec"
    "github.com/protolambda/rumor/p2p/gossip/database"
    "github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/metrics"
    "github.com/protolambda/zrnt/eth2/beacon"
    pgossip "github.com/protolambda/rumor/p2p/gossip"
    "github.com/protolambda/rumor/control/actor/blocks"
    "github.com/protolambda/rumor/control/actor/states"
    bdb "github.com/protolambda/rumor/chain/db/blocks"
	sdb "github.com/protolambda/rumor/chain/db/states"

)

type GossipImportCmd struct {
    *base.Base

    GossipMetrics   *metrics.GossipMetrics
    BlocksDB         bdb.DBs
    BlockDBState    *blocks.DBState
    StatesDB        sdb.DBs
    StateDBState    *states.DBState

    BName       bdb.DBID    `ask:"<bname>"  help:"Name of the network that will be used for the Blocks and State DBs"`
    SName       sdb.DBID    `ask:"<sname>"  help:"Name of the network that will be used for the Blocks and State DBs"`
    ForkDigest  string  `ask:"--fork-digest" help:"Fork digest of the given network"`
    Path        string  `ask:"[path]"        help:"The path used for the DB. It will be a memory DB if left empty."`

    IP          string  `ask:"--ip"      help:"IP where the local beacon node can be reacheable.(default "localhost")"`
    Port        string  `ask:"--port"    help:"Port where the local beacon node offers the API to request the Beacon State.(Default 3500, endpoint of Prysm Beacon Nodes)"`
}


func (c *GossipImportCmd) Help() string {
    return "Import Beacon Blocks and States from the gossip messages (requires, gossip BeaconBlock topic logged and both States and Blocks will be initialized)"
}

func (c *GossipImportCmd) Default() {
    c.SName       = "mainnet"
    c.BName       = "mainnet"
    c.ForkDigest    = "b5303f2a"
    c.IP            = "localhost"
    c.Port          = "3500"
}

func (c *GossipImportCmd) Run( ctx context.Context, args ...string) error {
    // Define Harcoded Variables for the mainnet gossip topic
    encoding    := "ssz_snappy"
    eth2Topic   := "beacon_block"
    // Compose the full topic for the Eth2 network
    topicName := pgossip.GenerateEth2Topics(c.ForkDigest, eth2Topic, encoding)
    // Check if the beacon_block topic has its messageDB by checking if it has a notification channel
    if _, ok := c.GossipMetrics.TopicDatabase.NotChan[topicName]; ok {
        // Generate the DBIDs for the block and state DB
        // If there is a notification channel, we generate a DB for the Blocks and for the States
        blockDB, err := c.BlocksDB.Create(c.BName, c.Path, c.GossipMetrics.TopicDatabase.Spec)
        if err != nil {
            return err
        }
        c.BlockDBState.CurrentDB = c.BName
        stateDB, err := c.StatesDB.Create(c.SName, c.Path, c.GossipMetrics.TopicDatabase.Spec)
        if err != nil {
            return err
        }
        c.StateDBState.CurrentDB = c.SName
        reqUrl := ComposePrysmBSRequest(c.IP, c.Port, PrysmBSQuery)
        // Generate the GO routine that will keep in the backgournd importing the Blocks and the States from the received BeaconBlocks
        stopping := make(chan bool, 2)
        go func() {
            c.Log.Infof("Gossip Import has been launched. Beacon Blocks and Beacons States will be imported from the Gossip Messages on:", topicName)
            for {
                select {
                case newBlock := <-c.GossipMetrics.TopicDatabase.NotChan[topicName]:
                    if newBlock == true {
                        fmt.Println("New block has been received on the MessageDB")
                        // read the Received BeaconBlock from the MessageDB
                        bblockMsg, err := c.GossipMetrics.TopicDatabase.ReadMessage(topicName)
                        if err != nil {
                            fmt.Println("Error Reading the message from the messageDB")
                        } else {
                            var state *beacon.BeaconStateView
                            bblock := bblockMsg.(*database.ReceivedBeaconBlock)
                            sRoot := bblock.SignedBeaconBlock.Message.StateRoot
                            fmt.Println("Readed Block on slot:", bblock.SignedBeaconBlock.Message.Slot, "on Beacon State:", sRoot)
                            // Rquest the stire BeaconState from the associated Beacon Node
			    slotString := fmt.Sprint(bblock.SignedBeaconBlock.Message.Slot)
                            rawbstate, err := RequestBeaconState(reqUrl, slotString)
                            if err != nil {
                                fmt.Errorf("No Beacon State was received")
                                continue
                            }

                            stateSize := rawbstate.ByteLength(c.GossipMetrics.TopicDatabase.Spec)
			    fmt.Println("Size of the received State:", stateSize)
			    bbytes, err := json.Marshal(rawbstate)
			    if err != nil {
				fmt.Println("Error Marshaling again the BeaconState")
			    }
			    bsreader := bytes.NewReader(bbytes)
			    // With the spec.BeaconState() and the raw/plain BeaconState
                            // Import the received BeaconState from the Local Beacon Node
                            state, err = beacon.AsBeaconStateView(c.GossipMetrics.TopicDatabase.Spec.BeaconState().Deserialize(codec.NewDecodingReader(bsreader, uint64(stateSize))))
                            if err != nil {
				fmt.Println("Error generating the BeaconStateView", err)
                                fmt.Errorf("%s",err)
                                continue
                            } else {
			    	fmt.Println("BeaconStateView acchieved: %+v", state)
			    }
			    fmt.Println("Storing the State on the DB")
                            _, err = stateDB.Store(ctx, state)
                            if err != nil {
                                fmt.Println("Error Storing the State on the DB", err)
                                continue
                            }
                            fmt.Println("State", sRoot.String(), "Has been Saved on the StatesDB")
                            // Import the BeaconBlock to the DB, (Maybe the Store thingy might be implemented) 
//                          var sbblock beacon.SignedBeaconBlock
                            bbuff, err := json.Marshal(bblock.SignedBeaconBlock)
                            if err != nil {
                                fmt.Errorf("Error Marshalling the Beacon Block while trying to Import it to the DB")
                                continue
                            }
//                          sbblock.Serialize(c.GossipMetrics.TopicDatabase.Spec, codec.NewEncodingWriter(bbuff))
                            _, err = blockDB.Import(bytes.NewReader(bbuff))
                            if err != nil {
                                fmt.Errorf("%s", err)
                                continue
                            }
                            fmt.Println("Beacon Block", bblock.SignedBeaconBlock.Message.Slot, "has been succesfully added")
                            // Should be sucessfully added
                            continue
                        }
                    }
                case stop := <-stopping:
                    if stop == true {
                        fmt.Println("The GossipImport has been stopped")
                        return
                    }
                }
            }
        }()
        c.Control.RegisterStop(func(ctx context.Context) error{
            stopping <- true
            c.Log.Infof("Stopped Importing From Gossip")
            return nil
        })
    } else {
        // there is no DB with the same topic name, and therefore no NotChan
        return fmt.Errorf("There is no Database for the given topic:",topicName)
    }
    return nil
}

// List of queries for the different clients that can be used to obtain the BeaconState
// EXAMPLE:  "http://localhost:3500/eth/v1alpha1/debug/state?=2832"
// 	      http://localhost:3500/eth/v1alpha1/debug/state?=468221
var PrysmBSQuery string = "/eth/v1alpha1/debug/state?="

func ComposePrysmBSRequest(ip string, port string, query string) string {
    composedUrl := "http://" + ip + ":" + port + PrysmBSQuery
    return composedUrl
}

func RequestBeaconState(url string, slot string) (beacon.BeaconState, error) {
    reqUrl := url + slot
    fmt.Println("Url:",reqUrl)
    resp, err := http.Get(reqUrl)
    if err != nil {
        fmt.Println("Error while getting the Beacon State from the local node, request:", reqUrl)
    }
    fmt.Println("response:", resp)
    defer resp.Body.Close()
    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    // Here I don't know if the BeaconState comes serialized or in a Json.
    var beaconState beacon.BeaconState
    if len(bodyBytes) == 0 {
        return beaconState, fmt.Errorf("Error Unmarshalling the response from the Local Beacon Node, response:", resp)
    }
    err = json.Unmarshal(bodyBytes, &beaconState)
    if err != nil {
	fmt.Println("Error unmarshalling the BeaconState received from the local node")
    }
    fmt.Println("BeaconState", slot , "successfully obtained from local node: %+v", beaconState)
    return beaconState, nil
}



