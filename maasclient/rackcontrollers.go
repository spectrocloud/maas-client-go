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

const (
	RackControllersAPIPath = "/rackcontrollers/"
)

type RackControllers interface {
	ImportBootImages(ctx context.Context) error
}

type rackControllers struct {
	Controller
}

func NewRackControllersClient(client *authenticatedClient) RackControllers {
	return &rackControllers{
		Controller: Controller{
			client:  client,
			apiPath: RackControllersAPIPath,
			params:  ParamsBuilder(),
		},
	}
}

func (r *rackControllers) ImportBootImages(ctx context.Context) error {
	r.params.Reset()
	r.params.Set(Operation, OperationImportBootImages)

	data, err := r.client.Post(ctx, r.apiPath, r.params.Values())
	if err != nil {
		return err
	}

	return unMarshalJson(data, nil)
}
