package maasclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type BootResource struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Architecture string `json:"architecture"`
	SubArches string `json:"subarches"`
	ResourceURI string `json:"resource_uri"`
	Title string `json:"title"`
	Sets map[string]Set `json:"sets"`
}


type Set struct {
	Version string `json:"version"`
	Label string `json:"label"`
	Size int `json:"size"`
	Complete bool `json:"complete"`
	Progress int `json:"progress"`
	Files map[string]File `json:"files"`
}

type File struct {
	FileName string `json:"filename"`
	FileType string `json:"filetype"`
	SHA256 string `json:"sha256"`
	Size int `json:"size"`
	Complete bool `json:"complete"`
	Progress int `json:"progress"`
	UploadURI string `json:"upload_uri"`
}

func (b *BootResource) LatestSet() (*Set, error) {
	if len(b.Sets) == 0 {
		return nil, errors.New("no set found in bootresource")
	}

	keys := make([]string, 0)
	for key := range b.Sets {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	// return first element
	set := b.Sets[keys[0]]
	return &set, nil
}

func (s *Set) GetUploadURI() (string, error) {

	if len(s.Files) != 1 {
		return "", errors.New("multiple files present in set")
	}
	for _, file := range s.Files {
		return strings.TrimPrefix(file.UploadURI, "/MAAS/api/2.0"), nil
	}
	return "", nil
}

func (c *Client) ListBootResources(ctx context.Context) ([]*BootResource, error) {
	q := url.Values{}
	var res []*BootResource
	if err := c.send(ctx, http.MethodGet, "/boot-resources/", q, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) BootResourcesImporting(ctx context.Context) (*bool, error) {
	q := url.Values{}
	q.Add("op", "is_importing")
	var res *bool
	if err := c.send(ctx, http.MethodGet, "/boot-resources/", q, &res); err != nil {
		return nil, err
	}
	return res, nil
}

type UploadBootResourceInput struct {
	Name string
	Architecture string
	Digest string
	Size string
	Title string
	File string
}

func (c *Client) UploadBootResource(ctx context.Context, input UploadBootResourceInput) (*BootResource, error) {
	q := url.Values{}

	q.Add("name", input.Name)
	q.Add("architecture", input.Architecture)
	q.Add("sha256", input.Digest)
	q.Add("size", input.Size)
	q.Add("title", input.Title)


	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	err := writeMultiPartParams(writer, q)
	if err != nil {
		return nil, err
	}
	writer.Close()

	var res *BootResource
	if err := c.sendRequestWithBody(ctx, http.MethodPost, "/boot-resources/", writer.FormDataContentType(),q, buf, &res); err != nil {
		return nil, err
	}

	lset, err := res.LatestSet()
	if err != nil {
		return nil, err
	}

	uri, err := lset.GetUploadURI()
	if err != nil {
		return nil, err
	}


	f, err := os.Open(input.File)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	fileBuf := make([]byte, 1 << 22)

	for {
		n, err := reader.Read(fileBuf)

		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		if n < (1 << 22) {
			fileBufEnd := make([]byte, n)
			copy(fileBufEnd, fileBuf)
			var res *string
			if err := c.sendRequestPutWithBody(ctx, http.MethodPut, uri, q, fileBufEnd, n, &res); err != nil {
				return nil, err
			}
			break
		}

		var res *string
		if err := c.sendRequestPutWithBody(ctx, http.MethodPut, uri, q, fileBuf, n, &res); err != nil {
			return nil, err
		}
	}
	return res, nil
}


func (c *Client) sendRequestPutWithBody(ctx context.Context, method string, apiPath string, params url.Values, fileBuffer []byte, size int, v interface{}) error {

	var err error
	var req *http.Request

	req, err = http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.baseURL, apiPath),
		bytes.NewBuffer(fileBuffer),
	)
	if err != nil {
		return err
	}

	return c.sendRequestUploadPut(req, params, len(fileBuffer), v)
}

func (c *Client) sendRequestUploadPut(req *http.Request, params url.Values, size int, v interface{}) error {
	//func (c *Client) sendRequest(req *http.Request, urlValues *url.Values, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/octet-stream")

	// for post requests longer than 300 seconds
	ticker := time.NewTicker(2 * time.Minute)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("refresing auth token", time.Unix(t.Unix(), 0).Format(time.RFC3339))
				authHeader := authHeader(req, params, c.apiKey)
				req.Header.Set("Authorization", authHeader)
			}
		}
	}()


	defer func() {
		ticker.Stop()
		done <- true
	}()

	req.ContentLength = int64(size)

	authHeader := authHeader(req, params, c.apiKey)
	req.Header.Set("Authorization", authHeader)


	res, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	defer res.Body.Close()

	// Try to unmarshall into errorResponse
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unknown error, status code: %d, body: %s", res.StatusCode, string(bodyBytes))
	} else if res.StatusCode == http.StatusNoContent {
		return nil
	}

	responseString, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if strings.ToUpper(string(responseString)) != "OK" {
		fmt.Println("expected output to be", string(responseString), res.StatusCode)
		return err
	}

	return nil
}



func writeMultiPartParams(writer *multipart.Writer, params url.Values) error {
	for key, values := range params {
		for _, value := range values {
			fw, err := writer.CreateFormField(key)
			if err != nil {
				return err
			}
			buffer := bytes.NewBufferString(value)
			io.Copy(fw, buffer)
		}
	}
	return nil

}

func (c *Client) GetBootResource(ctx context.Context, id string) (*BootResource, error) {
	q := url.Values{}
	var res *BootResource
	if err := c.send(ctx, http.MethodGet, fmt.Sprintf("/boot-resources/%s/", id), q, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) DeleteBootResource(ctx context.Context, id string) error {
	q := url.Values{}
	var res interface{}
	if err := c.send(ctx, http.MethodDelete, fmt.Sprintf("/boot-resources/%s/", id), q, &res); err != nil {
		return err
	}
	return nil
}

