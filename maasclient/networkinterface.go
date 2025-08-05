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
	"net"
)

// NetworkInterfaces provides methods to interact with machine network interfaces
type NetworkInterfaces interface {
	// Get retrieves all network interfaces for a machine
	Get(ctx context.Context, systemID string) ([]NetworkInterface, error)
	// Interface returns a NetworkInterface for a specific interface ID
	Interface(systemID, interfaceID string) NetworkInterface
	// SetBootInterfaceStaticIP sets a static IP on the boot interface directly
	SetBootInterfaceStaticIP(ctx context.Context, systemID, ipAddress string) error
}

// NetworkInterface represents a single network interface on a machine
type NetworkInterface interface {
	// Get retrieves the interface details
	Get(ctx context.Context) (NetworkInterface, error)
	// LinkSubnet links a subnet to this interface
	LinkSubnet(ctx context.Context, subnetID string, ipAddress string) error
	// UnlinkSubnet unlinks a subnet from this interface
	UnlinkSubnet(ctx context.Context, linkID string) error
	// UpdateIPConfiguration updates an existing link's IP configuration directly
	UpdateIPConfiguration(ctx context.Context, config IPConfigurationUpdate) error
	// SetStaticIP sets a static IP on the interface
	// Handles two valid scenarios:
	// 1. Interface has direct links - configures directly
	// 2. Interface has children (bridge) - configures on child with links
	SetStaticIP(ctx context.Context, ipAddress string) error
	// SetDHCP sets the interface to use DHCP (handles existing links automatically)
	SetDHCP(ctx context.Context, subnetID string) error

	// Getters for interface properties
	ID() string
	Name() string
	Type() string
	Enabled() bool
	MACAddress() string
	Links() []NetworkInterfaceLink
	Children() []string
	VLAN() VLAN
}

// IPConfigurationUpdate represents the parameters for updating IP configuration
type IPConfigurationUpdate struct {
	LinkID    string  // Required: ID of the link to update
	Mode      string  // Required: "static", "dhcp", or "link_up"
	IPAddress *string // Required for static mode: IP address
	SubnetID  *string // Required: subnet ID
}

// NetworkInterfaceLink represents a link (IP configuration) on a network interface
type NetworkInterfaceLink interface {
	ID() string
	Mode() string
	Subnet() Subnet
	IPAddress() net.IP
}

type networkInterfaces struct {
	Controller
}

type networkInterface struct {
	Controller
	systemID    string
	interfaceID string
	id          string
	name        string
	ifType      string
	enabled     bool
	macAddress  string
	links       []*networkInterfaceLink
	children    []string
	vlan        *vlan
}

type networkInterfaceLink struct {
	id        string
	mode      string
	subnet    *subnet
	ipAddress net.IP
}

type vlan struct {
	id         int
	vid        int
	name       string
	fabricID   int
	fabricName string
	mtu        int
	dhcpOn     bool
}

// NetworkInterfaces implementation
func (ni *networkInterfaces) Get(ctx context.Context, systemID string) ([]NetworkInterface, error) {
	path := fmt.Sprintf("/nodes/%s/interfaces/", systemID)
	res, err := ni.client.Get(ctx, path, ni.params.Values())
	if err != nil {
		return nil, err
	}

	var interfaces []*networkInterface
	err = unMarshalJson(res, &interfaces)
	if err != nil {
		return nil, err
	}

	// Set system ID for each interface
	for _, iface := range interfaces {
		iface.systemID = systemID
		iface.client = ni.client
	}

	return networkInterfaceSliceToInterface(interfaces, ni.client), nil
}

func (ni *networkInterfaces) Interface(systemID, interfaceID string) NetworkInterface {
	return &networkInterface{
		Controller: Controller{
			client:  ni.client,
			apiPath: fmt.Sprintf("/nodes/%s/interfaces/%s/", systemID, interfaceID),
			params:  ParamsBuilder(),
		},
		systemID:    systemID,
		interfaceID: interfaceID,
	}
}

