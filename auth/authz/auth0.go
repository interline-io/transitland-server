package authz

import (
	"context"

	"github.com/auth0/go-auth0/management"
)

type Auth0Client struct {
	client *management.Management
}

func NewAuth0Client(domain string, clientId string, clientSecret string) (*Auth0Client, error) {
	auth0API, err := management.New(
		domain,
		management.WithClientCredentials(clientId, clientSecret),
	)
	if err != nil {
		return nil, err
	}
	return &Auth0Client{client: auth0API}, nil
}

func (c *Auth0Client) UserByID(ctx context.Context, id string) (*User, error) {
	user, err := c.client.User.Read(id)
	if err != nil {
		return nil, err
	}
	u := &User{
		Id:    user.GetID(),
		Name:  user.GetName(),
		Email: user.GetEmail(),
	}
	return u, nil
}

func (c *Auth0Client) Users(ctx context.Context, userQuery string) ([]*User, error) {
	ul, err := c.client.User.List(management.Query(userQuery))
	if err != nil {
		return nil, err
	}
	var ret []*User
	for _, user := range ul.Users {
		ret = append(ret, &User{
			Id:    user.GetID(),
			Name:  user.GetName(),
			Email: user.GetEmail(),
		})
	}
	return ret, nil
}
