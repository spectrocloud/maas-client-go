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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	BootResourcesAPIPath      = "/boot-resources/"
	BootResourceAPIPathFormat = "/boot-resources/%d/"
)

// Implements subset of https://maas.io/docs/api#boot-resources
// Usage bootstrap_test.go
type BootResources interface {
	List(ctx context.Context, params Params) ([]BootResource, error)
	BootResource(id int) BootResource
	Builder(name, architecture, hash, filePath string, size int) BootResourceBuilder
}

type BootResource interface {
	BootResourceUploader
	Get(ctx context.Context) (BootResource, error)
	Delete(ctx context.Context) error
	ID() int
	Type() string
	Name() string
	Architecture() string
	SubArches() string
	Title() string
	Sets() map[string]Set
}

type BootResourceBuilder interface {
	WithTitle(title string) BootResourceBuilder
	WithFileType(fileType string) BootResourceBuilder
	Create(ctx context.Context) (BootResource, error)
}

type BootResourceUploader interface {
	Upload(ctx context.Context) error
}

type Set struct {
	Version  string             `json:"version"`
	Label    string             `json:"label"`
	Size     int                `json:"size"`
	Complete bool               `json:"complete"`
	Progress int                `json:"progress"`
	Files    map[string]SetFile `json:"files"`
}

type SetFile struct {
	FileName  string `json:"filename"`
	FileType  string `json:"filetype"`
	SHA256    string `json:"sha256"`
	Size      int    `json:"size"`
	Complete  bool   `json:"complete"`
	Progress  int    `json:"progress"`
	UploadURI string `json:"upload_uri"`
}

func (s *Set) getUploadURI() (string, error) {
	if len(s.Files) != 1 {
		return "", errors.New("multiple files present in set")
	}
	for _, file := range s.Files {
		return strings.TrimPrefix(file.UploadURI, "/MAAS/api/2.0"), nil
	}
	return "", nil
}

type bootResources struct {
	Controller
	filePath string
}

func (brs *bootResources) BootResource(id int) BootResource {
	return bootResourceStructToInterface(&bootResource{
		id: id,
	}, brs.client)
}

func (brs *bootResources) WithTitle(title string) BootResourceBuilder {
	brs.params.Set(TitleKey, title)
	return brs
}

func (brs *bootResources) WithFileType(fileType string) BootResourceBuilder {
	brs.params.Set(FileTypeKey, fileType)
	return brs
}

func (brs *bootResources) Create(ctx context.Context) (BootResource, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	err := writeMultiPartParams(writer, brs.params.Values())
	if err != nil {
		return nil, err
	}
	writer.Close()

	res, err := brs.client.PostForm(ctx, brs.apiPath, writer.FormDataContentType(), brs.params.Values(), buf)
	if err != nil {
		return nil, err
	}

	var obj *bootResource
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	bootResourceStructToInterface(obj, brs.client)
	obj.filePath = brs.filePath

	return obj, nil
}

func (brs *bootResources) List(ctx context.Context, params Params) ([]BootResource, error) {
	res, err := brs.client.Get(ctx, brs.apiPath, brs.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*bootResource
	err = unMarshalJson(res, &obj)
	if err != nil {
		return nil, err
	}

	return bootResourceStructSliceToInterface(obj, brs.client), nil
}

func bootResourceStructSliceToInterface(in []*bootResource, client Client) []BootResource {
	var out []BootResource
	for _, br := range in {
		out = append(out, bootResourceStructToInterface(br, client))
	}
	return out
}

func bootResourceStructToInterface(in *bootResource, client Client) BootResource {
	in.client = client
	in.apiPath = fmt.Sprintf(BootResourceAPIPathFormat, in.id)
	in.params = ParamsBuilder()
	return in
}

func (brs *bootResources) Builder(name, architecture, hash, filePath string, size int) BootResourceBuilder {
	brs.params.Reset()
	brs.params.Set(NameKey, name)
	brs.params.Set(ArchitectureKey, architecture)
	brs.params.Set(SHA256Key, hash)
	brs.params.Set(SizeKey, strconv.Itoa(size))

	brs.filePath = filePath

	return brs
}

type bootResource struct {
	id               int
	name             string
	bootResourceType string
	architecture     string
	subarches        string
	title            string
	sets             map[string]Set

	filePath string
	Controller
}

func (b *bootResource) Upload(ctx context.Context) error {
	lset, err := b.LatestSet()
	if err != nil {
		return err
	}

	uri, err := lset.getUploadURI()
	if err != nil {
		return err
	}

	f, err := os.Open(b.filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	fileBuf := make([]byte, 1<<22)

	for {
		n, err := reader.Read(fileBuf)

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if n < (1 << 22) {
			fileBufEnd := make([]byte, n)
			copy(fileBufEnd, fileBuf)

			if err := b.uploadBuffer(ctx, uri, fileBufEnd, n); err != nil {
				return err
			}
			break
		}

		if err := b.uploadBuffer(ctx, uri, fileBuf, n); err != nil {
			return err
		}
	}
	return nil
}

func (b *bootResource) uploadBuffer(ctx context.Context, uri string, fileBuf []byte, contentLength int) error {
	res, err := b.client.Put(ctx, uri, b.params.Values(), bytes.NewReader(fileBuf), contentLength)
	if err != nil {
		return err
	}

	return unMarshalJson(res, nil)
}

func (b *bootResource) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id           int            `json:"id"`
		Type         string         `json:"type"`
		Name         string         `json:"name"`
		Architecture string         `json:"architecture"`
		SubArches    string         `json:"subarches"`
		Title        string         `json:"title"`
		Sets         map[string]Set `json:"sets"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	b.id = des.Id
	b.bootResourceType = des.Type
	b.name = des.Name
	b.architecture = des.Architecture
	b.subarches = des.SubArches
	b.title = des.Title
	b.sets = des.Sets

	return nil
}

func (b *bootResource) Delete(ctx context.Context) error {
	res, err := b.client.Delete(ctx, b.apiPath, nil)
	if err != nil {
		return err
	}

	return unMarshalJson(res, &b)
}

func (b *bootResource) Get(ctx context.Context) (BootResource, error) {
	res, err := b.client.Get(ctx, b.apiPath, b.params.Values())
	if err != nil {
		return nil, err
	}

	return b, unMarshalJson(res, &b)
}

func (b *bootResource) Type() string {
	return b.bootResourceType
}

func (b *bootResource) Name() string {
	return b.name
}

func (b *bootResource) Architecture() string {
	return b.architecture
}

func (b *bootResource) SubArches() string {
	return b.subarches
}

func (b *bootResource) Title() string {
	return b.title
}

func (b *bootResource) Sets() map[string]Set {
	return b.sets
}

func (b *bootResource) ID() int {
	return b.id
}

func (b *bootResource) LatestSet() (*Set, error) {
	if len(b.Sets()) == 0 {
		return nil, errors.New("no set found in bootresource")
	}

	keys := make([]string, 0)
	for key := range b.Sets() {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	// return first element
	set := b.Sets()[keys[0]]
	return &set, nil
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

func NewBootResourcesClient(client Client) BootResources {
	return &bootResources{
		Controller: Controller{
			client:  client,
			apiPath: BootResourcesAPIPath,
			params:  ParamsBuilder(),
		},
	}
}
