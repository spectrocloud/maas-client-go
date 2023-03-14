# Go-OAuth1.0
Golang lightweight package/ implementation example of OAuth1.0 Authentication Header/ Signature calculation (Twitter etc..)

To quickly import the package into your project:
> ```
>  go get github.com/klaidas/go-oauth1
> ```

&nbsp;

Example usage: 
```Go
package main

import (
	"fmt"
	oauth1 "go-oauth1"
	"net/http"
)

func main() {
	method := http.MethodPost
	url := "https://api.twitter.com/1.1/statuses/update.json?include_entities=true"
	
	auth := oauth1.OAuth1{
		ConsumerKey: "<consumer-key>",
		ConsumerSecret: "<consumer-secret>",
		AccessToken: "<access-token>",
		AccessSecret: "<access-secret>",
	}

	authHeader := auth.BuildOAuth1Header(method, url, map[string]string {
		"include_entities": "true",
		"status": "Hello Ladies + Gentlemen, a signed OAuth request!",
	})
	
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", authHeader)
	
	if res, err := http.DefaultClient.Do(req); err == nil {
		fmt.Println(res.StatusCode)
	}
}
```

&nbsp;

- Simply import the package
- Create an OAuth1 Object with the information you have (In some cases, AccessSecret will be unknown, this is fine)
- Call BuildOAuth1Header to generate your Authorization Header for the request you're making

&nbsp;

Output: 
```
OAuth oauth_consumer_key="<oauth-consumer-key>",oauth_token="<oauth-token>",oauth_signature_method="HMAC-SHA1",oauth_timestamp="1318622958",oauth_nonce="<oauth-nonce>",oauth_version="1.0",oauth_signature="<oauth-signature>"
```
