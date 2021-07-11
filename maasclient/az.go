package maasclient

import (
	"context"
	"net/http"
)

type Zone struct {
	ResourceUri string `json:"resourceURI"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) GetZones(client Client) ([]Zone, error) {

	var azs []Zone
	err := client.Send(context.Background(), http.MethodGet, "/zones/", nil, &azs)

	if err != nil {
		return nil, err
	}
	return azs, nil
}


