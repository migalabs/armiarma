package visualizer

import (
    "fmt"
    "net/http"
)

type PeersPage struct {
    ChainVisualizer *ChainVisualizer
    Title       string
    ReqPath     string
    HtmlFolder  string
}

// Generate a new PeersPage
func NewPeersPage(cv *ChainVisualizer, reqpath string, title string) (*PeersPage, error){
    if cv == nil {
        return nil, fmt.Errorf("No ChainVisualizer has been given. Empty Pointer")
    }
    pp := &PeersPage {
        ChainVisualizer: cv,
        Title: title,
        ReqPath: reqpath,
        HtmlFolder: cv.HtmlFolder,
    }
    fmt.Println("DEBUG: New PeersPage", pp)
    return pp, nil
}

// Request Handler function for the new page
func (c *PeersPage) HandlerFunction(w http.ResponseWriter, r http.Request){
    fmt.Fprintln(w, "Peer Page. Incoming")
}


