package maasclient

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) RackControllerBootImageImport(ctx context.Context) error {
	q := url.Values{}
	q.Add("op", "import_boot_images")
	res := ""
	if err := c.sendTextBodyResponse(ctx, http.MethodPost, "/rackcontrollers/", q, &res); err != nil {
		return err
	}
	return nil
}