func (ni *networkInterfaces) SetBootInterfaceStaticIP(ctx context.Context, systemID, ipAddress string) error {
	// Get machine details to find boot interface ID
	machineClient := &machine{
		Controller: Controller{
			client:  ni.client,
			apiPath: fmt.Sprintf("/machines/%s/", systemID),
			params:  ParamsBuilder(),
		},
		systemID: systemID,
	}

	machineDetails, err := machineClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get machine details: %v", err)
	}

	bootInterfaceID := machineDetails.BootInterfaceID()
	if bootInterfaceID == "" {
		return fmt.Errorf("no boot interface found for machine %s", systemID)
	}

	// Get the boot interface directly
	bootInterface := ni.Interface(systemID, bootInterfaceID)

	// Populate the interface data by calling Get()
	bootInterface, err = bootInterface.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get boot interface details: %v", err)
	}

	// Use the enhanced SetStaticIP that handles both direct links and bridge scenarios
	return bootInterface.SetStaticIP(ctx, ipAddress)
}

// NetworkInterface implementation
func (ni *networkInterface) Get(ctx context.Context) (NetworkInterface, error) {
	res, err := ni.client.Get(ctx, ni.apiPath, ni.params.Values())
	if err != nil {
		return nil, err
	}

	err = unMarshalJson(res, ni)
	if err != nil {
		return nil, err
	}

	return ni, nil
}

func (ni *networkInterface) LinkSubnet(ctx context.Context, subnetID string, ipAddress string) error {
	ni.params.Reset()
	ni.params.Set(Operation, OperationLinkSubnet)
	ni.params.Set(SubnetKey, subnetID)

	// Set mode based on whether IP address is provided
	if ipAddress != "" {
		ni.params.Set(IPAddressKey, ipAddress)
		ni.params.Set(ModeKey, ModeStatic) // Static mode when IP is provided
	} else {
		ni.params.Set(ModeKey, ModeDHCP) // DHCP mode when no IP
	}

	res, err := ni.client.Post(ctx, ni.apiPath, ni.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, nil)
}

func (ni *networkInterface) UnlinkSubnet(ctx context.Context, linkID string) error {
	ni.params.Reset()
	ni.params.Set(Operation, OperationUnlinkSubnet)
	ni.params.Set(LinkIDKey, linkID)

	res, err := ni.client.Post(ctx, ni.apiPath, ni.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, nil)
}

func (ni *networkInterface) UpdateIPConfiguration(ctx context.Context, config IPConfigurationUpdate) error {
	// Validate required parameters
	if config.LinkID == "" {
		return fmt.Errorf("LinkID is required")
	}
	if config.Mode == "" {
		return fmt.Errorf("Mode is required")
	}
	if config.SubnetID == nil {
		return fmt.Errorf("SubnetID is required")
	}

	// Validate mode-specific requirements
	if config.Mode == ModeStatic {
		if config.IPAddress == nil || *config.IPAddress == "" {
			return fmt.Errorf("IPAddress is required for static mode")
		}
	}

	// MAAS doesn't have update_link operation, so we unlink and relink
	// First unlink the existing configuration
	err := ni.UnlinkSubnet(ctx, config.LinkID)
	if err != nil {
		return fmt.Errorf("failed to unlink existing configuration: %w", err)
	}

	// Then link with new configuration
	subnetID := *config.SubnetID

	// For static mode, use the IP address; for DHCP, use empty string
	if config.Mode == ModeStatic {
		ipAddress := *config.IPAddress
		return ni.LinkSubnet(ctx, subnetID, ipAddress)
	} else {
		// For DHCP or other modes
		ni.params.Reset()
		ni.params.Set(Operation, OperationLinkSubnet)
		ni.params.Set(SubnetKey, subnetID)
		ni.params.Set(ModeKey, config.Mode)

		res, err := ni.client.Post(ctx, ni.apiPath, ni.params.Values())
		if err != nil {
			return err
		}
		return unMarshalJson(res, nil)
	}
}

