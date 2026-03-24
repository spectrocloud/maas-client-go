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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
)

const (
	IPRangesAPIPath = "/ipranges/"
)

// IPRange represents a single MAAS IP range entry
type IPRange interface {
	ID() int
	Type() string
	StartIP() string
	EndIP() string
	Comment() string
	Subnet() Subnet
}

type ipRange struct {
	id      int
	typ     string
	startIP string
	endIP   string
	comment string
	subnet  *subnet
}

func (r *ipRange) ID() int          { return r.id }
func (r *ipRange) Type() string     { return r.typ }
func (r *ipRange) StartIP() string  { return r.startIP }
func (r *ipRange) EndIP() string    { return r.endIP }
func (r *ipRange) Comment() string  { return r.comment }
func (r *ipRange) Subnet() Subnet   { return r.subnet }

func (r *ipRange) UnmarshalJSON(data []byte) error {
	des := &struct {
		ID      int     `json:"id"`
		Type    string  `json:"type"`
		StartIP string  `json:"start_ip"`
		EndIP   string  `json:"end_ip"`
		Comment string  `json:"comment"`
		Subnet  *subnet `json:"subnet"`
	}{}

	if err := json.Unmarshal(data, des); err != nil {
		return err
	}

	r.id = des.ID
	r.typ = des.Type
	r.startIP = des.StartIP
	r.endIP = des.EndIP
	r.comment = des.Comment
	r.subnet = des.Subnet
	return nil
}

func ipRangeSliceToInterface(in []*ipRange) []IPRange {
	var out []IPRange
	for _, r := range in {
		out = append(out, r)
	}
	return out
}

// IPRanges interface for IP range operations
type IPRanges interface {
	// List returns all IP ranges
	List(ctx context.Context) ([]IPRange, error)
	// ListBySubnet returns all IP ranges belonging to the given subnet ID
	ListBySubnet(ctx context.Context, subnetID int) ([]IPRange, error)
	// IsIPInRange returns true if the given IP falls within any range of the given subnet
	IsIPInRange(ctx context.Context, subnetID int, ip string) (bool, error)
}

type ipRanges struct {
	Controller
}

// NewIPRangesClient creates a new IP ranges client
func NewIPRangesClient(client *authenticatedClient) IPRanges {
	return &ipRanges{
		Controller: Controller{
			client:  client,
			apiPath: IPRangesAPIPath,
			params:  ParamsBuilder(),
		},
	}
}

// List returns all IP ranges
func (r *ipRanges) List(ctx context.Context) ([]IPRange, error) {
	res, err := r.client.Get(ctx, r.apiPath, r.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*ipRange
	if err = unMarshalJson(res, &obj); err != nil {
		return nil, err
	}

	return ipRangeSliceToInterface(obj), nil
}

// ListBySubnet returns all IP ranges for the given subnet ID
func (r *ipRanges) ListBySubnet(ctx context.Context, subnetID int) ([]IPRange, error) {
	all, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	var out []IPRange
	for _, rng := range all {
		if rng.Subnet() != nil && rng.Subnet().ID() == subnetID {
			out = append(out, rng)
		}
	}
	return out, nil
}

// IsIPInRange returns true if ip falls within any IP range belonging to subnetID
func (r *ipRanges) IsIPInRange(ctx context.Context, subnetID int, ip string) (bool, error) {
	checkIP := net.ParseIP(ip)
	if checkIP == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	ranges, err := r.ListBySubnet(ctx, subnetID)
	if err != nil {
		return false, err
	}

	for _, rng := range ranges {
		if ipBetween(rng.StartIP(), rng.EndIP(), checkIP) {
			return true, nil
		}
	}
	return false, nil
}

// ipBetween returns true if check is within [start, end] inclusive
func ipBetween(startStr, endStr string, check net.IP) bool {
	start := net.ParseIP(startStr)
	end := net.ParseIP(endStr)
	if start == nil || end == nil {
		return false
	}

	// Normalise to 16-byte representation for consistent comparison
	check = check.To16()
	start = start.To16()
	end = end.To16()

	return bytes.Compare(check, start) >= 0 && bytes.Compare(check, end) <= 0
}
