package maasclient

import (
	"context"
	"net/http"
)

type Space struct {
	ResourceUri string   `json:"resourceURI"`
	Name        string   `json:"name"`
	Subnets     []Subnet `json:"subnets"`
}

type Subnet struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	ResourceUri string `json:"resourceURI"`
	Space       string `json:"space"`
	Vlan       Vlan   `json:"vlan"`
}

type Vlan struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	ResourceUri string `json:"resourceURI"`
	Fabric      string `json:"fabric"`
}

func (c *Client) GetSubnets() ([]Space, error) {

	var spaces []Space
	err := c.Send(context.Background(), http.MethodGet, "/spaces/", nil, &spaces)

	if err != nil {
		return nil, err
	}
	return spaces, nil
}