func (ni *networkInterface) SetStaticIP(ctx context.Context, ipAddress string) error {
	// Ensure the interface has proper API path and client setup
	if ni.apiPath == "" {
		ni.apiPath = fmt.Sprintf("/nodes/%s/interfaces/%s/", ni.systemID, ni.interfaceID)
	}
	if ni.params == nil {
		ni.params = ParamsBuilder()
	}

	// Get current links and children to determine the strategy
	// Boot interfaces should always have either direct links or children (bridge scenario)
	links := ni.Links()
	children := ni.Children()

	// Case 1: Interface has direct links - configure directly
	if len(links) > 0 {
		var targetLink NetworkInterfaceLink

		// Prefer DHCP links first, then any link with a subnet
		for _, link := range links {
			if link.Mode() == ModeDHCP {
				targetLink = link
				break
			}
		}

		// If no DHCP link found, use any link with a subnet
		if targetLink == nil {
			for _, link := range links {
				if link.Subnet() != nil {
					targetLink = link
					break
				}
			}
		}

		// Fallback to first link if no suitable link found
		if targetLink == nil {
			targetLink = links[0]
		}

		// Use the subnet from the existing link
		if targetLink.Subnet() == nil {
			return fmt.Errorf("target link has no subnet information")
		}
		subnetID := fmt.Sprintf("%d", targetLink.Subnet().ID())

		config := IPConfigurationUpdate{
			LinkID:    targetLink.ID(),
			Mode:      ModeStatic,
			IPAddress: &ipAddress,
			SubnetID:  &subnetID,
		}
		return ni.UpdateIPConfiguration(ctx, config)
	}

	// Case 2: No direct links but has children (bridge scenario)
	if len(children) > 0 {
		// Get all interfaces from the parent client to find children
		networkInterfaces := &networkInterfaces{
			Controller: Controller{
				client: ni.client,
				params: ParamsBuilder(),
			},
		}

		allInterfaces, err := networkInterfaces.Get(ctx, ni.systemID)
		if err != nil {
			return fmt.Errorf("failed to get interfaces for bridge detection: %v", err)
		}

		// Find a child interface with actual network links
		for _, childName := range children {
			for _, iface := range allInterfaces {
				if iface.Name() == childName && len(iface.Links()) > 0 {
					// Recursively call SetStaticIP on the child interface with links
					return iface.SetStaticIP(ctx, ipAddress)
				}
			}
		}

		return fmt.Errorf("no child interface with links found for bridge configuration")
	}

	// Case 3: Invalid configuration - boot interface should have either links or children
	return fmt.Errorf("invalid boot interface configuration: no links and no children found for interface %s", ni.name)
}

func (ni *networkInterface) SetDHCP(ctx context.Context, subnetID string) error {
	// Ensure the interface has proper API path and client setup
	if ni.apiPath == "" {
		ni.apiPath = fmt.Sprintf("/nodes/%s/interfaces/%s/", ni.systemID, ni.interfaceID)
	}
	if ni.params == nil {
		ni.params = ParamsBuilder()
	}

	// Get current links from the existing interface object
	links := ni.Links()

	// If there are existing links, update the first one to DHCP
	if len(links) > 0 {
		firstLink := links[0]
		config := IPConfigurationUpdate{
			LinkID:   firstLink.ID(),
			Mode:     ModeDHCP,
			SubnetID: &subnetID,
		}
		return ni.UpdateIPConfiguration(ctx, config)
	}

	// If no existing links, create a new DHCP link
	return ni.LinkSubnet(ctx, subnetID, "")
}

// Getters
func (ni *networkInterface) ID() string {
	return ni.id
}

func (ni *networkInterface) Name() string {
	return ni.name
}

func (ni *networkInterface) Type() string {
	return ni.ifType
}

func (ni *networkInterface) Enabled() bool {
	return ni.enabled
}

func (ni *networkInterface) MACAddress() string {
	return ni.macAddress
}

func (ni *networkInterface) Links() []NetworkInterfaceLink {
	return networkInterfaceLinkSliceToInterface(ni.links)
}

func (ni *networkInterface) Children() []string {
	return ni.children
}

func (ni *networkInterface) VLAN() VLAN {
	return ni.vlan
}

