package maasclient

import (
	"context"
	"net/http"
)

type Zone struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ResourceUri string `json:"resource_uri"`
}

func (c *Client) GetZones() ([]Zone, error) {

	var azs []Zone
	err := c.send(context.Background(), http.MethodGet, "/zones/", nil, &azs)

	if err != nil {
		return nil, err
	}
	return azs, nil
}
