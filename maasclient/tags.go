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
	"net/url"
	"strings"
)

const (
	TagsAPIPath = "/tags/"
)

type Tags interface {
	List(ctx context.Context) ([]Tag, error)
	// Create creates a new tag with the given name
	Create(ctx context.Context, tagName string) error
	// Assign applies the given tag name to the provided machine system IDs
	Assign(ctx context.Context, tagName string, systemIDs []string) error
	// Unassign removes the given tag name from the provided machine system IDs
	Unassign(ctx context.Context, tagName string, systemIDs []string) error
}

type Tag interface {
	Name() string
	Definition() string
	Comment() string
	KernelOpts() string
	ResourceUri() string
}

type tags struct {
	Controller
}

func (ds *tags) List(ctx context.Context) ([]Tag, error) {
	res, err := ds.client.Get(ctx, ds.apiPath, ds.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*tag
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return tagsStructSliceToInterface(obj, ds.client), nil
}

func (ds *tags) Create(ctx context.Context, tagName string) error {
	if tagName == "" {
		return nil
	}
	params := url.Values{}
	params.Set("name", tagName)
	_, err := ds.client.Post(ctx, ds.apiPath, params)
	return err
}

func (ds *tags) Assign(ctx context.Context, tagName string, systemIDs []string) error {
	if tagName == "" || len(systemIDs) == 0 {
		return nil
	}
	params := url.Values{}
	params.Set("op", "assign")
	params.Set("machines", strings.Join(systemIDs, ","))
	path := fmt.Sprintf("%s%s/", ds.apiPath, url.PathEscape(tagName))
	_, err := ds.client.Post(ctx, path, params)
	return err
}

func (ds *tags) Unassign(ctx context.Context, tagName string, systemIDs []string) error {
	if tagName == "" || len(systemIDs) == 0 {
		return nil
	}
	params := url.Values{}
	params.Set("op", "remove")
	params.Set("machines", strings.Join(systemIDs, ","))
	path := fmt.Sprintf("%s%s/", ds.apiPath, url.PathEscape(tagName))
	_, err := ds.client.Post(ctx, path, params)
	return err
}

func tagsStructSliceToInterface(in []*tag, client Client) []Tag {
	var out []Tag
	for _, d := range in {
		out = append(out, tagStructToInterface(d, client))
	}
	return out
}

func tagStructToInterface(in *tag, client Client) Tag {
	in.client = client
	in.apiPath = TagsAPIPath
	in.params = ParamsBuilder()
	return in
}

type tag struct {
	name         string
	definition   string
	comment      string
	kernel_opts  string
	resource_uri string
	Controller
}

func (d *tag) Name() string {
	return d.name
}

func (d *tag) Definition() string {
	return d.definition
}

func (d *tag) Comment() string {
	return d.comment
}

func (d *tag) KernelOpts() string {
	return d.kernel_opts
}

func (d *tag) ResourceUri() string {
	return d.resource_uri
}

func (d *tag) UnmarshalJSON(data []byte) error {
	des := struct {
		ResourceUri string `json:"resource_uri"`
		Name        string `json:"name"`
		Definition  string `json:"definition"`
		Comment     string `json:"comment"`
		KernelOpts  string `json:"kernel_opts"`
	}{}

	err := json.Unmarshal(data, &des)
	if err != nil {
		return err
	}

	d.name = des.Name
	d.comment = des.KernelOpts
	d.resource_uri = des.ResourceUri
	d.kernel_opts = des.KernelOpts
	d.definition = des.Definition

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
