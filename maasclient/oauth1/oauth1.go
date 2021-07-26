package oauth1

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	guuid "github.com/google/uuid"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OAuth struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

// TODO add comment on-top about what is changed, and the header

func NewOAuth(consumerKey, consumerSecret, accessToken, accessSecret string) *OAuth {
	return &OAuth{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    accessToken,
		AccessSecret:   accessSecret,
	}
}

// Params being any key-value url query parameter pairs
func (auth OAuth) BuildOAuthHeader(method, path string, params map[string]string) string {
	vals := url.Values{}
	vals.Add("oauth_nonce", guuid.NewString())
	vals.Add("oauth_consumer_key", auth.ConsumerKey)
	vals.Add("oauth_signature_method", "HMAC-SHA1")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", auth.AccessToken)
	vals.Add("oauth_version", "1.0")

	for k, v := range params {
		vals.Add(k, v)
	}
	// net/url package QueryEscape escapes " " into "+", this replaces it with the percentage encoding of " "
	parameterString := strings.Replace(vals.Encode(), "+", "%20", -1)

	// Calculating Signature Base String and Signing Key
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(strings.Split(path, "?")[0]) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(auth.ConsumerSecret) + "&" + url.QueryEscape(auth.AccessSecret)
	signature := calculateSignature(signatureBase, signingKey)

	vals.Add("oauth_signature", signature)

	var authHeader []string
	for key := range vals {
		authHeader = append(authHeader, fmt.Sprintf(`%s="%s"`, key, url.QueryEscape(vals.Get(key))))
	}
	return "OAuth " + strings.Join(authHeader, ", ")
}

func calculateSignature(base, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}
