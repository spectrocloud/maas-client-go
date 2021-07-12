package maasclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"github.com/spectrocloud/maas-client-go/maasclient/oauth1"
)

// Client
type Client struct {
	baseURL    string
	HTTPClient *http.Client
	apiKey     string
}

// NewClient creates new MaaS client with given API key
func NewClient(maasEndpoint string, apiKey string) *Client {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}
	httpClient := &http.Client{
		Transport: tr,
	}

	return &Client{
		apiKey:     apiKey,
		HTTPClient: httpClient,
		baseURL: fmt.Sprintf("%s/api/2.0", maasEndpoint),
	}
}

// send sends the request
// Content-type and body should be already added to req
func (c *Client) send(ctx context.Context, method string, apiPath string, params url.Values, v interface{}) error {
	var err error
	var req *http.Request

	if method == http.MethodGet {
		req, err = http.NewRequestWithContext(
			ctx,
			method,
			fmt.Sprintf("%s%s", c.baseURL, apiPath),
			nil,
		)
		if err != nil {
			return err
		}

		req.URL.RawQuery = params.Encode()
	} else {
		req, err = http.NewRequestWithContext(
			ctx,
			method,
			fmt.Sprintf("%s%s", c.baseURL, apiPath),
			strings.NewReader(params.Encode()),
		)
		if err != nil {
			return err
		}
	}

	return c.sendRequest(req, params, v)
}

func (c *Client) sendRequest(req *http.Request, params url.Values, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer closeResponseBody(res)

	// Try to unmarshall into errorResponse
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("unknown error, status code: %d, body: %s", res.StatusCode, string(bodyBytes))
	} else if res.StatusCode == http.StatusNoContent {
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func closeResponseBody(response *http.Response) {
	err := response.Body.Close()
	if err != nil {
		log.Printf("Unable to close response body")
	}
}

func authHeader(req *http.Request, queryParams url.Values, apiKey string) string {
	key := strings.SplitN(apiKey, ":", 3)

	if len(key) != 3 {
		return ""
	}
	auth := oauth1.NewOAuth(key[0], "", key[1], key[2])

	params := make(map[string]string)
	if req.Method != http.MethodPut {
		// for some bizarre-reason PUT doesn't need this
		for k, v := range queryParams {
			params[k] = v[0]
		}
	}

	return auth.BuildOAuthHeader(req.Method, req.URL.String(), params)
}
