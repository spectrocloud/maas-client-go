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
	"net/url"
)

const (
	SubnetsAPIPath = "/subnets/"
)

// SubnetIPAddress represents an IP address entry from a subnet's ip_addresses endpoint
type SubnetIPAddress struct {
	IP        string `json:"ip"`
	AllocType int    `json:"alloc_type"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
	User      string `json:"user"`
}

// Subnets interface for subnet operations
type Subnets interface {
	// List returns all subnets
	List(ctx context.Context) ([]Subnet, error)
	// GetIDByCIDR returns the subnet ID for the given CIDR, or an error if not found
	GetIDByCIDR(ctx context.Context, cidr string) (int, error)
	// GetIPAddresses returns all IP addresses tracked in the given subnet
	GetIPAddresses(ctx context.Context, subnetID int) ([]SubnetIPAddress, error)
	// IsIPInUse returns true if the given IP is tracked in the given subnet
	IsIPInUse(ctx context.Context, subnetID int, ip string) (bool, error)
}

// subnets controller implementation
type subnets struct {
	Controller
}

// NewSubnetsClient creates a new subnets client
func NewSubnetsClient(client *authenticatedClient) Subnets {
	return &subnets{
		Controller: Controller{
			client:  client,
			apiPath: SubnetsAPIPath,
			params:  ParamsBuilder(),
		},
	}
}

// List returns all subnets
func (s *subnets) List(ctx context.Context) ([]Subnet, error) {
	res, err := s.client.Get(ctx, s.apiPath, s.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*subnet
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return subnetSliceToInterface(obj), nil
}

// GetIDByCIDR returns the subnet ID matching the given CIDR
func (s *subnets) GetIDByCIDR(ctx context.Context, cidr string) (int, error) {
	subnets, err := s.List(ctx)
	if err != nil {
		return 0, err
	}

	for _, sn := range subnets {
		if sn.CIDR() == cidr {
			return sn.ID(), nil
		}
	}

	return 0, fmt.Errorf("no subnet found with CIDR %s", cidr)
}

// GetIPAddresses fetches all IP addresses tracked in the given subnet
func (s *subnets) GetIPAddresses(ctx context.Context, subnetID int) ([]SubnetIPAddress, error) {
	path := fmt.Sprintf("%s%d/", s.apiPath, subnetID)

	params := url.Values{}
	params.Set("op", "ip_addresses")
	params.Set("with_username", "1")
	params.Set("with_summary", "1")

	res, err := s.client.Get(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var ipAddresses []SubnetIPAddress
	err = unMarshalJson(res, &ipAddresses)
	if err != nil {
		return nil, err
	}

	return ipAddresses, nil
}

// IsIPInUse returns true if the given IP is tracked (in any alloc state) in the given subnet
func (s *subnets) IsIPInUse(ctx context.Context, subnetID int, ip string) (bool, error) {
	ipAddresses, err := s.GetIPAddresses(ctx, subnetID)
	if err != nil {
		return false, err
	}

	for _, addr := range ipAddresses {
		if addr.IP == ip {
			return true, nil
		}
	}

	return false, nil
}

