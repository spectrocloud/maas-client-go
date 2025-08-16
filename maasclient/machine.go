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
	"net/url"
	"strconv"
)

type Machines interface {
	List(ctx context.Context, params Params) ([]Machine, error)
	Machine(systemId string) Machine
	Allocator() MachineAllocator
}

type Machine interface {
	Get(ctx context.Context) (Machine, error)
	Delete(ctx context.Context) error
	Releaser() MachineReleaser
	Modifier() MachineModifier
	Deployer() MachineDeployer
	SystemID() string
	FQDN() string
	Zone() Zone
	PowerState() string
	Hostname() string
	IPAddresses() []net.IP
	State() string
	OSSystem() string
	DistroSeries() string
	SwapSize() int
	PowerManagerOn() PowerManagerOn
	// BootInterfaceID returns the ID of the boot interface, or empty string if not available
	BootInterfaceID() string
	// TotalStorageGB returns total storage in GB using decimal units (as MAAS reports)
	TotalStorageGB() float64
	// GetBootInterfaceType returns the type of the boot interface ("physical", "bridge", "bond", etc.)
	GetBootInterfaceType() string
}

type PowerManagerOn interface {
	WithPowerOnComment(comment string) PowerManagerOn
	PowerOn(ctx context.Context) (Machine, error)
}

type MachineReleaser interface {
	Release(ctx context.Context) (Machine, error)
	WithErase() MachineReleaser
	WithQuickErase() MachineReleaser
	WithSecureErase() MachineReleaser
	WithForce() MachineReleaser
	WithComment(comment string) MachineReleaser
}

type MachineModifier interface {
	SetSwapSize(size int) MachineModifier
	SetHostname(hostname string) MachineModifier
	Update(ctx context.Context) (Machine, error)
}

type MachineAllocator interface {
	Allocate(ctx context.Context) (Machine, error)
	WithZone(zone string) MachineAllocator
	WithSystemID(id string) MachineAllocator
	WithName(name string) MachineAllocator
	WithCPUCount(cpuCount int) MachineAllocator
	WithMemory(memory int) MachineAllocator
	WithTags(tags []string) MachineAllocator
	WithResourcePool(pool string) MachineAllocator
}

type MachineDeployer interface {
	SetOSSystem(ossytem string) MachineDeployer
	SetUserData(userdata string) MachineDeployer
	SetDistroSeries(distroseries string) MachineDeployer
	Deploy(ctx context.Context) (Machine, error)
}

type machines struct {
	Controller
}

func (m *machines) WithTags(tags []string) MachineAllocator {
	for _, tag := range tags {
		m.params.Set(TagKey, tag)
	}
	return m
}

func (m *machines) Allocate(ctx context.Context) (Machine, error) {
	res, err := m.client.Post(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return nil, err
	}

	var obj *machine
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return m.machineToInterface(obj), nil
}

func (m *machines) WithZone(zone string) MachineAllocator {
	m.params.Set(ZoneKey, zone)
	return m
}

func (m *machines) WithSystemID(id string) MachineAllocator {
	m.params.Set(SystemIDKey, id)
	return m
}

func (m *machines) WithName(name string) MachineAllocator {
	m.params.Set(NameKey, name)
	return m
}

func (m *machines) WithCPUCount(cpuCount int) MachineAllocator {
	m.params.Set(CPUCountKey, strconv.Itoa(cpuCount))
	return m
}

func (m *machines) WithMemory(memory int) MachineAllocator {
	m.params.Set(MemoryKey, strconv.Itoa(memory))
	return m
}

func (m *machines) WithResourcePool(pool string) MachineAllocator {
	m.params.Set(PoolLabel, pool)
	return m
}

