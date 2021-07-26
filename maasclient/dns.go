package maasclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	DNSResourcesAPIPath  = "/dnsresources/"
	DNSResourceAPIFormat = "/dnsresources/%d/"
)

type DNSResources interface {
	List(ctx context.Context, params Params) ([]DNSResource, error)
	Builder() DNSResourceBuilder
	DNSResource(id int) DNSResource
}

type DNSResource interface {
	Delete(ctx context.Context) error
	Modifier() DNSResourceModifier
	Get(ctx context.Context) error
	ID() int
	FQDN() string
	AddressTTL() int
	IPAddresses() []IPAddress
}

type DNSResourceModifier interface {
	SetFQDN(fqdn string) DNSResourceModifier
	SetAddressTTL(addressTTL int) DNSResourceModifier
	SetIPAddresses(address []string) DNSResourceModifier
	SetName(name string) DNSResourceModifier
	SetDomain(name string) DNSResourceModifier
	Modify(ctx context.Context) error
}

type DNSResourceBuilder interface {
	WithFQDN(fqdn string) DNSResourceBuilder
	WithDomain(domain string) DNSResourceBuilder
	WithName(name string) DNSResourceBuilder
	WithAddressTTL(addressTTL string) DNSResourceBuilder
	WithIPAddresses(ipAddresses []string) DNSResourceBuilder
	Create(ctx context.Context) (DNSResource, error)
}

// DNSResource
type dnsResource struct {
	id          int
	fqdn        string
	addressTTL  *int
	ipAddresses []*ipaddress
	Controller
}

func (d *dnsResource) Get(ctx context.Context) error {
	res, err := d.client.Get(ctx, d.apiPath, d.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(res, d)
}

func (d *dnsResource) Delete(ctx context.Context) error {
	data, err := d.client.Delete(ctx, d.apiPath, nil)
	if err != nil {
		return err
	}

	return unMarshalJson(data, nil)
}

func (d *dnsResource) Modifier() DNSResourceModifier {
	d.params.Reset()
	return d
}

func (d *dnsResource) SetFQDN(fqdn string) DNSResourceModifier {
	d.params.Add(FQDNKey, fqdn)
	return d
}

func (d *dnsResource) SetAddressTTL(addressTTL int) DNSResourceModifier {
	d.params.Add(AddressTTLKey, strconv.Itoa(addressTTL))
	return d
}

func (d *dnsResource) SetIPAddresses(address []string) DNSResourceModifier {
	d.params.Add(IPAddressesKey, strings.Join(address, " "))
	return d
}

func (d *dnsResource) SetName(name string) DNSResourceModifier {
	d.params.Add(NameKey, name)
	return d
}

func (d *dnsResource) SetDomain(domain string) DNSResourceModifier {
	d.params.Add(DomainKey, domain)
	return d
}

func (d *dnsResource) Modify(ctx context.Context) error {
	d.params.Set(IDKey, strconv.Itoa(d.ID()))
	data, err := d.client.PutParams(ctx, d.apiPath, d.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(data, d)
}

func (d *dnsResource) ID() int {
	return d.id
}

func (d *dnsResource) FQDN() string {
	return d.fqdn
}

func (d *dnsResource) AddressTTL() int {
	return *d.addressTTL
}

func (d *dnsResource) IPAddresses() []IPAddress {
	return ipStructSliceToInterface(d.ipAddresses, d.client)
}

func (d *dnsResource) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id          int          `json:"id"`
		Fqdn        string       `json:"fqdn"`
		AddressTTL  *int         `json:"address_ttl"`
		IpAddresses []*ipaddress `json:"ip_addresses"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	d.id = des.Id
	d.fqdn = des.Fqdn
	d.addressTTL = des.AddressTTL
	d.ipAddresses = des.IpAddresses

	return nil
}

type dnsResources struct {
	client  *authenticatedClient
	apiPath string
	params  Params
}

func (r *dnsResources) DNSResource(id int) DNSResource {
	d := &dnsResource{}
	return dnsResourceStructToInterface(d, r.client)
}

func (r *dnsResources) WithDomain(domain string) DNSResourceBuilder {
	r.params.Set(DomainKey, domain)
	return r
}

func (r *dnsResources) WithName(name string) DNSResourceBuilder {
	r.params.Set(NameKey, name)
	return r
}

func (r *dnsResources) WithFQDN(fqdn string) DNSResourceBuilder {
	r.params.Add(FQDNKey, fqdn)

	return r
}

func (r *dnsResources) WithAddressTTL(addressTTL string) DNSResourceBuilder {
	r.params.Add(AddressTTLKey, addressTTL)
	return r
}

func (r *dnsResources) WithIPAddresses(ipAddresses []string) DNSResourceBuilder {
	r.params.Add(IPAddressesKey, strings.Join(ipAddresses, " "))
	return r
}

func (r *dnsResources) Create(ctx context.Context) (DNSResource, error) {
	data, err := r.client.Post(ctx, r.apiPath, r.params.Values())
	if err != nil {
		return nil, err
	}

	var obj *dnsResource
	err = unMarshalJson(data, &obj)
	if err != nil {
		return nil, err
	}

	return dnsResourceStructToInterface(obj, r.client), nil
}

func (r *dnsResources) List(ctx context.Context, params Params) ([]DNSResource, error) {
	if params == nil {
		params = ParamsBuilder()
		params.Set(AllKey, strconv.FormatBool(true))
	}

	data, err := r.client.Get(ctx, r.apiPath, params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*dnsResource
	err = unMarshalJson(data, &obj)
	if err != nil {
		return nil, err
	}

	return dnsResourceSliceToInterfaceSlice(obj, r.client), nil
}

func dnsResourceSliceToInterfaceSlice(d []*dnsResource, client Client) []DNSResource {
	var result []DNSResource
	for _, dr := range d {
		result = append(result, dnsResourceStructToInterface(dr, client))
	}

	return result
}

func dnsResourceStructToInterface(d *dnsResource, client Client) DNSResource {
	d.client = client
	d.apiPath = fmt.Sprintf(DNSResourceAPIFormat, d.id)
	d.params = ParamsBuilder()
	return d
}

func (r *dnsResources) Builder() DNSResourceBuilder {
	r.params.Reset()
	return r
}

func NewDNSResourcesClient(client *authenticatedClient) DNSResources {
	return &dnsResources{
		client:  client,
		params:  ParamsBuilder(),
		apiPath: DNSResourcesAPIPath,
	}
}
