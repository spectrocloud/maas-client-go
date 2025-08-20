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
	"strconv"
)

// VMHosts interface for managing VM hosts collection
type VMHosts interface {
	List(ctx context.Context, params Params) ([]VMHost, error)
	Create(ctx context.Context, params Params) (VMHost, error)
	VMHost(systemID string) VMHost
}

// VMHost interface for managing individual VM host
type VMHost interface {
	Get(ctx context.Context) (VMHost, error)
	Update(ctx context.Context, params Params) (VMHost, error)
	Delete(ctx context.Context) error

	// VM Host specific operations
	Composer() VMComposer
	Machines() VMHostMachines

	// Properties
	SystemID() string
	Name() string
	Type() string // "lxd", "virsh"
	PowerAddress() string
	Zone() Zone
	ResourcePool() ResourcePool

	// Resource information
	TotalCores() int
	TotalMemory() int
	UsedCores() int
	UsedMemory() int
	AvailableCores() int
	AvailableMemory() int

	// Capabilities
	Capabilities() []string
	Projects() []string
	StoragePools() []StoragePool
}

// VMComposer interface for composing VMs on a VM host
type VMComposer interface {
	Compose(ctx context.Context, params Params) (Machine, error)
}

// VMHostMachines interface for managing machines on a VM host
type VMHostMachines interface {
	List(ctx context.Context) ([]Machine, error)
}

// vmHosts implements VMHosts interface following the Controller pattern
type vmHosts struct {
	Controller
}

// vmHost implements VMHost interface following the Controller pattern
type vmHost struct {
	Controller
	systemID string
	data     vmHostDetails
}

// vmComposer implements VMComposer interface
type vmComposer struct {
	Controller
	systemID string
}

// vmHostMachines implements VMHostMachines interface
type vmHostMachines struct {
	Controller
	systemID string
}

// NewVMHostsClient creates a new VMHosts client
func NewVMHostsClient(client *authenticatedClient) VMHosts {
	return &vmHosts{Controller{
		client:  client,
		apiPath: "/vm-hosts/",
		params:  ParamsBuilder(),
	}}
}

// Implementation of VMHosts interface
func (c *vmHosts) List(ctx context.Context, params Params) ([]VMHost, error) {
	qsp := url.Values{}
	if params != nil {
		qsp = params.Values()
	}

	resp, err := c.client.Get(ctx, c.apiPath, qsp)
	if err != nil {
		return nil, err
	}

	var vmHostsResponse []vmHostDetails
	if err := unMarshalJson(resp, &vmHostsResponse); err != nil {
		return nil, err
	}

	var result []VMHost
	for _, vmHostData := range vmHostsResponse {
		systemIDStr := strconv.Itoa(vmHostData.ID)
		result = append(result, &vmHost{
			Controller: Controller{
				client:  c.client,
				apiPath: fmt.Sprintf("/vm-hosts/%s/", systemIDStr),
				params:  ParamsBuilder(),
			},
			systemID: systemIDStr,
			data:     vmHostData,
		})
	}

	return result, nil
}

func (c *vmHosts) Create(ctx context.Context, params Params) (VMHost, error) {
	resp, err := c.client.Post(ctx, c.apiPath, params.Values())
	if err != nil {
		return nil, err
	}

	var vmHostData vmHostDetails
	if err := unMarshalJson(resp, &vmHostData); err != nil {
		return nil, err
	}

	systemIDStr := strconv.Itoa(vmHostData.ID)

	return &vmHost{
		Controller: Controller{
			client:  c.client,
			apiPath: fmt.Sprintf("/vm-host/%s/", systemIDStr),
			params:  ParamsBuilder(),
		},
		systemID: systemIDStr,
		data:     vmHostData,
	}, nil
}

func (c *vmHosts) VMHost(systemID string) VMHost {
	return &vmHost{
		Controller: Controller{
			client:  c.client,
			apiPath: fmt.Sprintf("/vm-hosts/%s/", systemID),
			params:  ParamsBuilder(),
		},
		systemID: systemID,
		data:     vmHostDetails{}, // Empty data, will be populated on Get()
	}
}

// Implementation of VMHost interface
func (c *vmHost) Get(ctx context.Context) (VMHost, error) {
	resp, err := c.client.Get(ctx, c.apiPath, url.Values{})
	if err != nil {
		return nil, err
	}

	var vmHostData vmHostDetails
	if err := unMarshalJson(resp, &vmHostData); err != nil {
		return nil, err
	}

	// Update the vmHost with the fetched data
	return &vmHost{
		Controller: c.Controller,
		systemID:   strconv.Itoa(vmHostData.ID),
		data:       vmHostData,
	}, nil
}

func (c *vmHost) Update(ctx context.Context, params Params) (VMHost, error) {
	resp, err := c.client.PutParams(ctx, c.apiPath, params.Values())
	if err != nil {
		return nil, err
	}

	var vmHostData vmHostDetails
	if err := unMarshalJson(resp, &vmHostData); err != nil {
		return nil, err
	}

	return &vmHost{
		Controller: c.Controller,
		systemID:   c.SystemID(),
		data:       vmHostData,
	}, nil
}

