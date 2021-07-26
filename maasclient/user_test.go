package maasclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestUsers(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list users", func(t *testing.T) {
		res, err := c.Users().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
	})

	t.Run("whoami", func(t *testing.T) {
		res, err := c.Users().WhoAmI(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
		assert.Equal(t, res.UserName(), "dev")
	})

}
