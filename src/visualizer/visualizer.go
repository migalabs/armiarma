package visualizer

import (
    "fmt"
    "net/http"
)

type ChainVisualizer struct {
    BlocksBuffer    Buffer
    StatesBuffer    Buffer
    Pages           []VisualPage

    Ip              string
    Port            string

    HtmlFolder      string
}

// interface that all the pages will need to serve 
type VisualPage interface {
    HandlerFunction(w http.ResponseWriter, r *http.Request)
    RequestPath() string
    Info() string
}

// Generate a new ChainVisualizer
func NewChainVisualizer(length int, ip string, port string) *ChainVisualizer{
    cv := &ChainVisualizer {
        BlocksBuffer: *NewBuffer(length),
        StatesBuffer: *NewBuffer(length),
        Pages:  make([]VisualPage, 0),
        Ip:     ip,
        Port:   port,
        HtmlFolder:    "htmls/",
    }
    return cv
}

// Intitilize / Set up the host
func (c *ChainVisualizer) InitializeHost() error {
    for _, vp := range c.Pages {
        http.HandleFunc(vp.RequestPath(), vp.HandlerFunction)
    }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "Available Pages\n")
        for _, vp := range c.Pages{
            req := vp.RequestPath() + "\n"
            fmt.Fprintf(w, req)
        }
    })
    // After the Page Initialization, set the Host listening at the given port
    // NOTE:  Port has to include the ':' before the port -> e.g. ":9100"
    port := ":" + c.Port
    fmt.Println("Initializing the Host at port ->", port)
    err := http.ListenAndServe(port, nil)
    if err != nil {
        return err
    }
    return nil
}

// Add new Page to the 
func (c *ChainVisualizer)AddNewPage(vp VisualPage) error {
    if vp == nil {
        return fmt.Errorf("Received new VisualPage was empty")
    }
    c.Pages = append(c.Pages, vp)
    return nil
}
