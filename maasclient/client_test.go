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
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthHeader_ReturnsEmptyForInvalidKey(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/2.0/machines/", nil)
	result := authHeader(req, url.Values{}, "invalid-key")
	assert.Empty(t, result)
}

func TestAuthHeader_ValidOAuthFormat(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/2.0/machines/", nil)
	result := authHeader(req, url.Values{"op": []string{"allocate"}}, "consumer:token:secret")
	assert.True(t, strings.HasPrefix(result, "OAuth "), "auth header should start with 'OAuth '")
}

func TestAuthHeader_MultiValuedParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/api/2.0/machines/", nil)
	params := url.Values{}
	params.Add("tags", "prod")
	params.Add("tags", "virtual")

	assert.NotPanics(t, func() {
		result := authHeader(req, params, "consumer:token:secret")
		assert.True(t, strings.HasPrefix(result, "OAuth "), "auth header should be valid with multi-valued params")
	})
}

func TestAuthHeader_PutMethodIgnoresParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPut, "http://example.com/api/2.0/machines/abc/", nil)
	params := url.Values{}
	params.Add("tags", "tag1")
	params.Add("tags", "tag2")

	result := authHeader(req, params, "consumer:token:secret")
	assert.True(t, strings.HasPrefix(result, "OAuth "), "PUT should produce valid OAuth header")
}
