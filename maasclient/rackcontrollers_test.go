package maasclient_test

import (
	"context"
	. "github.com/spectrocloud/maas-client-go/maasclient"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRackControllers(t *testing.T) {
	c := NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))

	ctx := context.Background()

	t.Run("begin rack import", func(t *testing.T) {
		err := c.RackControllerBootImageImport(ctx)
		assert.Nil(t, err, "expecting nil error")
	})
}
