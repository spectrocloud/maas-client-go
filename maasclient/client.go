/*
Copyright 2021 Spectro Cloud

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
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spectrocloud/maas-client-go/maasclient/oauth1"
)

// authenticatedClient
type authenticatedClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

type Client interface {
	Get(ctx context.Context, path string, params url.Values) (*http.Response, error)
	PostForm(ctx context.Context, path string, contentType string, params url.Values, body io.Reader) (*http.Response, error)
	Post(ctx context.Context, path string, params url.Values) (*http.Response, error)
	Put(ctx context.Context, path string, params url.Values, body io.Reader, contentLength int) (*http.Response, error)
	PutParams(ctx context.Context, param string, params url.Values) (*http.Response, error)
	Delete(ctx context.Context, path string, params url.Values) (*http.Response, error)
}

func (c *authenticatedClient) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s%s", c.baseURL, path),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) Post(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) PostForm(ctx context.Context, path string, contentType string, params url.Values, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.baseURL, path),
		body,
	)
	if err != nil {
		return nil, err
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

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) Put(ctx context.Context, path string, params url.Values, body io.Reader, contentLength int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s%s", c.baseURL, path),
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))
	req.ContentLength = int64(contentLength)

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) PutParams(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) Delete(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s%s", c.baseURL, path),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authHeader(req, params, c.apiKey))

	return c.dispatchRequest(req)
}

func (c *authenticatedClient) dispatchRequest(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

type authenticatedClientSet struct {
	client                      *authenticatedClient
	rackControllers             RackControllers
	dnsResouceController        DNSResources
	userController              Users
	zoneController              Zones
	sshKeyController            SSHKeys
	bootController              BootResources
	domainController            Domains
	tagController               Tags
	resourcePoolController      ResourcePools
	spaceController             Spaces
	machineController           Machines
	networkInterfacesController NetworkInterfaces
	ipAddressesController       IPAddresses
	vmHostsController           VMHosts
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

func (m *authenticatedClientSet) SSHKeys() SSHKeys {
	return m.sshKeyController
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

func (m *authenticatedClientSet) Tags() Tags {
	return m.tagController
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

func (m *authenticatedClientSet) NetworkInterfaces() NetworkInterfaces {
	return m.networkInterfacesController
}

func (m *authenticatedClientSet) IPAddresses() IPAddresses {
	return m.ipAddressesController
}

func (m *authenticatedClientSet) VMHosts() VMHosts {
	return m.vmHostsController
}

func NewAuthenticatedClientSet(maasEndpoint, apiKey string, options ...func(client *authenticatedClientSet)) ClientSetInterface {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 : already addressed in PCP-3389
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
	clientSet.sshKeyController = NewSSHKeysClient(client)
	clientSet.zoneController = NewZonesClient(client)
	clientSet.bootController = NewBootResourcesClient(client)
	clientSet.tagController = NewTagsClient(client)
	clientSet.domainController = NewDomainsClient(client)
	clientSet.resourcePoolController = NewResourcePoolsClient(client)
	clientSet.spaceController = NewSpacesClient(client)
	clientSet.machineController = NewMachinesClient(client)
	clientSet.networkInterfacesController = NewNetworkInterfacesClient(client)
	clientSet.ipAddressesController = NewIPAddressesClient(client)
	clientSet.vmHostsController = NewVMHostsClient(client)

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
