/*
Copyright 2021 Spectrocloud

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
	"time"
)

const (
	DomainsAPIPath      = "/domains/"
	DomainAPIPathFormat = "/domains/%d"
)

type Domains interface {
	List(ctx context.Context) ([]Domain, error)
	Domain(id int) Domain
}

type Domain interface {
	ID() int
	IsAuthoritative() bool
	TTL() time.Duration
	IsDefault() bool
	Name() string
	ResourceRecordCount() int
}

type domains struct {
	Controller
}

func (ds *domains) Domain(id int) Domain {
	return domainStructToInterface(&domain{
		id: id,
	}, ds.client)
}

func (ds *domains) List(ctx context.Context) ([]Domain, error) {
	res, err := ds.client.Get(ctx, ds.apiPath, ds.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*domain
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return domainStructSliceToInterface(obj, ds.client), nil
}

func domainStructSliceToInterface(in []*domain, client Client) []Domain {
	var out []Domain
	for _, d := range in {
		out = append(out, domainStructToInterface(d, client))
	}
	return out
}

func domainStructToInterface(in *domain, client Client) Domain {
	in.client = client
	in.apiPath = fmt.Sprintf(DomainAPIPathFormat, in.id)
	in.params = ParamsBuilder()
	return in
}

type domain struct {
	id                  int
	isAuthoritative     bool
	ttl                 int
	isDefault           bool
	name                string
	resourceRecordCount int
	Controller
}

func (d *domain) ID() int {
	return d.id
}

func (d *domain) IsAuthoritative() bool {
	return d.isAuthoritative
}

func (d *domain) TTL() time.Duration {
	return time.Duration(d.ttl) * time.Second
}

func (d *domain) IsDefault() bool {
	return d.isDefault
}

func (d *domain) Name() string {
	return d.name
}

func (d *domain) ResourceRecordCount() int {
	return d.resourceRecordCount
}

func (d *domain) UnmarshalJSON(data []byte) error {
	des := &struct {
		Authoritative       bool   `json:"authoritative"`
		TTL                 int    `json:"ttl"`
		ResourceRecordCount int    `json:"resource_record_count"`
		Name                string `json:"name"`
		Id                  int    `json:"id"`
		IsDefault           bool   `json:"is_default"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	d.id = des.Id
	d.name = des.Name
	d.ttl = des.TTL
	d.isAuthoritative = des.Authoritative
	d.isDefault = des.IsDefault
	d.resourceRecordCount = des.ResourceRecordCount

	return nil
}

func NewDomainsClient(client *authenticatedClient) Domains {
	return &domains{
		Controller: Controller{
			client:  client,
			apiPath: DomainsAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
