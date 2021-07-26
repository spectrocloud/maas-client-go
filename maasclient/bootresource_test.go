package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetBootResources(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list-all", func(t *testing.T) {
		list, err := c.BootResources().List(ctx, nil)
		assert.Nil(t, err, "expecting nil error")
		assert.NotEmpty(t, list)
	})
	//
	t.Run("list-by-id", func(t *testing.T) {
		res := c.BootResources().BootResource(7)
		err := res.Get(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	t.Run("import image", func(t *testing.T) {

		res, err := c.BootResources().Builder("test-image",
			"amd64/generic",
			"e9844638c7345d182c5d88e1eaeae74749d02beeca38587a530207fddc0a280a",
			"/Users/deepak/maas/ubuntu.tar.gz", 1262032476).Create(ctx)
		assert.Nil(t, err)
		err = res.Upload(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})
}
