package openapi

import "github.com/getkin/kin-openapi/openapi3"

type Parameter = openapi3.Parameter
type ParameterRef = openapi3.ParameterRef

type RequestInfo struct {
	Path  string
	Query string
	Get   *openapi3.Operation
}
