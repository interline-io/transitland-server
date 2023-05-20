package authz

import (
	"context"
)

type MockAuthnClient struct {
	users map[string]User
}

func NewMockAuthnClient() *MockAuthnClient {
	return &MockAuthnClient{
		users: map[string]User{
			"ian":   {Name: "Ian", ID: "ian", Email: "ian@example.com"},
			"drew":  {Name: "Drew", ID: "drew", Email: "drew@example.com"},
			"nisar": {Name: "Nisar", ID: "nisar", Email: "nisar@example.com"},
		},
	}
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
