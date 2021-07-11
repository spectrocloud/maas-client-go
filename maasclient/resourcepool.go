package maasclient

import (
	"context"
	"net/http"
)

type ResourcePool struct {
	ResourceUri string `json:"resourceURI"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) GetPools(client Client) ([]ResourcePool, error) {

	var pools []ResourcePool
	err := client.Send(context.Background(), http.MethodGet, "/resourcepools/", nil, &pools)

	if err != nil {
		return nil, err
	}
	return pools, nil
}
