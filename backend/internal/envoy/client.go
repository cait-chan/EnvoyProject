package envoy

import (
	"io"		// read response body data
	"net/http"  // make HTTP requests
)


type Client struct {
	adminURL string
	httpClient *http.Client 
}

func NewClient(adminURL string) *Client {		// creates new Client w/ given URL and default HTTP client
	return &Client{				
		adminURL: adminURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) GetStats() ([]byte, error) {	// (c *Client) means this function belongs to the Client
	// build URL
	url := c.adminURL + "/stats?format=json"

	// make the GET request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	// close response body when done
	defer resp.Body.Close()

	// read all data from response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// return data
	return body, nil
}

func (c *Client) GetClusters() ([]byte, error) {	// (c *Client) means this function belongs to the Client
	// build URL
	url := c.adminURL + "/clusters?format=json"

	// make the GET request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	// close response body when done
	defer resp.Body.Close()

	// read all data from response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// return data
	return body, nil
}