package visualizer

import (
    "fmt"
    "strconv"
    "net/http"
    bdb "github.com/protolambda/rumor/chain/db/blocks"
)

// BeaconBlock Page Struct 
type BeaconBlockPage struct {
    ChainVisualizer *ChainVisualizer
    Title       string
    ReqPath string
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
    }
    fmt.Println("DEBUG: New BeaconBlockPage", bbp)
    return bbp, nil
}

// Request Handler function for the new page
func (c *BeaconBlockPage) HandlerFuntion(w http.ResponseWriter, r *http.Request){
    fmt.Println(c.ChainVisualizer)
    // Get the Length of the BeaconBlock Buffer
    buffLen := c.ChainVisualizer.BlocksBuffer.Len
    header := "Latest " + strconv.Itoa(buffLen) + " Beacon Blocks From the Mainnet Chain\n"
    fmt.Fprintf(w, header)
    fmt.Fprintf(w, " -- Type -- | --------------------------------- Root --------------------------- | --- Slot --- |\n")
    for _, block := range c.ChainVisualizer.BlocksBuffer.Items {
        if block == nil {
            break
        }
        bwr := block.(*bdb.BlockWithRoot)
        blockMsg := "->   Block    " + bwr.Root.String() + "  " + bwr.Block.Message.Slot.String() + "\n"
        fmt.Fprintf(w, blockMsg)
    }
    fmt.Fprintf(w, "\n")
    fmt.Fprintf(w, "----  Done ----\n")
}

// Return the info regarding the BeaconBlockPage
func (c BeaconBlockPage) Info() string {
    return "Short Summary of the last Blocks Viewed in the Chain (Latest BeaconBlocks)"
}

// Return the RequestPath for the BeaconBlockPage
func (c BeaconBlockPage) RequestPath() string{
    return c.ReqPath
}

