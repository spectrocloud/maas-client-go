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
	Get(ctx context.Context) error
	Delete(ctx context.Context) error
	Releaser() MachineReleaser
	Modifier() MachineModifier
	Deployer() MachineDeployer
	Deploy(ctx context.Context) error
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
}

type MachineReleaser interface {
	Release(ctx context.Context) error
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
	WithResourcePool(pool string) MachineAllocator
}

type MachineDeployer interface {
	SetOSSystem(ossytem string) MachineDeployer
	SetUserData(userdata string) MachineDeployer
	SetDistroSeries(distroseries string) MachineDeployer
	Deploy(ctx context.Context) error
}

type machines struct {
	Controller
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
	systemID     string
	fqdn         string
	zone         *zone
	powerState   string
	hostname     string
	ipaddresses  []net.IP
	state        string
	osSystem     string
	distroSeries string
	swapSize     int
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

func (m *machine) Deploy(ctx context.Context) error {
	res, err := m.client.Post(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, &m)
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

func (m *machine) Get(ctx context.Context) error {
	res, err := m.client.Get(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, &m)
}

func (m *machine) Delete(ctx context.Context) error {
	res, err := m.client.Delete(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, nil)
}

func (m *machine) Release(ctx context.Context) error {
	res, err := m.client.Post(ctx, m.apiPath, m.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, &m)
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

func (m *machine) UnmarshalJSON(data []byte) error {
	des := &struct {
		SystemID     string   `json:"system_id"`
		FQDNLabel    string   `json:"fqdn"`
		Zone         *zone    `json:"zone"`
		PowerState   string   `json:"power_state"`
		Hostname     string   `json:"hostname"`
		IpAddresses  []string `json:"ip_addresses"`
		State        string   `json:"status_name"`
		OSSystem     string   `json:"osystem"`
		DistroSeries string   `json:"distro_series"`
		SwapSize     int      `json:"swap_size"`
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

	return nil
}

func NewMachinesClient(client *authenticatedClient) Machines {
	return &machines{Controller{
		client:  client,
		apiPath: "/machines/",
		params:  ParamsBuilder(),
	}}
}
