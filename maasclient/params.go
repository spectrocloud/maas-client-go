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

import "net/url"

type params struct {
	values url.Values
}

func ParamsBuilder() Params {
	return &params{values: url.Values{}}
}

func (p *params) Add(key, value string) Params {
	p.values.Add(key, value)
	return p
}

func (p *params) Set(key, value string) Params {
	p.values.Set(key, value)
	return p
}

func (p *params) Reset() {
	p.values = url.Values{}
}

func (p *params) Values() url.Values {
	return p.values
}

func (p *params) Copy(in Params) {
	for key, value := range in.Values() {
		p.values[key] = value
	}
}

type Params interface {
	Add(key, value string) Params
	Set(key, value string) Params
	Reset()
	Values() url.Values
	Copy(in Params)
}
