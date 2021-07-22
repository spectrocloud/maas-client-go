package maasclient_test

import (
	"github.com/spectrocloud/maas-client-go/maasclient"
	"os"
	"testing"
)

const MAAS_ENDPOINT = "MAAS_ENDPOINT"
const MAAS_APIKEY = "MAAS_API_KEY"

func TestMaasAuthenticate(t *testing.T) {

	client := maasclient.NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))
	err := client.Authenticate()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("Authenticate %#v", err)

}