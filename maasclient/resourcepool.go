package maasclient

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	ResourcePoolsAPIPath      = "/resourcepools/"
	ResourcePoolAPIPathFormat = "/resourcepools/%d"
)

type ResourcePools interface {
	List(ctx context.Context, params Params) ([]ResourcePool, error)
	ResourcePool(id int) ResourcePool
}

type ResourcePool interface {
	Name() string
	Description() string
	ID() int
}

type resourcePools struct {
	Controller
}

func (rps *resourcePools) List(ctx context.Context, params Params) ([]ResourcePool, error) {
	res, err := rps.client.Get(ctx, rps.apiPath, rps.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*resourcePool
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return resourcePoolStructSliceToInterface(obj, rps.client), nil
}

func resourcePoolStructSliceToInterface(in []*resourcePool, client Client) []ResourcePool {
	var out []ResourcePool
	for _, pool := range in {
		out = append(out, resourcePoolStructToInterface(pool, client))
	}
	return out
}

func resourcePoolStructToInterface(in *resourcePool, client Client) ResourcePool {
	in.client = client
	in.apiPath = fmt.Sprintf(ResourcePoolAPIPathFormat, in.ID())
	return in
}

func (rps *resourcePools) ResourcePool(id int) ResourcePool {
	return resourcePoolStructToInterface(&resourcePool{id: id}, rps.client)
}

type resourcePool struct {
	name        string
	id          int
	description string

	Controller
}

func (rp *resourcePool) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	rp.id = des.Id
	rp.name = des.Name
	rp.description = des.Description

	return nil
}

func (rp *resourcePool) Name() string {
	return rp.name
}

func (rp *resourcePool) Description() string {
	return rp.description
}

func (rp *resourcePool) ID() int {
	return rp.id
}

func NewResourcePoolsClient(client *authenticatedClient) ResourcePools {
	return &resourcePools{
		Controller: Controller{
			client:  client,
			apiPath: ResourcePoolsAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
