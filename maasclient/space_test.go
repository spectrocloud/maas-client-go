package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSpaces(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("space list", func(t *testing.T) {
		res, err := c.Spaces().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
	})
}
