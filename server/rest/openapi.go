package rest

import "github.com/getkin/kin-openapi/openapi3"

type Schema = openapi3.Schema
type SchemaRef = openapi3.SchemaRef
type Parameter = openapi3.Parameter
type ParameterRef = openapi3.ParameterRef

func newSchemaRef(st string, enum []any) *SchemaRef {
	return &SchemaRef{
		Value: &Schema{
			Type: &openapi3.Types{st},
			Enum: enum,
		},
	}
}

var ComponentParameters = openapi3.ParametersMap{
	"radiusParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search radius (meters); requires lat and lon",
		Schema:      newSchemaRef("number", nil),
	}},
	"formatParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Response format",
		Schema:      newSchemaRef("string", []any{"json", "geojson", "geojsonl", "png"}),
	}},
	"feedParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search for records in this feed",
		Schema:      newSchemaRef("string", nil),
	}},
	"adm1NameParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search by state/province/division name",
		Schema:      newSchemaRef("string", nil),
	}},
	"licenseCommercialUseAllowedParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Filter entities by feed license 'commercial_use_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.",
		Schema:      newSchemaRef("string", []any{"yes", "no", "unknown", "exclude_no"}),
	}},
	"licenseCreateDerivedProductParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Filter entities by feed license 'create_derived_product' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.",
		Schema:      newSchemaRef("string", []any{"yes", "no", "unknown", "exclude_no"}),
	}},
	"licenseUseWithoutAttributionParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Filter entities by feed license 'use_without_attribution' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.",
		Schema:      newSchemaRef("string", []any{"yes", "no", "unknown", "exclude_no"}),
	}},
	"bboxParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Geographic search using a bounding box, with coordinates in (min_lon, min_lat, max_lon, max_lat) order as a comma separated string",
		Schema:      newSchemaRef("string", nil),
	}},
	"searchParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Full text search",
		Schema:      newSchemaRef("string", nil),
	}},
	"adm0NameParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search by country name",
		Schema:      newSchemaRef("string", nil),
	}},
	"includeAlertsParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Include alerts from GTFS Realtime feeds",
		Schema:      newSchemaRef("string", []any{"true", "false"}),
	}},
	"relativeDateParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search for departures on a relative date label, e.g. TODAY, TUESDAY, NEXT_WEDNESDAY",
		Schema:      newSchemaRef("string", []any{"TODAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY", "NEXT_MONDAY", "NEXT_TUESDAY", "NEXT_WEDNESDAY", "NEXT_THURSDAY", "NEXT_FRIDAY", "NEXT_SATURDAY", "NEXT_SUNDAY"}),
	}},
	"lonParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Longitude",
		Schema:      newSchemaRef("number", nil),
	}},
	"adm1IsoParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search by state/province/division ISO 3166-2 code",
		Schema:      newSchemaRef("string", nil),
	}},
	"onestopParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search for a specific Onestop ID",
		Schema:      newSchemaRef("string", nil),
	}},
	"idParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search for a specific internal ID",
		Schema:      newSchemaRef("integer", nil),
	}},
	"limitParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Maximum number of records to return",
		Schema:      newSchemaRef("integer", nil),
	}},
	"afterParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Pagination cursor value. This should be treated as an opaque value created by the server and returned as the link to the next result page, which may be empty. For historical reasons, this is based on the integer record ID values, but that should not be assumed to be the case in the future.",
		Schema:      newSchemaRef("integer", nil),
	}},
	"sha1Param": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search for records in this feed version",
		Schema:      newSchemaRef("string", nil),
	}},
	"adm0IsoParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search by country 2 letter ISO 3166 code",
		Schema:      newSchemaRef("string", nil),
	}},
	"cityNameParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Search by city name",
		Schema:      newSchemaRef("string", nil),
	}},
	"licenseShareAlikeOptionalParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Filter entities by feed license 'share_alike_optional' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.",
		Schema:      newSchemaRef("string", []any{"yes", "no", "unknown", "exclude_no"}),
	}},
	"licenseRedistributionAllowedParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Filter entities by feed license 'redistribution_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.",
		Schema:      newSchemaRef("string", []any{"yes", "no", "unknown", "exclude_no"}),
	}},
	"latParam": &ParameterRef{Value: &Parameter{
		In:          "query",
		Description: "Latitude",
		Schema:      newSchemaRef("number", nil),
	}}}

type RequestInfo struct {
	Path  string
	Query string
	Get   RequestInfoMethod
}

type RequestInfoMethod struct {
	Summary     string
	Description string
	Alternates  []AltRequestPath
	Parameters  []RequestInfoParam
}

type RequestInfoParam struct {
	Component   string
	Ref         string
	In          string
	Name        string
	Description string
	SchemaType  string
}

type RequestInfoParamExample struct {
	Description string
	URL         string
}

type AltRequestPath struct {
	Path        string
	Description string
}
