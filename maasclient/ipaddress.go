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
	"encoding/json"
	"net"
)

type IPAddress interface {
	IP() net.IP
}

type ipaddress struct {
	ip net.IP
}

func (i *ipaddress) IP() net.IP {
	return i.ip
}

func (i *ipaddress) UnmarshalJSON(data []byte) error {
	des := &struct {
		Ip string `json:"ip"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	i.ip = net.ParseIP(des.Ip)

	return nil
}

func ipStructToInterface(in *ipaddress, client Client) IPAddress {
	return in
}

func ipStructSliceToInterface(in []*ipaddress, client Client) []IPAddress {
	var out []IPAddress
	for _, ip := range in {
		out = append(out, ipStructToInterface(ip, client))
	}
	return out
}
