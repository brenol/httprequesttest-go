package example

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Api struct {
	token      string
	httpclient *http.Client
}

type Option func(*Api)

// New builds a API client from the provided token and options.
func New(token string, opts ...Option) *Api {
	a := &Api{
		token:      token,
		httpclient: &http.Client{},
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// OptionHTTPClient - provide a custom http client to the client
func OptionHTTPClient(c *http.Client) func(*Api) {
	return func(a *Api) {
		a.httpclient = c
	}
}

type ResponseBody struct {
	Text string `json:"text"`
}

func (api *Api) Do() (*ResponseBody, error) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", api.token))

	resp, err := api.httpclient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var body ResponseBody
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, err
	}

	return &body, nil
}
