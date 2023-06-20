package authz

import (
	"context"
	"errors"
	"strings"
)

type MockAuthnClient struct {
	users map[string]*User
}

func NewMockAuthnClient() *MockAuthnClient {
	return &MockAuthnClient{
		users: map[string]*User{},
	}
}

func (c *MockAuthnClient) AddUser(key string, u *User) {
	c.users[key] = &User{Id: u.Id, Name: u.Name, Email: u.Email}
}

func (c *MockAuthnClient) UserByID(ctx context.Context, id string) (*User, error) {
	if user, ok := c.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("unauthorized")
}

func (c *MockAuthnClient) Users(ctx context.Context, userQuery string) ([]*User, error) {
	var ret []*User
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
