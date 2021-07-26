package maasclient

import (
	"context"
	"encoding/json"
)

const (
	ZonesAPIPath = "/zones/"
)

type Zones interface {
	List(ctx context.Context) ([]Zone, error)
}

type Zone interface {
	ID() int
	Name() string
	Description() string
}

type zones struct {
	Controller
}

func (z *zones) List(ctx context.Context) ([]Zone, error) {
	res, err := z.client.Get(ctx, z.apiPath, z.params.Values())
	if err != nil {
		return nil, err
	}

	var out []*zone
	err = unMarshalJson(res, &out)
	if err != nil {
		return nil, err
	}

	return zoneStructSliceToInterface(out, z.client), err
}

func zoneStructSliceToInterface(in []*zone, client Client) []Zone {
	var out []Zone
	for _, z := range in {
		out = append(out, zoneStructToInterface(z, client))
	}
	return out
}

func zoneStructToInterface(in *zone, client Client) Zone {
	return in
}

type zone struct {
	id          int
	name        string
	description string
}

func (z *zone) ID() int {
	return z.id
}

func (z *zone) Name() string {
	return z.name
}

func (z *zone) Description() string {
	return z.description
}

func (z *zone) UnmarshalJSON(data []byte) error {
	des := &struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	z.id = des.ID
	z.name = des.Name
	z.description = des.Description

	return nil
}

func NewZonesClient(client Client) Zones {
	return &zones{
		Controller: Controller{
			client:  client,
			apiPath: ZonesAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
