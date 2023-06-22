package authz

import (
	"context"
	"errors"
	"strings"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
)

type MockAuthnClient struct {
	users map[string]*azpb.User
}

func NewMockAuthnClient() *MockAuthnClient {
	return &MockAuthnClient{
		users: map[string]*azpb.User{},
	}
}

func (c *MockAuthnClient) AddUser(key string, u *azpb.User) {
	c.users[key] = &azpb.User{Id: u.Id, Name: u.Name, Email: u.Email}
}

func (c *MockAuthnClient) UserByID(ctx context.Context, id string) (*azpb.User, error) {
	if user, ok := c.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("unauthorized")
}

func (c *MockAuthnClient) Users(ctx context.Context, userQuery string) ([]*azpb.User, error) {
	var ret []*azpb.User
	uq := strings.ToLower(userQuery)
	for _, user := range c.users {
		user := user
		un := strings.ToLower(user.Name)
		if userQuery == "" || strings.Contains(un, uq) {
			ret = append(ret, user)
		}
	}
	return ret, nil
}