// JSON unmarshaling
func (ni *networkInterface) UnmarshalJSON(data []byte) error {
	type Alias networkInterface
	aux := &struct {
		ID         int                     `json:"id"`
		Name       string                  `json:"name"`
		Type       string                  `json:"type"`
		Enabled    bool                    `json:"enabled"`
		MACAddress string                  `json:"mac_address"`
		Links      []*networkInterfaceLink `json:"links"`
		Children   []string                `json:"children"`
		VLAN       *vlan                   `json:"vlan"`
		*Alias
	}{
		Alias: (*Alias)(ni),
	}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	ni.id = fmt.Sprintf("%d", aux.ID)
	ni.name = aux.Name
	ni.ifType = aux.Type
	ni.enabled = aux.Enabled
	ni.macAddress = aux.MACAddress
	ni.links = aux.Links
	ni.children = aux.Children
	ni.vlan = aux.VLAN

	return nil
}

func (link *networkInterfaceLink) UnmarshalJSON(data []byte) error {
	// First try to unmarshal with int ID (real MAAS API)
	auxInt := &struct {
		ID        int     `json:"id"`
		Mode      string  `json:"mode"`
		Subnet    *subnet `json:"subnet"`
		IPAddress string  `json:"ip_address"`
	}{}

	if err := json.Unmarshal(data, &auxInt); err == nil {
		// Success with int ID
		link.id = fmt.Sprintf("%d", auxInt.ID)
		link.mode = auxInt.Mode
		link.subnet = auxInt.Subnet
		if auxInt.IPAddress != "" {
			link.ipAddress = net.ParseIP(auxInt.IPAddress)
		}
		return nil
	}

	// Fallback to string ID (for tests/mocks)
	auxString := &struct {
		ID        string  `json:"id"`
		Mode      string  `json:"mode"`
		Subnet    *subnet `json:"subnet"`
		IPAddress string  `json:"ip_address"`
	}{}

	if err := json.Unmarshal(data, &auxString); err != nil {
		return err
	}

	link.id = auxString.ID
	link.mode = auxString.Mode
	link.subnet = auxString.Subnet
	if auxString.IPAddress != "" {
		link.ipAddress = net.ParseIP(auxString.IPAddress)
	}

	return nil
}

func (v *vlan) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ID         int    `json:"id"`
		VID        int    `json:"vid"`
		Name       string `json:"name"`
		FabricID   int    `json:"fabric_id"`
		FabricName string `json:"fabric"`
		MTU        int    `json:"mtu"`
		DHCPOn     bool   `json:"dhcp_on"`
	}{}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	v.id = aux.ID
	v.vid = aux.VID
	v.name = aux.Name
	v.fabricID = aux.FabricID
	v.fabricName = aux.FabricName
	v.mtu = aux.MTU
	v.dhcpOn = aux.DHCPOn

	return nil
}

// VLAN interface implementation
func (v *vlan) VID() int {
	return v.vid
}

func (v *vlan) MTU() int {
	return v.mtu
}

func (v *vlan) IsDHCPOn() bool {
	return v.dhcpOn
}

func (v *vlan) FabricID() int {
	return v.fabricID
}

func (v *vlan) FabricName() string {
	return v.fabricName
}

// NetworkInterfaceLink implementation
func (link *networkInterfaceLink) ID() string {
	return link.id
}

func (link *networkInterfaceLink) Mode() string {
	return link.mode
}

func (link *networkInterfaceLink) Subnet() Subnet {
	return link.subnet
}

func (link *networkInterfaceLink) IPAddress() net.IP {
	return link.ipAddress
}

// Helper conversion functions
func networkInterfaceSliceToInterface(in []*networkInterface, client Client) []NetworkInterface {
	var out []NetworkInterface
	for _, ni := range in {
		out = append(out, networkInterfaceStructToInterface(ni, client))
	}
	return out
}

func networkInterfaceStructToInterface(in *networkInterface, client Client) NetworkInterface {
	in.client = client
	// Set up the interface ID for API calls if available
	if in.id != "" {
		in.interfaceID = in.id
	}
	return in
}

func networkInterfaceLinkSliceToInterface(in []*networkInterfaceLink) []NetworkInterfaceLink {
	var out []NetworkInterfaceLink
	for _, link := range in {
		out = append(out, link)
	}
	return out
}

// Constructor
func NewNetworkInterfacesClient(client *authenticatedClient) NetworkInterfaces {
	return &networkInterfaces{
		Controller: Controller{
			client:  client,
			apiPath: "/nodes/", // Base path, will be extended per operation
			params:  ParamsBuilder(),
		},
	}
}
