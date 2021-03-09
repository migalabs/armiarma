package chainvisualizer

import(
    "fmt"
    "context"
    "github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/control/actor/blocks"
    "github.com/protolambda/rumor/control/actor/states"
    "github.com/protolambda/rumor/visualizer"
)

type VisualizerState struct {
    ChainVisualizer *visualizer.ChainVisualizer
    Already bool
}

type VisualizerCmd struct {
    *base.Base
    *VisualizerState

    BlockDBState *blocks.DBState
    StateDBState *states.DBState

    Ip      string `ask:"--ip"   help:"Define the IP where the Visualizer will be hosted"`
    Port    string `ask:"--port" help:"Define the Port to listen and serve the visualizer host"`
    Len     int    `ask:"--len"  help:"Define the range of Blocks and States that will be offered on the Visualizer"`
}

func (c *VisualizerCmd) Default() {
    c.Ip    = "localhost"
    c.Port  = "9100"
    c.Len   = 10
}

func (c *VisualizerCmd) Help() string {
    return "Serve a Chain Status Visualizer as a Web App on the given Ip and Port"
}

func (c *VisualizerCmd) Run(ctx context.Context, args ...string) error{
    fmt.Println("Launching the Visualizer")
    // check if there is already another Visulizer
    if c.VisualizerState.Already != false {
        c.Log.Error("There is already a Chain Visualized hosted.")
        return fmt.Errorf("There is already a Chain Visualized hosted.")
    }
    fmt.Println("After checking the empty/full pointer")
    // Initialize the ChainVisualizer 
    cv := visualizer.NewChainVisualizer(c.Len, c.Ip, c.Port)
    // Generate the Pages that will be available from the Visualizer
    // --- Beacon Block ----
    bbp, err := visualizer.NewBeaconBlockPage(cv, "/blocks", "Beacon Blocks")
    if err != nil {
        c.Log.Error("Error Generating the Block Page on the Visualizer")
        return err
    }
    err = cv.AddNewPage(bbp)
    if err != nil {
        c.Log.Error("Error Adding the Block Page on the Visualizer")
        return err
    }
    // --- Beacon State ---
    bsp, err := visualizer.NewBeaconStatePage(cv, "/states", "Beacon States")
    if err != nil {
        c.Log.Error("Error Generating the State Page on the Visualizer")
        return err
    }
    err = cv.AddNewPage(bsp)
    if err != nil {
        c.Log.Error("Error Adding the State Page on the Visualizer")
        return err
    }
    // Add the Cain Visualizer to the VisualizerState
    c.VisualizerState.ChainVisualizer = cv
    // (Pages to show needs to be ready from before)
    // Generate the host in a nother Go routine
    // TODO: the end of the go routine needs to be defined (ctx?, channel to cancel?)
    fmt.Println("-------------------> INITIALIZING HOST -------------------")
    c.Log.Info("Initializing the Visualizer Host")
    err = cv.InitializeHost()
    if err != nil {
        c.Log.WithError(err).Error("Error Initializing the Host")
    }
    fmt.Println("------------------> HOST FINISHED -------------------")
//  go func() {
//      c.Log.Info("Initializing the Visualizer Host")
//      err := cv.InitializeHost()
//      if err != nil {
//          c.Log.WithError(err).Error("Error Initializing the Host")
//      }
//  }()
    c.VisualizerState.Already = true
    return nil
}
