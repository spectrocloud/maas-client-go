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
	"encoding/json"
	"fmt"
)

// NodeDevices provides methods to interact with devices discovered on a machine
type NodeDevices interface {
	// List retrieves devices for the machine identified by systemID.
	// Optional params filter the result, e.g. HardwareTypeKey=HardwareTypeGPU
	// to return only GPU devices, or BusKey, VendorIDKey, ProductIDKey,
	// VendorNameKey, ProductNameKey, CommissioningDriverKey. A nil params
	// returns all devices.
	List(ctx context.Context, systemID string, params Params) ([]NodeDevice, error)
}

// NodeDevice represents a single device discovered on a machine during commissioning
type NodeDevice interface {
	ID() string
	// BusName is the bus the device is attached to, e.g. "PCIE" or "USB"
	BusName() string
	// HardwareTypeName is the type of device, e.g. "node", "cpu", "memory", "storage", "network", "gpu"
	HardwareTypeName() string
	VendorID() string
	ProductID() string
	VendorName() string
	ProductName() string
	CommissioningDriver() string
	BusNumber() int
	DeviceNumber() int
	PCIAddress() string
	NUMANode() int
	SystemID() string
	ResourceUri() string
}

type nodeDevices struct {
	Controller
}

func (nd *nodeDevices) List(ctx context.Context, systemID string, params Params) ([]NodeDevice, error) {
	if params == nil {
		params = ParamsBuilder()
	}
	path := fmt.Sprintf("/nodes/%s/devices/", systemID)
	res, err := nd.client.Get(ctx, path, params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*nodeDevice
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return nodeDeviceSliceToInterface(obj), nil
}

type nodeDevice struct {
	id                  string
	busName             string
	hardwareTypeName    string
	vendorID            string
	productID           string
	vendorName          string
	productName         string
	commissioningDriver string
	busNumber           int
	deviceNumber        int
	pciAddress          string
	numaNode            int
	systemID            string
	resourceUri         string
}

func (d *nodeDevice) ID() string {
	return d.id
}

func (d *nodeDevice) BusName() string {
	return d.busName
}

func (d *nodeDevice) HardwareTypeName() string {
	return d.hardwareTypeName
}

func (d *nodeDevice) VendorID() string {
	return d.vendorID
}

func (d *nodeDevice) ProductID() string {
	return d.productID
}

func (d *nodeDevice) VendorName() string {
	return d.vendorName
}

func (d *nodeDevice) ProductName() string {
	return d.productName
}

func (d *nodeDevice) CommissioningDriver() string {
	return d.commissioningDriver
}

func (d *nodeDevice) BusNumber() int {
	return d.busNumber
}

func (d *nodeDevice) DeviceNumber() int {
	return d.deviceNumber
}

func (d *nodeDevice) PCIAddress() string {
	return d.pciAddress
}

func (d *nodeDevice) NUMANode() int {
	return d.numaNode
}

func (d *nodeDevice) SystemID() string {
	return d.systemID
}

func (d *nodeDevice) ResourceUri() string {
	return d.resourceUri
}

func (d *nodeDevice) UnmarshalJSON(data []byte) error {
	des := struct {
		ID                  int    `json:"id"`
		BusName             string `json:"bus_name"`
		HardwareTypeName    string `json:"hardware_type_name"`
		VendorID            string `json:"vendor_id"`
		ProductID           string `json:"product_id"`
		VendorName          string `json:"vendor_name"`
		ProductName         string `json:"product_name"`
		CommissioningDriver string `json:"commissioning_driver"`
		BusNumber           int    `json:"bus_number"`
		DeviceNumber        int    `json:"device_number"`
		PCIAddress          string `json:"pci_address"`
		NUMANode            int    `json:"numa_node"`
		SystemID            string `json:"system_id"`
		ResourceUri         string `json:"resource_uri"`
	}{}

	err := json.Unmarshal(data, &des)
	if err != nil {
		return err
	}

	d.id = fmt.Sprintf("%d", des.ID)
	d.busName = des.BusName
	d.hardwareTypeName = des.HardwareTypeName
	d.vendorID = des.VendorID
	d.productID = des.ProductID
	d.vendorName = des.VendorName
	d.productName = des.ProductName
	d.commissioningDriver = des.CommissioningDriver
	d.busNumber = des.BusNumber
	d.deviceNumber = des.DeviceNumber
	d.pciAddress = des.PCIAddress
	d.numaNode = des.NUMANode
	d.systemID = des.SystemID
	d.resourceUri = des.ResourceUri

	return nil
}

func nodeDeviceSliceToInterface(in []*nodeDevice) []NodeDevice {
	var out []NodeDevice
	for _, d := range in {
		out = append(out, d)
	}
	return out
}

func NewNodeDevicesClient(client *authenticatedClient) NodeDevices {
	return &nodeDevices{
		Controller: Controller{
			client:  client,
			apiPath: "/nodes/", // Base path, will be extended per operation
			params:  ParamsBuilder(),
		},
	}
}
