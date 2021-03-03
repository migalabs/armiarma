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
}

// interface that all the pages will need to serve 
type VisualPage interface {
    HandlerFuntion(w http.ResponseWriter, r *http.Request)
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
    }
    return cv
}

// Intitilize / Set up the host
func (c *ChainVisualizer) InitializeHost() error {
    fmt.Println("Init the host")
    fmt.Println("Len of the Pages:", len(c.Pages), "Pages", c.Pages)
    for _, vp := range c.Pages {
        fmt.Println("Initializing New Page")
        fmt.Println(vp)
        fmt.Println("Requester Path:", vp.RequestPath())
        http.HandleFunc(vp.RequestPath(), vp.HandlerFuntion)
    }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "The Server is Working\n")
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
    fmt.Println("Addind New Page to the ChainVisualizer")
    fmt.Println(vp.Info())
    c.Pages = append(c.Pages, vp)
    fmt.Println("Remaining Pages on the Visualizer host:", c.Pages)
    return nil
}
