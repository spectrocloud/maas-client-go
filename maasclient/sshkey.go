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
	SSHKeysAPIPath = "/account/prefs/sshkeys/"
)

type SSHKeys interface {
	List(ctx context.Context) ([]SSHKey, error)
}

type SSHKey interface {
	ID() int
	Key() string
	KeySource() string
}

type sshKeys struct {
	Controller
}

func (s *sshKeys) List(ctx context.Context) ([]SSHKey, error) {
	res, err := s.client.Get(ctx, s.apiPath, s.params.Values())
	if err != nil {
		return nil, err
	}

	var out []*sshKey
	err = unMarshalJson(res, &out)
	if err != nil {
		return nil, err
	}

	return sshKeyStructSliceToInterface(out, s.client), err
}

func sshKeyStructSliceToInterface(in []*sshKey, client Client) []SSHKey {
	var out []SSHKey
	for _, s := range in {
		out = append(out, sshKeyStructToInterface(s, client))
	}
	return out
}

func sshKeyStructToInterface(in *sshKey, client Client) SSHKey {
	return in
}

type sshKey struct {
	id        int
	key       string
	keySource string
}

func (s *sshKey) ID() int {
	return s.id
}

func (s *sshKey) Key() string {
	return s.key
}

func (s *sshKey) KeySource() string {
	return s.keySource
}

func (s *sshKey) UnmarshalJSON(data []byte) error {
	des := &struct {
		ID        int    `json:"id"`
		Key       string `json:"key"`
		KeySource string `json:"keySource"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	s.id = des.ID
	s.key = des.Key
	s.keySource = des.KeySource

	return nil
}

func NewSSHKeysClient(client Client) SSHKeys {
	return &sshKeys{
		Controller: Controller{
			client:  client,
			apiPath: SSHKeysAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
