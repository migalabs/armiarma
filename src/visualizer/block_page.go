package visualizer

import (
    "fmt"
    "github.com/olekukonko/tablewriter"
    "strconv"
//    "html/template"
    "net/http"
    bdb "github.com/protolambda/rumor/chain/db/blocks"
)

// BeaconBlock Page Struct 
type BeaconBlockPage struct {
    ChainVisualizer *ChainVisualizer
    ReqPath     string
    Title       string
    HtmlFolder  string
}

// Generate a new BeaconBlockPage
func NewBeaconBlockPage(cv *ChainVisualizer, reqpath string, title string) (*BeaconBlockPage, error){
    if cv == nil {
        return nil, fmt.Errorf("No ChainVisualizer has been given, Empty pointer")
    }
    bbp := &BeaconBlockPage{
        ChainVisualizer: cv,
        ReqPath: reqpath,
        Title:   title,
        HtmlFolder: cv.HtmlFolder,
    }
    return bbp, nil
}


// Request Handler function for the new page
func (c *BeaconBlockPage) HandlerFunction(w http.ResponseWriter, r *http.Request){
    // Get the Length of the BeaconBlock Buffer
    buffLen := c.ChainVisualizer.BlocksBuffer.Len
    // Set the header of the block page
    typeStr     := "Type"
    rootStr     := "Root"
    slotStr     := "Slot"
    parentStr   := "Parent Root"
    stateStr    := "State Root"
    proposerStr := "Proposer ID"

    // Autorefresh the webpage every 12 seconds (everytime I get a new block)
//  var metadata string = "<meta http-equiv=\"refresh\" content=\"3\" />"
//  fmt.Fprintln(w, metadata)

    header := "Latest " + strconv.Itoa(buffLen) + " Beacon Blocks From the Mainnet Chain"
    fmt.Fprintln(w, header)
    fmt.Fprintln(w, "")
    tw := tablewriter.NewWriter(w)
    tableHeader := []string{ typeStr, slotStr, rootStr, proposerStr, stateStr,parentStr }
    tw.SetHeader(tableHeader)
    for _, block := range c.ChainVisualizer.BlocksBuffer.Items {
        if block == nil {
            break
        }
        bwr := block.(*bdb.BlockWithRoot)
        blockMsg := []string{"Block", bwr.Block.Message.Slot.String(), bwr.Root.String(), bwr.Block.Message.ProposerIndex.String(), bwr.Block.Message.StateRoot.String(), bwr.Block.Message.ParentRoot.String() }
        tw.Append(blockMsg)
    }
    tw.SetBorder(false)
    tw.Render()
}

// Return the info regarding the BeaconBlockPage
func (c *BeaconBlockPage) Info() string {
    return "Short Summary of the last Blocks Viewed in the Chain (Latest BeaconBlocks)"
}

// Return the RequestPath for the BeaconBlockPage
func (c *BeaconBlockPage) RequestPath() string{
    return c.ReqPath
}

