/*
Copyright 2021 Spectro Cloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package maasclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Controller struct {
	client  Client
	apiPath string
	params  Params
}

func unMarshalJson(res *http.Response, v interface{}) error {
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// 200 json body unmarshal ok
	// cases where response in json caller can send nil and avoid unmarshalling overall
	// for > 300 errors we return the body as it is
	switch {
	case res.StatusCode == http.StatusNoContent:
		return nil
	case statusAcceptable(res.StatusCode):
		if v != nil {
			return json.Unmarshal(bodyBytes, v)
		}
		return nil
	case res.StatusCode >= 300:
		return fmt.Errorf("status: %d, message: %s", res.StatusCode, bodyBytes)
	}
	return fmt.Errorf("status: %d, message: %s", res.StatusCode, bodyBytes)
}

func statusAcceptable(status int) bool {
	return status == http.StatusOK ||
		status == http.StatusCreated ||
		status == http.StatusAccepted
}
