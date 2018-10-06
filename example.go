package example

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Api struct {
	formattedToken string
	httpclient     *http.Client
}

type Option func(*Api)

// New builds a API client from the provided token and options.
func New(token string, opts ...Option) *Api {
	a := &Api{
		formattedToken: fmt.Sprintf("Bearer %s", token),
		httpclient:     &http.Client{},
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

	req.Header.Add("Authorization", api.formattedToken)

	resp, err := api.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // always discard body if it won't be read. (i.e. statusCode != 200)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}

	var body ResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}
