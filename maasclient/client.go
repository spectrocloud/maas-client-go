/*
Copyright 2021 Spectrocloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package maasclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/spectrocloud/maas-client-go/maasclient/oauth1"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// authenticatedClient
type authenticatedClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

type HTTPResponse struct {
	body   []byte
	status int
}

type Client interface {
	Get(ctx context.Context, path string, params url.Values) (HTTPResponse, error)
	PostForm(ctx context.Context, path string, contentType string, params url.Values, body io.Reader) (HTTPResponse, error)
	Post(ctx context.Context, path string, params url.Values) (HTTPResponse, error)
	Put(ctx context.Context, path string, params url.Values, body io.Reader, contentLength int) (HTTPResponse, error)
	PutParams(ctx context.Context, param string, params url.Values) (HTTPResponse, error)
	Delete(ctx context.Context, path string, params url.Values) (HTTPResponse, error)
}

func (c *authenticatedClient) Get(ctx context.Context, path string, params url.Values) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s%s", c.baseURL, path),
		nil,
	)
	if err != nil {
		return result, err
	}
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) Post(ctx context.Context, path string, params url.Values) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return result, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) PostForm(ctx context.Context, path string, contentType string, params url.Values, body io.Reader) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.baseURL, path),
		body,
	)
	if err != nil {
		return result, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", contentType)

	// for post requests longer than 300 seconds
	ticker := time.NewTicker(2 * time.Minute)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("refresing auth token", time.Unix(t.Unix(), 0).Format(time.RFC3339))
				authHeader := authHeader(req, params, c.apiKey)
				req.Header.Set("Authorization", authHeader)
			}
		}
	}()

	defer func() {
		ticker.Stop()
		done <- true
	}()

	authHeader := authHeader(req, params, c.apiKey)
	req.Header.Set("Authorization", authHeader)

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) Put(ctx context.Context, path string, params url.Values, body io.Reader, contentLength int) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s%s", c.baseURL, path),
		body,
	)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))
	req.ContentLength = int64(contentLength)

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) PutParams(ctx context.Context, path string, params url.Values) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) Delete(ctx context.Context, path string, params url.Values) (HTTPResponse, error) {
	result := HTTPResponse{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return result, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req, &result)
}

func (c *authenticatedClient) dispatchRequest(req *http.Request, result *HTTPResponse) (HTTPResponse, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return *result, err
	}

	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return *result, err
	}

	result.body = bodyBytes
	result.status = res.StatusCode

	return *result, nil
}

type authenticatedClientSet struct {
	client                 *authenticatedClient
	rackControllers        RackControllers
	dnsResouceController   DNSResources
	userController         Users
	zoneController         Zones
	bootController         BootResources
	domainController       Domains
	resourcePoolController ResourcePools
	spaceController        Spaces
	machineController      Machines
}

func (m *authenticatedClientSet) RackControllers() RackControllers {
	return m.rackControllers
}

func (m *authenticatedClientSet) DNSResources() DNSResources {
	return m.dnsResouceController
}

func (m *authenticatedClientSet) Users() Users {
	return m.userController
}

func (m *authenticatedClientSet) Zones() Zones {
	return m.zoneController
}

func (m *authenticatedClientSet) BootResources() BootResources {
	return m.bootController
}

func (m *authenticatedClientSet) Domains() Domains {
	return m.domainController
}

func (m *authenticatedClientSet) ResourcePools() ResourcePools {
	return m.resourcePoolController
}

func (m *authenticatedClientSet) Spaces() Spaces {
	return m.spaceController
}

func (m *authenticatedClientSet) Machines() Machines {
	return m.machineController
}

func NewAuthenticatedClientSet(maasEndpoint, apiKey string, options ...func(client *authenticatedClientSet)) ClientSetInterface {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}

	httpClient := &http.Client{Transport: transport}

	client := &authenticatedClient{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("%s/api/2.0", maasEndpoint),
	}

	clientSet := &authenticatedClientSet{
		client: client,
	}

	for _, option := range options {
		option(clientSet)
	}

	clientSet.rackControllers = NewRackControllersClient(client)
	clientSet.dnsResouceController = NewDNSResourcesClient(client)
	clientSet.userController = NewUsersClient(client)
	clientSet.zoneController = NewZonesClient(client)
	clientSet.bootController = NewBootResourcesClient(client)
	clientSet.domainController = NewDomainsClient(client)
	clientSet.resourcePoolController = NewResourcePoolsClient(client)
	clientSet.spaceController = NewSpacesClient(client)
	clientSet.machineController = NewMachinesClient(client)

	return clientSet
}

func (m *authenticatedClientSet) WithHTTPClient(client *http.Client) ClientSetInterface {
	m.client.httpClient = client
	return m
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
