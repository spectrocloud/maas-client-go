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
