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
)

const (
	TagsAPIPath       = "/tags/"
	TagsAPIPathFormat = "/tags/%d"
)

type Tags interface {
	List(ctx context.Context) ([]Domain, error)
}

type Tag interface {
	ID() int
	Name() string
}

type tags struct {
	Controller
}

func (ds *tags) List(ctx context.Context) ([]Domain, error) {
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

type tag struct {
	id   int
	name string
	Controller
}

func (d *tag) ID() int {
	return d.id
}

func (d *tag) Name() string {
	return d.name
}

func (d *tag) UnmarshalJSON(data []byte) error {
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

	return nil
}

func NewTagsClient(client *authenticatedClient) Tags {
	return &tags{
		Controller: Controller{
			client:  client,
			apiPath: TagsAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
