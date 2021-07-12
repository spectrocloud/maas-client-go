package maasclient

import (
	"context"
	"net/http"
)

type Domain struct {
	ResourceUri string `json:"resourceURI"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) GetDomain() ([]Domain, error) {

	var domains []Domain
	err := c.send(context.Background(), http.MethodGet, "/domains/", nil, &domains)

	if err != nil {
		return nil, err
	}
	return domains, nil
}


