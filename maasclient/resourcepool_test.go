package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestResourcePool(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list resourcepools", func(t *testing.T) {
		res, err := c.ResourcePools().List(ctx, nil)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
	})
}
