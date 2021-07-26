package maasclient

import (
	"context"
	"encoding/json"
)

const (
	UsersAPIPath = "/users/"
)

type Users interface {
	List(ctx context.Context) ([]User, error)
	WhoAmI(ctx context.Context) (User, error)
}

type User interface {
	IsSuperUser() bool
	UserName() string
	IsLocal() bool
	Email() string
}

type user struct {
	superuser bool
	local     bool
	username  string
	email     string
}

func (u *user) IsSuperUser() bool {
	return u.superuser
}

func (u *user) UserName() string {
	return u.username
}

func (u *user) IsLocal() bool {
	return u.local
}

func (u *user) Email() string {
	return u.email
}

func (u *user) UnmarshalJSON(data []byte) error {
	des := &struct {
		SuperUser bool   `json:"is_superuser"`
		UserName  string `json:"username"`
		Email     string `json:"email"`
		Local     bool   `json:"is_local"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	u.superuser = des.SuperUser
	u.local = des.Local
	u.email = des.Email
	u.username = des.UserName

	return nil
}

type users struct {
	client  *authenticatedClient
	apiPath string
	params  Params
}

func (u *users) List(ctx context.Context) ([]User, error) {
	res, err := u.client.Get(ctx, u.apiPath, u.params.Values())
	if err != nil {
		return nil, err
	}

	var obj []*user
	err = unMarshalJson(res, &obj)

	return userSliceToInterface(obj), err
}

func userSliceToInterface(input []*user) []User {
	var res []User
	for _, u := range input {
		res = append(res, u)
	}
	return res
}

func (u *users) WhoAmI(ctx context.Context) (User, error) {
	u.params.Reset()
	u.params.Set(Operation, OperationWhoAmI)

	res, err := u.client.Get(ctx, u.apiPath, u.params.Values())
	if err != nil {
		return nil, err
	}

	var obj *user
	err = unMarshalJson(res, &obj)
	return obj, err
}

func NewUsersClient(client *authenticatedClient) Users {
	return &users{
		client:  client,
		apiPath: UsersAPIPath,
		params:  ParamsBuilder(),
	}
}
