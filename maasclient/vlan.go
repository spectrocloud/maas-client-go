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

type VLAN interface {
	VID() int
	MTU() int
	IsDHCPOn() bool
	FabricID() int
	FabricName() string
}

type vLAN struct {
	vid        int
	mtu        int
	dhcpOn     bool
	fabricId   int
	fabricName string
	id         int
	name       string
}

func (v *vLAN) VID() int {
	return v.vid
}

func (v *vLAN) MTU() int {
	return v.mtu
}

func (v *vLAN) IsDHCPOn() bool {
	return v.dhcpOn
}

func (v *vLAN) FabricID() int {
	return v.fabricId
}

func (v *vLAN) FabricName() string {
	return v.fabricName
}

func (v *vLAN) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id       int    `json:"id"`
		VId      int    `json:"vid"`
		Name     string `json:"name"`
		Fabric   string `json:"fabric"`
		FabricID int    `json:"fabric_id"`
		MTU      int    `json:"mtu"`
		DHCPOn   bool   `json:"dhcp_on"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	v.id = des.Id
	v.vid = des.VId
	v.name = des.Name
	v.fabricId = des.FabricID
	v.fabricName = des.Fabric
	v.mtu = des.MTU
	v.dhcpOn = des.DHCPOn

	return nil
}
