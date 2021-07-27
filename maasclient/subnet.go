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

import "encoding/json"

type Subnet interface {
	ID() int
	Name() string
	Space() string
	VLAN() VLAN
}

type subnet struct {
	id    int
	name  string
	space string
	vlan  *vLAN
}

func (s *subnet) ID() int {
	return s.id
}

func (s *subnet) Name() string {
	return s.name
}

func (s *subnet) Space() string {
	return s.space
}

func (s *subnet) VLAN() VLAN {
	return s.vlan
}

func (s *subnet) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Space string `json:"space"`
		Vlan  *vLAN  `json:"vlan"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	s.id = des.Id
	s.name = des.Name
	s.space = des.Space
	s.vlan = des.Vlan

	return nil
}

func subnetSliceToInterface(in []*subnet) []Subnet {
	var out []Subnet
	for _, s := range in {
		out = append(out, s)
	}
	return out
}
