package authz

import "context"

type MockAuthzClient struct{}

func NewMockAuthzClient() *MockAuthzClient {
	return &MockAuthzClient{}
}

func (c *MockAuthzClient) Check(context.Context, TupleKey, ...TupleKey) (bool, error) {
	return false, nil
}

func (c *MockAuthzClient) ListObjects(context.Context, TupleKey) ([]TupleKey, error) {
	return nil, nil
}

func (c *MockAuthzClient) GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error) {
	return nil, nil
}

func (c *MockAuthzClient) WriteTuple(context.Context, TupleKey) error {
	return nil
}

func (c *MockAuthzClient) ReplaceTuple(context.Context, TupleKey) error {
	return nil
}

func (c *MockAuthzClient) DeleteTuple(context.Context, TupleKey) error {
	return nil
}
