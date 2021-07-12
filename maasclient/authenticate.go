package maasclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) Authenticate() error {

	elements := strings.Split(c.apiKey, ":")
	if len(elements) != 3 {
		return errors.New(fmt.Sprintf("invalid API key %q; expected \"<consumer secret>:<token key>:<token secret>\"", c.apiKey))
	}

	err := c.ifUserExist()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ifUserExist() error {

	var parsed interface{}
	err := c.Send(context.Background(), http.MethodGet, "/users/?op=whoami", nil, &parsed)

	if err != nil {
		return err
	}
	return nil
}
