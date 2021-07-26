package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDomain(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list domains", func(t *testing.T) {
		res, err := c.Domains().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})
}
