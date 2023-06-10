package authz

import (
	"context"
)

type MockAuthnClient struct {
	users map[string]User
}

func NewMockAuthnClient() *MockAuthnClient {
	return &MockAuthnClient{
		users: map[string]User{},
	}
}

func (c *MockAuthnClient) AddUser(key string, u User) {
	c.users[key] = u
}

func (c *MockAuthnClient) UserByID(ctx context.Context, id string) (*User, error) {
	if user, ok := c.users[id]; ok {
		user := user
		return &user, nil
	}
	return nil, nil
}

func (c *MockAuthnClient) Users(ctx context.Context, userQuery string) ([]*User, error) {
	var ret []*User
	for _, user := range c.users {
		user := user
		ret = append(ret, &user)
	}
	return ret, nil
}
