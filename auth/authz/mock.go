package authz

import (
	"context"
	"errors"
	"strings"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
)

type MockUserProvider struct {
	users map[string]*azpb.User
}

func NewMockUserProvider() *MockUserProvider {
	return &MockUserProvider{
		users: map[string]*azpb.User{},
	}
}

func (c *MockUserProvider) AddUser(key string, u *azpb.User) {
	c.users[key] = &azpb.User{Id: u.Id, Name: u.Name, Email: u.Email}
}

func (c *MockUserProvider) UserByID(ctx context.Context, id string) (*azpb.User, error) {
	if user, ok := c.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("unauthorized")
}

func (c *MockUserProvider) Users(ctx context.Context, userQuery string) ([]*azpb.User, error) {
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

//////

type MockFGAClient struct{}

func NewMockFGAClient() *MockFGAClient {
	return &MockFGAClient{}
}

func (c *MockFGAClient) Check(context.Context, TupleKey, ...TupleKey) (bool, error) {
	return false, nil
}

func (c *MockFGAClient) ListObjects(context.Context, TupleKey) ([]TupleKey, error) {
	return nil, nil
}

func (c *MockFGAClient) GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error) {
	return nil, nil
}

func (c *MockFGAClient) WriteTuple(context.Context, TupleKey) error {
	return nil
}

func (c *MockFGAClient) ReplaceTuple(context.Context, TupleKey) error {
	return nil
}

func (c *MockFGAClient) DeleteTuple(context.Context, TupleKey) error {
	return nil
}
