package endpoint

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// ***** TODO: Move all this code into an Infura - Eth2 golang SDK *****

var ()

type InfuraClient struct {
	endpoint string
	timeout  time.Duration
}

// *** Set Infura Clinet ***

func NewInfuraClient(infuraEndpoint string) (InfuraClient, error) {
	infuraCli := InfuraClient{}
	infuraCli.Default()
	// check if there is any other endpoint asigned
	if infuraEndpoint == "" {
		return infuraCli, errors.New("empty infura endpoint given")
	}
	infuraCli.endpoint = infuraEndpoint
	return infuraCli, nil

}

func (c *InfuraClient) Default() {
	c.timeout = 10 * time.Second
}

func (c *InfuraClient) IsInitialized() bool {
	return c.endpoint != ""
}

type ResponseData struct {
	Content interface{} `json:"data"`
}

// *** Make HTTPS requests to the endpoint ***
func (c *InfuraClient) NewHttpsRequest(ctx context.Context, request string, output interface{}) error {
	// generate a new context out of the given one (with timeout)
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	url := c.endpoint + request
	// make the API call to the endpoint
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to generate timeout request fo infura endpoint")
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to get API request from Infura endpoint")
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}
	response := ResponseData{
		Content: output,
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal received resp into Data")
	}
	return nil
}
