package visualizer

import (
    "fmt"
    "strconv"
    "github.com/olekukonko/tablewriter"
    "net/http"
    "github.com/protolambda/zrnt/eth2/beacon"
)

type BeaconStatePage struct {
    ChainVisualizer *ChainVisualizer
    ReqPath     string
    Title       string
    HtmlFolder  string
}

// Generate a new BeaconStatePage
func NewBeaconStatePage(cv *ChainVisualizer, reqpath string, title string) (*BeaconStatePage, error){
    if cv == nil {
        return nil, fmt.Errorf("No ChainVisualizer has been given, EmShort Summary of the last Beacon States of the BeaconChainpty pointer")
    }
    bsp := &BeaconStatePage {
        ChainVisualizer: cv,
        ReqPath:    reqpath,
        Title:      title,
        HtmlFolder: cv.HtmlFolder,
    }
    fmt.Println("DEBUG: New BeaconStatePage", bsp)
    return bsp, nil
}

// Request Handler Function for the BeaconStatePage
func (c *BeaconStatePage) HandlerFunction(w http.ResponseWriter, r *http.Request) {
    // Get the Length of the BeaconBlock Buffer
    buffLen := c.ChainVisualizer.BlocksBuffer.Len
    // Set the header of the block page
    typeStr     := "Type"
    slotStr     := "Slot"
    justifEpoch := "Justified Epoch"
    justiBRoot  := "Justified Epoch Root"
    finalizEpoch:= "Finalized Epoch"
    finalizBRoot:= "Finalized Epoch Root"

    header := "Latest " + strconv.Itoa(buffLen) + " Beacon States From the Mainnet Chain"
    fmt.Fprintln(w, header)
    fmt.Fprintln(w, "")
    tw := tablewriter.NewWriter(w)
    tableHeader := []string{ typeStr, slotStr, justifEpoch, justiBRoot, finalizEpoch, finalizBRoot}
    tw.SetHeader(tableHeader)
    for _, block := range c.ChainVisualizer.StatesBuffer.Items {
        if block == nil {
            break
        }
        bs := block.(*beacon.BeaconState)
        stateMsg := []string{"State", bs.Slot.String(), bs.CurrentJustifiedCheckpoint.Epoch.String(), bs.CurrentJustifiedCheckpoint.Root.String(),
                            bs.FinalizedCheckpoint.Epoch.String(), bs.FinalizedCheckpoint.Root.String()}
        tw.Append(stateMsg)
    }
    tw.SetBorder(false)
    tw.Render()
}

//
func (c *BeaconStatePage) Info() string {
    return "Short Summary of the last Beacon States of the BeaconChain"
}

// 
func (c *BeaconStatePage) RequestPath() string {
    return c.ReqPath
}