func (m *machines) List(ctx context.Context, params Params) ([]Machine, error) {
	res, err := m.client.Get(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*machine
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return m.machineSliceToInterface(obj), nil
}

func (m *machines) machineSliceToInterface(in []*machine) []Machine {
	var out []Machine
	for _, machine := range in {
		out = append(out, m.machineToInterface(machine))
	}
	return out
}

func (m *machines) machineToInterface(in *machine) Machine {
	in.client = m.client
	in.apiPath = fmt.Sprintf("/machines/%s/", in.systemID)
	in.params = ParamsBuilder()
	return in
}

func (m *machines) Machine(systemId string) Machine {
	return &machine{
		Controller: Controller{
			client:  m.client,
			apiPath: fmt.Sprintf("/machines/%s/", systemId),
			params:  ParamsBuilder(),
		},
	}
}

func (m *machines) Allocator() MachineAllocator {
	m.params.Reset()
	m.params.Set(Operation, OperationAllocate)
	return m
}

type machine struct {
	Controller
	systemID          string
	fqdn              string
	zone              *zone
	powerState        string
	hostname          string
	ipaddresses       []net.IP
	state             string
	osSystem          string
	distroSeries      string
	swapSize          int
	bootInterfaceID   string
	bootInterfaceType string  // Type of boot interface (physical, bridge, bond, etc.)
	memory            int     // Memory in MB
	storageMBDecimal  float64 // Total storage in decimal MB as reported by MAAS (e.g., 250059.35)
}

func (m *machine) PowerManagerOn() PowerManagerOn {
	m.params.Reset()
	return m
}

func (m *machine) WithPowerOnComment(comment string) PowerManagerOn {
	m.params.Set(CommentKey, url.QueryEscape(comment))
	return m
}

func (m *machine) PowerOn(ctx context.Context) (Machine, error) {
	res, err := m.client.Post(context.TODO(), fmt.Sprintf("%s%s", m.apiPath, "op-power_on"), m.params.Values())
	if err != nil {
		return m, err
	}

	return m, unMarshalJson(res, &m)
}

func (m *machine) SetOSSystem(ossytem string) MachineDeployer {
	m.params.Set(OSSystemKey, ossytem)
	return m
}

func (m *machine) SetUserData(userdata string) MachineDeployer {
	m.params.Set(UserDataKey, userdata)
	return m
}

func (m *machine) SetDistroSeries(distroseries string) MachineDeployer {
	m.params.Set(DistroSeriesKey, distroseries)
	return m
}

func (m *machine) Deploy(ctx context.Context) (Machine, error) {
	res, err := m.client.Post(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return nil, err
	}

	return m, unMarshalJson(res, &m)
}

func (m *machine) Deployer() MachineDeployer {
	m.params.Reset()
	m.params.Set(Operation, OperationDeploy)
	return m
}

func (m *machine) Modifier() MachineModifier {
	m.params.Reset()
	return m
}

func (m *machine) SetSwapSize(size int) MachineModifier {
	m.params.Set(SwapSizeKey, strconv.Itoa(size))
	return m
}

func (m *machine) SetHostname(hostname string) MachineModifier {
	m.params.Set(HostnameKey, hostname)
	return m
}

func (m *machine) Update(ctx context.Context) (Machine, error) {
	res, err := m.client.PutParams(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return m, err
	}

	err = unMarshalJson(res, &m)
	return m, err
}

func (m *machine) Allocator() MachineAllocator {
	panic("implement me")
}

func (m *machine) WithErase() MachineReleaser {
	m.params.Set(EraseKey, TrueKey)
	return m
}

func (m *machine) WithQuickErase() MachineReleaser {
	m.params.Set(QuickEraseKey, TrueKey)
	return m
}

func (m *machine) WithSecureErase() MachineReleaser {
	m.params.Set(SecureEraseKey, TrueKey)
	return m
}

func (m *machine) WithForce() MachineReleaser {
	m.params.Set(ForceKey, TrueKey)
	return m
}

func (m *machine) WithComment(comment string) MachineReleaser {
	m.params.Set(CommentKey, url.QueryEscape(comment))
	return m
}

func (m *machine) Releaser() MachineReleaser {
	m.params.Reset()
	m.params.Set(Operation, OperationReleaseMachine)
	return m
}

func (m *machine) Get(ctx context.Context) (Machine, error) {
	res, err := m.client.Get(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return nil, err
	}

	return m, unMarshalJson(res, &m)
}

func (m *machine) Delete(ctx context.Context) error {
	res, err := m.client.Delete(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, nil)
}

func (m *machine) Release(ctx context.Context) (Machine, error) {
	res, err := m.client.Post(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return nil, err
	}

	return m, unMarshalJson(res, &m)
}

func (m *machine) SystemID() string {
	return m.systemID
}

func (m *machine) FQDN() string {
	return m.fqdn
}

func (m *machine) Zone() Zone {
	return m.zone
}

func (m *machine) PowerState() string {
	return m.powerState
}

func (m *machine) Hostname() string {
	return m.hostname
}

func (m *machine) IPAddresses() []net.IP {
	return m.ipaddresses
}

func (m *machine) State() string {
	return m.state
}

func (m *machine) OSSystem() string {
	return m.osSystem
}

func (m *machine) DistroSeries() string {
	return m.distroSeries
}

func (m *machine) SwapSize() int {
	return m.swapSize
}

func (m *machine) BootInterfaceID() string {
	return m.bootInterfaceID
}

// TotalStorageGB returns total storage in GB using decimal units (as MAAS reports)
func (m *machine) TotalStorageGB() float64 {
	// MAAS storage field is in decimal MB (e.g., 250059.350016)
	// Convert MB (decimal) to GB (decimal) by dividing by 1000
	if m.storageMBDecimal <= 0 {
		return 0
	}
	return m.storageMBDecimal / 1000.0
}

// GetBootInterfaceType returns the type of the boot interface
func (m *machine) GetBootInterfaceType() string {
	return m.bootInterfaceType
}

func (m *machine) UnmarshalJSON(data []byte) error {
	des := &struct {
		SystemID      string   `json:"system_id"`
		FQDNLabel     string   `json:"fqdn"`
		Zone          *zone    `json:"zone"`
		PowerState    string   `json:"power_state"`
		Hostname      string   `json:"hostname"`
		IpAddresses   []string `json:"ip_addresses"`
		State         string   `json:"status_name"`
		OSSystem      string   `json:"osystem"`
		DistroSeries  string   `json:"distro_series"`
		SwapSize      int      `json:"swap_size"`
		Memory        int      `json:"memory"`
		Storage       float64  `json:"storage"`
		BootInterface struct {
			ID       int      `json:"id"`
			Type     string   `json:"type"`
			Children []string `json:"children"`
		} `json:"boot_interface"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	m.systemID = des.SystemID
	m.fqdn = des.FQDNLabel
	m.zone = des.Zone
	m.powerState = des.PowerState
	m.hostname = des.Hostname
	for _, ipAddress := range des.IpAddresses {
		m.ipaddresses = append(m.ipaddresses, net.ParseIP(ipAddress))
	}
	m.state = des.State
	m.osSystem = des.OSSystem
	m.distroSeries = des.DistroSeries
	m.swapSize = des.SwapSize
	m.memory = des.Memory
	m.storageMBDecimal = des.Storage

	// Handle boot interface
	if des.BootInterface.ID != 0 {
		m.bootInterfaceID = fmt.Sprintf("%d", des.BootInterface.ID)

		// Simple rule: children present = bridge, children empty/absent = physical
		if len(des.BootInterface.Children) > 0 {
			m.bootInterfaceType = "bridge"
		} else {
			m.bootInterfaceType = "physical"
		}
	}

	return nil
}

func NewMachinesClient(client *authenticatedClient) Machines {
	return &machines{Controller{
		client:  client,
		apiPath: "/machines/",
		params:  ParamsBuilder(),
	}}
}
