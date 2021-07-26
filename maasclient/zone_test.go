package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestZones(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list zones", func(t *testing.T) {
		zones, err := c.Zones().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, zones)
		assert.NotEmpty(t, zones)
	})
}
