package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRackControllers(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("begin rack import", func(t *testing.T) {
		err := c.RackControllers().ImportBootImages(ctx)
		assert.Nil(t, err, "expecting nil error")
	})
}
