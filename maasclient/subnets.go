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
)

// Subnets provides access to MAAS subnets (GET /subnets/).
type Subnets interface {
	// List returns all subnets. Use to resolve a CIDR (e.g. 192.168.95.0/24) to subnet ID.
	List(ctx context.Context) ([]Subnet, error)
}

type subnets struct {
	Controller
}

func (s *subnets) List(ctx context.Context) ([]Subnet, error) {
	res, err := s.client.Get(ctx, s.apiPath, s.params.Values())
	if err != nil {
		return nil, err
	}
	var obj []*subnet
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}
	return subnetSliceToInterface(obj), nil
}

// NewSubnetsClient returns a Subnets client.
func NewSubnetsClient(client *authenticatedClient) Subnets {
	return &subnets{
		Controller: Controller{
			client:  client,
			apiPath: "/subnets/",
			params:  ParamsBuilder(),
		},
	}
}
