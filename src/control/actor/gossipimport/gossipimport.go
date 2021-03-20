package gossipimport

import (
	"context"
	"fmt"
	"os"
	"strconv"

	cnodes "github.com/cortze/go-eth2-beacon-nodes/nodes"
	"github.com/protolambda/rumor/chain"
	bdb "github.com/protolambda/rumor/chain/db/blocks"
	sdb "github.com/protolambda/rumor/chain/db/states"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/blocks"
	cchain "github.com/protolambda/rumor/control/actor/chain"
	visualizer "github.com/protolambda/rumor/control/actor/chainvisualizer"
	pstatus "github.com/protolambda/rumor/control/actor/peer/status"
	"github.com/protolambda/rumor/control/actor/states"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type GossipImportCmd struct {
	*base.Base

	GossipMetrics   *metrics.GossipMetrics
	VisualizerState *visualizer.VisualizerState
	*pstatus.PeerStatusState

	BlocksDB     bdb.DBs
	BlockDBState *blocks.DBState
	StatesDB     sdb.DBs
	StateDBState *states.DBState
	chain.Chains
	*cchain.ChainState

	BName       bdb.DBID          `ask:"<bname>"  help:"Name of the network that will be used for the Blocks and State DBs"`
	SName       sdb.DBID          `ask:"<sname>"  help:"Name of the network that will be used for the Blocks and State DBs"`
	CName       chain.ChainID     `ask:"<cname>" help:"Name of the network that will be used for the Chain that the Rumor Node will follow"`
	ForkDigest  beacon.ForkDigest `ask:"--fork-digest" help:"Fork digest of the given network"`
	ProjectPath string            `ask:"--project-path"  help:"Path to the folder where the BD will be created"`
	DBPath      string            `ask:"--db-name" help:"The path used for the DB. It will be a memory DB if left empty."`

	IP   string `ask:"--ip"      help:"IP where the local beacon node can be reacheable.(default "localhost")"`
	Port string `ask:"--port"    help:"Port where the local beacon node offers the API to request the Beacon State.(Default 3500, endpoint of Prysm Beacon Nodes)"`
}

func (c *GossipImportCmd) Help() string {
	return "Import Beacon Blocks and States from the gossip messages (requires, gossip BeaconBlock topic logged and both States and Blocks will be initialized)"
}

func (c *GossipImportCmd) Default() {
	c.IP = "localhost"
	c.Port = "3500"
}

func (c *GossipImportCmd) Run(ctx context.Context, args ...string) error {

	// Check if the Message DB has been initialized
	if c.GossipMetrics.MessageDatabase != nil {
		// change directory to the Project folder (maybe it will solve the error of Path)
		err := os.Chdir(c.ProjectPath)
		if err != nil {
			c.Log.WithError(err).Error("Error changing directory to", c.ProjectPath)
			return err
		}
		// Generate the DBIDs for the block and state DB
		// If there is a notification channel, we generate a DB for the Blocks and for the States
		blockDB, err := c.BlocksDB.Create(c.BName, c.DBPath, c.GossipMetrics.MessageDatabase.Spec)
		if err != nil {
			c.Log.WithError(err).Error("Error creating BlockDB.")
			return err
		}
		c.BlockDBState.CurrentDB = c.BName
		// Currently StatesDB doesn't accept any kind of disk datbase
		//        stateDB, err := c.StatesDB.Create(c.SName, "", c.GossipMetrics.MessageDatabase.Spec)
		_, err = c.StatesDB.Create(c.SName, "", c.GossipMetrics.MessageDatabase.Spec)
		if err != nil {
			return err
		}
		c.StateDBState.CurrentDB = c.SName
		// generate the client struct from where to gather the chain metrics
		prysmClient := cnodes.NewPrysmClient(c.IP, c.Port)
		// Generate the GO routine that will keep in the backgournd importing the Blocks and the States from the received BeaconBlocks
		stopping := make(chan bool, 2)
		go func() {
			c.Log.Infof("Gossip Import has been launched. Beacon Blocks and Beacons States will be imported from the Gossip Messages on: %s", c.DBPath)
			for {
				select {
				case newBlock := <-c.GossipMetrics.MessageDatabase.BlockNotChan:
					// Check if the incoming New Block is an empty
					if newBlock != nil {
						c.Log.WithError(err).Warn("Error Reading the message from the notification channel - empty struct ")
					} else {
						fmt.Println("New Beacon Block to the chain")
						//						  sRoot := bblock.SignedBeaconBlock.Message.StateRoot
						// Rquest the stire BeaconState from the associated Beacon Node
						slotNumber, err := strconv.Atoi(newBlock.Message.Slot.String())
						bState, err := prysmClient.GetFlatBeaconStateFromSlot(slotNumber)
						if err != nil {
							c.Log.WithError(err).Warn("Error getting the BeaconStateView from the Client")
							continue
						}
						// Store the BeaconBlock to the DB, (Maybe the Store thingy might be implemented)
						blockWRoot := bdb.WithRoot(c.GossipMetrics.MessageDatabase.Spec, newBlock)
						// we dont mind if the block already exists
						_, err = blockDB.Store(ctx, blockWRoot)
						if err != nil {
							c.Log.WithError(err).Warn("Error Storing the Block into the Database")
							continue
						}
						c.Log.Info("New BeaconBlock has been added to the BlocksDBb. Block Slot: ", newBlock.Message.Slot.String())
						// Save both, the BeaconBlock and the BeaconState into cache for ChainVisualizer purposes
						// ------- Visualizer Part -------
						// Check if the Visualizer has been Set
						if c.VisualizerState.ChainVisualizer != nil {
							// If the Visualizer has been initialized, we have to add the states and the blocks to the Buffers
							err = c.VisualizerState.ChainVisualizer.BlocksBuffer.AddItem(blockWRoot)
							if err != nil {
								c.Log.WithError(err).Warn("Error Importing the Block to the Visualizer Buffer")
							}
							err = c.VisualizerState.ChainVisualizer.StatesBuffer.AddItem(bState)
							if err != nil {
								c.Log.WithError(err).Warn("Error Importing the State to the Visualizer Buffer")
							}
						}

						// CHAIN secuence so that will allow the Rumor Node to follow/store/interacte with the data of the Block/State chains
						currentChain, ok := c.Chains.Find(c.CName)
						if !ok { // If not OK we will have to generate a new Chain
							fmt.Println("There was no Chain available")
							bStateView, err := prysmClient.GetBeaconStateViewFromSlot(slotNumber)
							if err != nil {
								c.Log.WithError(err).Warn("Error getting the BeaconStateView from the Client")
								continue
							}
							//							_, err = stateDB.Store(ctx, bStateView)
							//							if err != nil {
							//								c.Log.WithError(err).Warn("Error Storing the State on the DB", err)
							//								continue
							//							}
							// generate a new hot entry to create a chain
							hotEntry, err := GenerateNewHotEntry(bStateView, c.GossipMetrics.MessageDatabase.Spec)
							_ = currentChain
							fmt.Println("Creating New Chain")
							_, err = c.Chains.Create(c.CName, hotEntry, c.GossipMetrics.MessageDatabase.Spec)
							if err != nil {
								c.Log.WithError(err).Warn("Error Creating a new chain")
								continue

							}
							c.ChainState.CurrentChain = c.CName
						} else { // If OK, just add the block to the chain
							err = currentChain.AddBlock(ctx, newBlock)
							if err != nil {
								c.Log.WithError(err).Warn("Error while including the SignedBeaconBlock to the Chain")
								continue
							}
						}
						// UPDATE the Rumor Node status (at least fake that we follow the chain)
						st, err := NodeStatusFromBeaconState(bState, c.ForkDigest)
						if err != nil {
							c.Log.WithError(err).Warn("Error generating the new Status from the BeaconState")
							continue
						}
						c.PeerStatusState.Local = *st
						continue
					}
				case stop := <-stopping:
					if stop == true {
						c.Log.Info("The GossipImport has been stopped")
						break
					}
					// TODO: Propably will be needed to import the attestations received from gossip messages
					// case newAttestation:
				}
			}
		}()
		c.Control.RegisterStop(func(ctx context.Context) error {
			stopping <- true
			c.Log.Infof("Stopped Importing From Gossip")
			return nil
		})
	} else {
		// there is no DB with the same topic name, and therefore no NotChan
		return fmt.Errorf("There is no Message Database Initialized")
	}
	return nil
}
