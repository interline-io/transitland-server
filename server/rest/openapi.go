package rest

import (
	oa "github.com/getkin/kin-openapi/openapi3"
)

func newSchema(st string, format string, enum []any) *oa.Schema {
	return &oa.Schema{
		Type:   &oa.Types{st},
		Format: format,
		Enum:   enum,
	}
}

var ParameterComponents = oa.ParametersMap{
	"adm0IsoParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "adm0_iso",
			In:          "query",
			Description: `Search by country 2 letter ISO 3166 code`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm0NameParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "adm0_name",
			In:          "query",
			Description: `Search by country name`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm1IsoParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "adm1_iso",
			In:          "query",
			Description: `Search by state/province/division ISO 3166-2 code`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm1NameParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "adm1_name",
			In:          "query",
			Description: `Search by state/province/division name`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"afterParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "after",
			In:          "query",
			Description: `Pagination cursor value. This should be treated as an opaque value created by the server and returned as the link to the next result page, which may be empty. For historical reasons, this is based on the integer record ID values, but that should not be assumed to be the case in the future.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"bboxParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "bbox",
			In:          "query",
			Description: `Geographic search using a bounding box, with coordinates in (min_lon, min_lat, max_lon, max_lat) order as a comma separated string`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"cityNameParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "city_name",
			In:          "query",
			Description: `Search by city name`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"feedParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "feed_onestop_id",
			In:          "query",
			Description: `Search for records in this feed`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"formatParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "format",
			In:          "query",
			Description: `Response format`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"json", "geojson", "geojsonl", "png"}),
			},
		},
	},
	"idParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "id",
			In:          "query",
			Description: `Search for a specific internal ID`,
			Schema: &oa.SchemaRef{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"includeAlertsParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "include_alerts",
			In:          "query",
			Description: `Include alerts from GTFS Realtime feeds`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"true", "false"}),
			},
		},
	},
	"latParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "lat",
			In:          "query",
			Description: `Latitude`,
			Schema: &oa.SchemaRef{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"licenseCommercialUseAllowedParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "license_commercial_use_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'commercial_use_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseCreateDerivedProductParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "license_create_derived_product",
			In:          "query",
			Description: `Filter entities by feed license 'create_derived_product' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseRedistributionAllowedParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "license_redistribution_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'redistribution_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseShareAlikeOptionalParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "license_share_alike_optional",
			In:          "query",
			Description: `Filter entities by feed license 'share_alike_optional' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseUseWithoutAttributionParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "license_use_without_attribution",
			In:          "query",
			Description: `Filter entities by feed license 'use_without_attribution' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"limitParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "limit",
			In:          "query",
			Description: `Maximum number of records to return`,
			Schema: &oa.SchemaRef{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"lonParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "lon",
			In:          "query",
			Description: `Longitude`,
			Schema: &oa.SchemaRef{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"onestopParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "onestop_id",
			In:          "query",
			Description: `Search for a specific Onestop ID`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"radiusParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "radius",
			In:          "query",
			Description: `Search radius (meters); requires lat and lon`,
			Schema: &oa.SchemaRef{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"relativeDateParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "relative_date",
			In:          "query",
			Description: `Search for departures on a relative date label, e.g. TODAY, TUESDAY, NEXT_WEDNESDAY`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", []interface{}{"TODAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY", "NEXT_MONDAY", "NEXT_TUESDAY", "NEXT_WEDNESDAY", "NEXT_THURSDAY", "NEXT_FRIDAY", "NEXT_SATURDAY", "NEXT_SUNDAY"}),
			},
		},
	},
	"searchParam": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "search",
			In:          "query",
			Description: `Full text search`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"sha1Param": &oa.ParameterRef{
		Value: &oa.Parameter{
			Name:        "feed_version_sha1",
			In:          "query",
			Description: `Search for records in this feed version`,
			Schema: &oa.SchemaRef{
				Value: newSchema("string", "", nil),
			},
		},
	},
}
var PathItems = map[string]*oa.PathItem{
	"/stops": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request stops in specified format", "method": "get", "path": "/stops.{format}"}, map[string]interface{}{"description": "Request a stop", "method": "get", "path": "/stops/{stop_key}"}, map[string]interface{}{"description": "Request a stop in a specified format", "method": "get", "path": "/stops/{stop_key}.{format}"}},
		},
		Get: &oa.Operation{
			Summary:     "Stops",
			Description: `Search for stops`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "stop_key",
						In:          "query",
						Description: `Stop lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs stop_id>' key, or a Onestop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/stops?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/stops?format=geojson"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/searchParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=embarcadero", "url": "/stops?search=embarcadero"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/onestopParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=...", "url": "/stops?onestop_id=s-9q8yyzcny3-embarcadero"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "stop_id",
						In:          "query",
						Description: `Search for records with this GTFS stop_id`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "stop_id=EMBR", "url": "/stops?feed_onestop_id=f-c20-trimet&stop_id=1108"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "served_by_onestop_ids",
						In:          "query",
						Description: `Search stops visited by a route or agency OnestopID. Accepts comma separated values.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "served_by_onestop_ids=o-9q9-bart,o-9q9-caltrain", "url": "/stops?served_by_onestop_ids=o-9q9-bart,o-9q9-caltrain"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "served_by_route_type",
						In:          "query",
						Description: `Search for stops served by a particular route (vehicle) type`,
						Schema: &oa.SchemaRef{
							Value: newSchema("integer", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "served_by_route_type=1", "url": "/stops?served_by_route_type=1"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/sha1Param",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=1c4721d4...", "url": "/stops?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/feedParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-c20-trimet", "url": "/stops?feed_onestop_id=f-c20-trimet"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for stops geographically; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/stops?lon=-122.3&lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/stops?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/stop_times": {
		Get: &oa.Operation{
			Summary:     "Stop times",
			Description: `Search for stop times`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "trip_id",
						In:          "query",
						Description: `Stop times with this internal trip ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("integer", "int64", nil),
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "stop_id",
						In:          "query",
						Description: `Stop times with this internal stop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("integer", "int64", nil),
						},
					},
				},
			},
		},
	},
	"/routes": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request routes in specified format", "method": "get", "path": "/routes.{format}"}, map[string]interface{}{"description": "Request a route", "method": "get", "path": "/routes/{route_key}"}, map[string]interface{}{"description": "Request a route in a specified format", "method": "get", "path": "/routes/{route_key}.{format}"}},
		},
		Get: &oa.Operation{
			Summary:     "Routes",
			Description: `Search for routes`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "route_key",
						In:          "query",
						Description: `Route lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs route_id>' key, or a Onestop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/routes?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=png", "url": "/routes?format=png&feed_onestop_id=f-dr5r7-nycdotsiferry"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/searchParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=daly+city", "url": "/routes?search=daly+city"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/onestopParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=r-9q9j-l1", "url": "/routes?onestop_id=r-9q9j-l1"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "route_id",
						In:          "query",
						Description: `Search for records with this GTFS route_id`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "route_id=Bu-130", "url": "/routes?feed_onestop_id=f-sf~bay~area~rg&route_id=AC:10"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "route_type",
						In:          "query",
						Description: `Search for routes with this GTFS route (vehicle) type`,
						Schema: &oa.SchemaRef{
							Value: newSchema("integer", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "route_type=1", "url": "/routes?route_type=1"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "operator_onestop_id",
						In:          "query",
						Description: `Search for records by operator OnestopID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "operator_onestop_id=...", "url": "/routes?operator_onestop_id=o-9q9-caltrain"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include route geometry`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "include_geometry=true", "url": "/routes?include_geometry=true"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/sha1Param",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=041ffeec...", "url": "/routes?feed_version_sha1=041ffeec98316e560bc2b91960f7150ad329bd5f"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/feedParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/routes?feed_onestop_id=f-sf~bay~area~rg"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for routes geographically, based on stops at this location; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/routes?lon=-122.3&lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/routes?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/operators": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request operators in specified format", "method": "get", "path": "/operators.{format}"}, map[string]interface{}{"description": "Request an operator by Onestop ID", "method": "get", "path": "/operators/{onestop_id}"}},
		},
		Get: &oa.Operation{
			Summary:     "Operators",
			Description: `Search for operators`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/searchParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=bart", "url": "/operators?search=caltrain"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/onestopParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=o-9q9-caltrain", "url": "/operators?onestop_id=o-9q9-caltrain"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/feedParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/operators?feed_onestop_id=f-sf~bay~area~rg"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "tag_key",
						In:          "query",
						Description: `Search for operators with a tag. Combine with tag_value also query for the value of the tag.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "tag_key=us_ntd_id", "url": "/operators?tag_key=us_ntd_id"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "tag_value",
						In:          "query",
						Description: `Search for feeds tagged with a given value. Must be combined with tag_key.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "tag_key=us_ntd_id&tag_value=40029", "url": "/operators?tag_key=us_ntd_id&tag_value=40029"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm0NameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_name=Mexico", "url": "/operators?adm0_name=Mexico"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm0IsoParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_iso=US", "url": "/operators?adm0_iso=US"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm1NameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_name=California", "url": "/operators?adm1_name=California"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm1IsoParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_iso=US-CA", "url": "/operators?adm1_iso=US-CA"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/cityNameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "city_name=Oakland", "url": "/operators?city_name=Oakland"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for operators geographically, based on stops at this location; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/operators?lon=-122.3&lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/operators?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/feeds": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request feeds in specified format", "method": "get", "path": "/feeds.{format}"}, map[string]interface{}{"description": "Request a feed by ID or Onestop ID", "method": "get", "path": "/feeds/{feed_key}"}, map[string]interface{}{"description": "Request a feed by ID or Onestop ID in specified format", "method": "get", "path": "/feeds/{feed_key}.{format}"}},
		},
		Get: &oa.Operation{
			Summary:     "Feeds",
			Description: `Search for feeds`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_key",
						In:          "query",
						Description: `Feed lookup key; can be an integer ID or a Onestop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/feeds?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/feeds?format=geojson"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/searchParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=caltrain", "url": "/feeds?search=caltrain"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/onestopParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=f-sf~bay~area~rg", "url": "/feeds?onestop_id=f-sf~bay~area~rg"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "spec",
						In:          "query",
						Description: `Type of data contained in this feed`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"gtfs", "gtfs-rt", "gbfs", "mds"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "spec=gtfs", "url": "/feeds?spec=gtfs"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "fetch_error",
						In:          "query",
						Description: `Search for feeds with or without a fetch error`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "fetch_error=true", "url": "/feeds?fetch_error=true"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "tag_key",
						In:          "query",
						Description: `Search for feeds with a tag. Combine with tag_value also query for the value of the tag.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "tag_key=gtfs_data_exchange", "url": "/feeds?tag_key=gtfs_data_exchange"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "tag_value",
						In:          "query",
						Description: `Search for feeds tagged with a given value. Must be combined with tag_key.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "tag_key=unstable_url&tag_value=true", "url": "/feeds?tag_key=unstable_url&tag_value=true"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for feeds geographically; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/feeds?lon=-122.3?lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feeds?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/feed_versions": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request feed versions in specified format", "method": "get", "path": "/feeds_versions.{format}"}, map[string]interface{}{"description": "Request a feed version by ID or SHA1", "method": "get", "path": "/feeds_versions/{feed_version_key}"}, map[string]interface{}{"description": "Request a feed version by ID or SHA1 in specified format", "method": "get", "path": "/feeds_versions/{feed_version_key}.{format}"}, map[string]interface{}{"description": "Request feed versions by feed ID or OnestopID", "method": "get", "path": "/feeds/{feed_key}/feed_versions"}},
		},
		Get: &oa.Operation{
			Summary:     "Feed Versions",
			Description: `Search for feed versions`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_version_key",
						In:          "query",
						Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_key",
						In:          "query",
						Description: `Feed lookup key; can be an integer ID or Onestop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/feed_versions?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/feed_versions?format=geojson"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "sha1",
						In:          "query",
						Description: `Feed version SHA1`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "sha1=e535eb2b3...", "url": "/feed_versions?sha1=dd7aca4a8e4c90908fd3603c097fabee75fea907"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_onestop_id",
						In:          "query",
						Description: `Feed OnestopID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/feed_versions?feed_onestop_id=f-sf~bay~area~rg"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "fetched_before",
						In:          "query",
						Description: `Filter for feed versions fetched earlier than given date time in UTC`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "datetime", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "fetched_before=2023-01-01T00:00:00Z", "url": "/feed_versions?fetched_before=2023-01-01T00:00:00Z"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "fetched_after",
						In:          "query",
						Description: `Filter for feed versions fetched since given date time in UTC`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "datetime", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "fetched_after=2023-01-01T00:00:00Z", "url": "/feed_versions?fetched_after=2023-01-01T00:00:00Z"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for feed versions geographically; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/feed_versions?lon=-122.3&lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feed_versions?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
			},
		},
	},
	"/agencies": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request agencies in specified format", "method": "get", "path": "/agencies.{format}"}, map[string]interface{}{"description": "Request an agency", "method": "get", "path": "/agencies/{agency_key}"}, map[string]interface{}{"description": "Request an agency in specified format", "method": "get", "path": "/agencies/{agency_key}.{format}"}},
		},
		Get: &oa.Operation{
			Summary:     "Agencies",
			Description: ``,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "agency_key",
						In:          "query",
						Description: `Agency lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs agency_id>' key, or a Onestop ID`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/agencies?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/agencies?format=geojson"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/searchParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=bart", "url": "/agencies?search=bart"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/onestopParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=o-9q9-caltrain", "url": "/agencies?onestop_id=o-9q9-caltrain"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/sha1Param",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=1c4721d4...", "url": "/agencies?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/feedParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/agencies?feed_onestop_id=f-sf~bay~area~rg"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "agency_id",
						In:          "query",
						Description: `Search for records with this GTFS agency_id (string)`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "agency_id=BART", "url": "/agencies?agency_id=BART"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "agency_name",
						In:          "query",
						Description: `Search for records with this GTFS agency_name`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "agency_name=Caltrain", "url": "/agencies?agency_name=Caltrain"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/radiusParam",
					Extensions: map[string]any{
						"x-description":      "Search for agencies geographically, based on stops at this location; radius is in meters, requires lon and lat",
						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/agencies?lon=-122.3&lat=37.8&radius=1000"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/bboxParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/agencies?bbox=-122.269,37.807,-122.267,37.808"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm0NameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_name=Mexico", "url": "/agencies?adm0_name=Mexico"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm0IsoParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_iso=US", "url": "/agencies?adm0_iso=US"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm1NameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_name=California", "url": "/agencies?adm1_name=California"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/adm1IsoParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_iso=US-CA", "url": "/agencies?adm1_iso=US-CA"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/cityNameParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "city_name=Oakland", "url": "/agencies?city_name=Oakland"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/stops/{stop_key}/departures": {
		Get: &oa.Operation{
			Summary:     "Stop departures",
			Description: `Departures from a given stop based on static and real-time data`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "stop_key",
						In:          "path",
						Description: `Stop lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs stop_id'> key, a Onestop ID`,
						Required:    true,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "f-sf~bay~area~rg:LAKE", "url": "/stops/f-sf~bay~area~rg:LAKE/departures"}, map[string]interface{}{"description": "s-9q9p1bc1td-lakemerritt", "url": "/stops/s-9q9p1bc1td-lakemerritt/departures"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "service_date",
						In:          "query",
						Description: `Search for departures on a specified GTFS service calendar date, in YYYY-MM-DD format`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "date", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "service_date=2022-09-28", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?service_date=2022-09-28"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "date",
						In:          "query",
						Description: `Search for departures on a specified calendar date, in YYYY-MM-DD format`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "date", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "date=2022-09-28", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?date=2022-09-28"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/relativeDateParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "relative_date=NEXT_MONDAY", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?relative_date=NEXT_MONDAY"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "next",
						In:          "query",
						Description: `Search for departures leaving within the next specified number of seconds in local time`,
						Schema: &oa.SchemaRef{
							Value: newSchema("integer", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "next=600", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?next=600"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "start_time",
						In:          "query",
						Description: `Search for departures leaving after a specified local time, in HH:MM:SS format`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "start_time=10:00:00", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?start_time=10:00:00&service_date=2022-09-28"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "end_time",
						In:          "query",
						Description: `Search for departures leaving before a specified local time, in HH:MM:SS format`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "end_time=11:00:00", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?end_time=11:00:00&service_date=2022-09-28"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include route geometry`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "include_geometry=true", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?include_geometry=true"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "use_service_window",
						In:          "query",
						Description: `Use a fall-back service date if the requested service_date is outside the active service period of the feed version. The fall-back date is selected as the matching day-of-week in the week which provides the best level of scheduled service in the feed version. This value defaults to true.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "use_service_window=false", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?use_service_window=false"}},
						},
					},
				},
			},
		},
	},
	"/routes/{route_key}/trips": {
		Extensions: map[string]any{
			"x-alternates": []interface{}{map[string]interface{}{"description": "Request trips in specified format", "method": "get", "path": "/routes/{route_key}/trips.{format}"}, map[string]interface{}{"description": "Request a single trip by ID", "method": "get", "path": "/routes/{route_key}/trips/{id}"}, map[string]interface{}{"description": "Request a single trip by ID in specified format", "method": "get", "path": "/routes/{route_key}/trips/{id}.format"}},
		},
		Get: &oa.Operation{
			Summary:     "Trips",
			Description: `Search for trips`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Ref: "#/components/parameters/includeAlertsParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/idParam",
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "route_key",
						In:          "path",
						Description: `Route lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs route_id>' key, or a Onestop ID`,
						Required:    true,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/afterParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/limitParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/routes/r-9q9j-l1/trips?limit=1"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/formatParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/routes/r-9q9j-l1/trips?limit=10&format=geojson"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "service_date",
						In:          "query",
						Description: `Search for trips active on this date`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "date", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "service_date=...", "url": "/routes/r-9q9j-l1/trips?service_date=2021-07-14"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/relativeDateParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "/routes/r-9q9j-l1/trips?relative_date=NEXT_MONDAY", "url": "/routes/r-9q9j-l1/trips?relative_date=NEXT_MONDAY"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "trip_id",
						In:          "query",
						Description: `Search for records with this GTFS trip_id`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "trip_id=305", "url": "/routes/r-9q9j-l1/trips?trip_id=305"}},
						},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include shape geometry`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "include_geometry=true", "url": "/routes/r-9q9j-l1/trips?limit=10&include_geometry=true"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/sha1Param",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=041ffeec...", "url": "/routes/r-9q9j-l1/trips?feed_version_sha1=041ffeec98316e560bc2b91960f7150ad329bd5f"}},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/feedParam",
					Extensions: map[string]any{
						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/routes/r-9q9j-l1/trips?feed_onestop_id=f-sf~bay~area~rg"}},
					},
				},
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "use_service_window",
						In:          "query",
						Description: `Use a fall-back service date if the requested service_date is outside the active service period of the feed version. The fall-back date is selected as the matching day-of-week in the week which provides the best level of scheduled service in the feed version. This value defaults to true.`,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", []interface{}{"true", "false"}),
						},
						Extensions: map[string]any{
							"x-example-requests": []interface{}{map[string]interface{}{"description": "use_service_window=false", "url": "/routes/r-9q9j-l1/trips?use_service_window=false"}},
						},
					},
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/latParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/lonParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&oa.ParameterRef{
					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},
	"/feeds/{feed_key}/download_latest_feed_version": {
		Get: &oa.Operation{
			Summary:     "",
			Description: `Download latest feed version for this feed`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_key",
						In:          "path",
						Description: `Feed lookup key; can be an integer ID or a Onestop ID`,
						Required:    true,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
			},
		},
	},
	"/feed_versions/{feed_version_key}/download": {
		Get: &oa.Operation{
			Summary:     "",
			Description: `Download this feed version`,
			Parameters: oa.Parameters{
				&oa.ParameterRef{
					Value: &oa.Parameter{
						Name:        "feed_version_key",
						In:          "path",
						Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,
						Required:    true,
						Schema: &oa.SchemaRef{
							Value: newSchema("string", "", nil),
						},
					},
				},
			},
		},
	},
}
