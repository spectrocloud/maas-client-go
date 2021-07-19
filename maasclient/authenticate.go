package maasclient

import (
	"errors"
	"fmt"
	"strings"
)

func (c *Client) Authenticate() error {

	elements := strings.Split(c.apiKey, ":")
	if len(elements) != 3 {
		return errors.New(fmt.Sprintf("invalid API key %q; expected \"<consumer secret>:<token key>:<token secret>\"", c.apiKey))
	}

	_, err := c.GetZones()
	if err != nil {
		return err
	}
	return nil
}
