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
	"fmt"
)

const (
	IPAddressesAPIPath = "/ipaddresses/"
)

// IPAddresses interface for managing IP addresses
type IPAddresses interface {
	List(ctx context.Context, params Params) ([]IPAddress, error)
	Get(ctx context.Context, ip string) (IPAddress, error)
	//GetAll retrieves a specific IP address regardless of ownership. This is only available to admins.
	GetAll(ctx context.Context, ip string) (IPAddress, error)
	Release(ctx context.Context, ip string) error
	ForceRelease(ctx context.Context, ip string) error
}

// IPAddresses controller implementation
type ipAddresses struct {
	Controller
}

// NewIPAddressesClient creates a new IP addresses client
func NewIPAddressesClient(client *authenticatedClient) IPAddresses {
	return &ipAddresses{
		Controller: Controller{
			client:  client,
			apiPath: IPAddressesAPIPath,
			params:  ParamsBuilder(),
		},
	}
}

// List retrieves all IP addresses
func (ips *ipAddresses) List(ctx context.Context, params Params) ([]IPAddress, error) {
	if params != nil {
		ips.params.Copy(params)
	}

	res, err := ips.client.Get(ctx, ips.apiPath, ips.params.Values())
	if err != nil {
		return nil, err
	}

	var ipAddresses []*ipaddress
	err = unMarshalJson(res, &ipAddresses)
	if err != nil {
		return nil, err
	}

	return ipStructSliceToInterface(ipAddresses, ips.client), nil
}

// GetAll retrieves a specific IP address even if it doesn't belong to the current user (admin only)
func (ips *ipAddresses) GetAll(ctx context.Context, ip string) (IPAddress, error) {
	ips.params.Reset()
	ips.params.Set("ip", ip)
	ips.params.Set("all", "true")

	res, err := ips.client.Get(ctx, ips.apiPath, ips.params.Values())
	if err != nil {
		return nil, err
	}

	var ipAddresses []*ipaddress
	err = unMarshalJson(res, &ipAddresses)
	if err != nil {
		return nil, err
	}

	if len(ipAddresses) == 0 {
		return nil, fmt.Errorf("no IP address found for %s", ip)
	}

	return ipStructToInterface(ipAddresses[0], ips.client), nil
}

// Get retrieves a specific IP address by IP string
func (ips *ipAddresses) Get(ctx context.Context, ip string) (IPAddress, error) {
	ips.params.Reset()
	ips.params.Set("ip", ip)

	res, err := ips.client.Get(ctx, ips.apiPath, ips.params.Values())
	if err != nil {
		return nil, err
	}

	var ipAddresses []*ipaddress
	err = unMarshalJson(res, &ipAddresses)
	if err != nil {
		return nil, err
	}

	if len(ipAddresses) == 0 {
		return nil, fmt.Errorf("no IP address found for %s", ip)
	}

	return ipStructToInterface(ipAddresses[0], ips.client), nil
}

// Release releases an IP address by IP string
func (ips *ipAddresses) Release(ctx context.Context, ip string) error {
	ips.params.Reset()
	ips.params.Set("op", "release")
	ips.params.Set("ip", ip)

	_, err := ips.client.Post(ctx, ips.apiPath, ips.params.Values())
	return err
}

// ForceRelease forcefully releases an IP address by IP string
func (ips *ipAddresses) ForceRelease(ctx context.Context, ip string) error {
	ips.params.Reset()
	ips.params.Set("op", "release")
	ips.params.Set("ip", ip)
	ips.params.Set("force", "true")

	_, err := ips.client.Post(ctx, ips.apiPath, ips.params.Values())
	return err
}