func (c *vmHost) Delete(ctx context.Context) error {
	_, err := c.client.Delete(ctx, c.apiPath, url.Values{})
	return err
}

func (c *vmHost) Composer() VMComposer {
	return &vmComposer{
		Controller: Controller{
			client:  c.client,
			apiPath: fmt.Sprintf("/vm-hosts/%s/", c.systemID),
			params:  ParamsBuilder(),
		},
		systemID: c.systemID,
	}
}

func (c *vmHost) Machines() VMHostMachines {
	return &vmHostMachines{
		Controller: Controller{
			client:  c.client,
			apiPath: fmt.Sprintf("/vm-hosts/%s/machines/", c.systemID),
			params:  ParamsBuilder(),
		},
		systemID: c.systemID,
	}
}

// Implementation of VMComposer interface
func (c *vmComposer) Compose(ctx context.Context, params Params) (Machine, error) {
	composePath := fmt.Sprintf("/vm-hosts/%s/?op=compose", c.systemID)

	resp, err := c.client.Post(ctx, composePath, params.Values())
	if err != nil {
		return nil, err
	}

	var machineData machineDetails
	if err := unMarshalJson(resp, &machineData); err != nil {
		return nil, err
	}

	// Return a Machine client for the composed machine
	return NewMachinesClient(c.client.(*authenticatedClient)).Machine(machineData.SystemID), nil
}

// Implementation of VMHostMachines interface
func (c *vmHostMachines) List(ctx context.Context) ([]Machine, error) {
	resp, err := c.client.Get(ctx, c.apiPath, url.Values{})
	if err != nil {
		return nil, err
	}

	var machinesData []machineDetails
	if err := unMarshalJson(resp, &machinesData); err != nil {
		return nil, err
	}

	var result []Machine
	machinesClient := NewMachinesClient(c.client.(*authenticatedClient))
	for _, machineData := range machinesData {
		result = append(result, machinesClient.Machine(machineData.SystemID))
	}

	return result, nil
}

func (c *vmHost) Zone() Zone {
	if c.data.Zone.Name == "" || c.data.Zone.ID == 0 {
		return nil
	}
	return &zone{name: c.data.Zone.Name, id: c.data.Zone.ID}
}

func (c *vmHost) ResourcePool() ResourcePool {
	if c.data.Pool.Name == "" || c.data.Pool.ID <= 0 {
		return nil
	}
	return &resourcePool{name: c.data.Pool.Name, id: c.data.Pool.ID}
}

// VMHost property implementations - these are populated from API response data
func (c *vmHost) SystemID() string            { return c.systemID }
func (c *vmHost) Name() string                { return c.data.Name }
func (c *vmHost) Type() string                { return c.data.Type }
func (c *vmHost) PowerAddress() string        { return c.data.PowerAddress }
func (c *vmHost) TotalCores() int             { return c.data.TotalResources.Cores }
func (c *vmHost) TotalMemory() int            { return c.data.TotalResources.Memory }
func (c *vmHost) UsedCores() int              { return c.data.UsedResources.Cores }
func (c *vmHost) UsedMemory() int             { return c.data.UsedResources.Memory }
func (c *vmHost) AvailableCores() int         { return c.data.AvailableResources.Cores }
func (c *vmHost) AvailableMemory() int        { return c.data.AvailableResources.Memory }
func (c *vmHost) Capabilities() []string      { return c.data.Capabilities }
func (c *vmHost) Projects() []string          { return c.data.Projects }
func (c *vmHost) StoragePools() []StoragePool { return c.data.StoragePools }

// Response structures for API unmarshaling
type vmHostDetails struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`

	PowerAddress string `json:"power_address"`

	Zone struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"zone"`

	Pool struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"pool"`

	TotalResources struct {
		Cores  int `json:"cores"`
		Memory int `json:"memory"`
	} `json:"total"`

	UsedResources struct {
		Cores  int `json:"cores"`
		Memory int `json:"memory"`
	} `json:"used"`

	AvailableResources struct {
		Cores  int `json:"cores"`
		Memory int `json:"memory"`
	} `json:"available"`

	Capabilities []string `json:"capabilities"`
	Projects     []string `json:"projects"`

	StoragePools []StoragePool `json:"storage_pools"`
}

// machineDetails structure for machine responses
type machineDetails struct {
	SystemID string `json:"system_id"`
}

// StoragePool represents one entry in the storage_pools array
type StoragePool struct {
	Name      string `json:"name"`
	Driver    string `json:"driver,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Used      int64  `json:"used,omitempty"`
	Pending   int64  `json:"pending,omitempty"`
	Available int64  `json:"avail,omitempty"`
	Remote    bool   `json:"remote,omitempty"`
}
