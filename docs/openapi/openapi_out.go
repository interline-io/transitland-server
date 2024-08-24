package openapi

import "github.com/getkin/kin-openapi/openapi3"

var parameterComponents = openapi3.ParametersMap{

	"adm0IsoParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "adm0_iso",
			In:          "query",
			Description: `Search by country 2 letter ISO 3166 code`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"adm0NameParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "adm0_name",
			In:          "query",
			Description: `Search by country name`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"adm1IsoParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "adm1_iso",
			In:          "query",
			Description: `Search by state/province/division ISO 3166-2 code`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"adm1NameParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "adm1_name",
			In:          "query",
			Description: `Search by state/province/division name`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"afterParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "after",
			In:          "query",
			Description: `Pagination cursor value. This should be treated as an opaque value created by the server and returned as the link to the next result page, which may be empty. For historical reasons, this is based on the integer record ID values, but that should not be assumed to be the case in the future.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   &openapi3.Types{"integer"},
					Format: "int32",
				},
			},
		},
	},

	"bboxParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "bbox",
			In:          "query",
			Description: `Geographic search using a bounding box, with coordinates in (min_lon, min_lat, max_lon, max_lat) order as a comma separated string`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"cityNameParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "city_name",
			In:          "query",
			Description: `Search by city name`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"feedParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "feed_onestop_id",
			In:          "query",
			Description: `Search for records in this feed`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"formatParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "format",
			In:          "query",
			Description: `Response format`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"json", "geojson", "geojsonl", "png"},
				},
			},
		},
	},

	"idParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "id",
			In:          "query",
			Description: `Search for a specific internal ID`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   &openapi3.Types{"integer"},
					Format: "int32",
				},
			},
		},
	},

	"includeAlertsParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "include_alerts",
			In:          "query",
			Description: `Include alerts from GTFS Realtime feeds`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},

					Enum: []interface{}{"true", "false"},
				},
			},
		},
	},

	"latParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "lat",
			In:          "query",
			Description: `Latitude`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"number"},
				},
			},
		},
	},

	"licenseCommercialUseAllowedParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "license_commercial_use_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'commercial_use_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"yes", "no", "unknown", "exclude_no"},
				},
			},
		},
	},

	"licenseCreateDerivedProductParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "license_create_derived_product",
			In:          "query",
			Description: `Filter entities by feed license 'create_derived_product' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"yes", "no", "unknown", "exclude_no"},
				},
			},
		},
	},

	"licenseRedistributionAllowedParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "license_redistribution_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'redistribution_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"yes", "no", "unknown", "exclude_no"},
				},
			},
		},
	},

	"licenseShareAlikeOptionalParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "license_share_alike_optional",
			In:          "query",
			Description: `Filter entities by feed license 'share_alike_optional' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"yes", "no", "unknown", "exclude_no"},
				},
			},
		},
	},

	"licenseUseWithoutAttributionParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "license_use_without_attribution",
			In:          "query",
			Description: `Filter entities by feed license 'use_without_attribution' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{

					Enum: []interface{}{"yes", "no", "unknown", "exclude_no"},
				},
			},
		},
	},

	"limitParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "limit",
			In:          "query",
			Description: `Maximum number of records to return`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   &openapi3.Types{"integer"},
					Format: "int32",
				},
			},
		},
	},

	"lonParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "lon",
			In:          "query",
			Description: `Longitude`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"number"},
				},
			},
		},
	},

	"onestopParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "onestop_id",
			In:          "query",
			Description: `Search for a specific Onestop ID`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"radiusParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "radius",
			In:          "query",
			Description: `Search radius (meters); requires lat and lon`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"number"},
				},
			},
		},
	},

	"relativeDateParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "relative_date",
			In:          "query",
			Description: `Search for departures on a relative date label, e.g. TODAY, TUESDAY, NEXT_WEDNESDAY`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},

					Enum: []interface{}{"TODAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY", "NEXT_MONDAY", "NEXT_TUESDAY", "NEXT_WEDNESDAY", "NEXT_THURSDAY", "NEXT_FRIDAY", "NEXT_SATURDAY", "NEXT_SUNDAY"},
				},
			},
		},
	},

	"searchParam": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "search",
			In:          "query",
			Description: `Full text search`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},

	"sha1Param": &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "feed_version_sha1",
			In:          "query",
			Description: `Search for records in this feed version`,

			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	},
}

var pathItems = map[string]*openapi3.PathItem{

	"/stops": {

		Get: &openapi3.Operation{
			Summary:     "Stops",
			Description: `Search for stops`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "stop_key",
						In:          "query",
						Description: `Stop lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs stop_id>' key, or a Onestop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/stops?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/stops?format=geojson"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=embarcadero", "url": "/stops?search=embarcadero"}},
					},

					Ref: "#/components/parameters/searchParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=...", "url": "/stops?onestop_id=s-9q8yyzcny3-embarcadero"}},
					},

					Ref: "#/components/parameters/onestopParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "stop_id",
						In:          "query",
						Description: `Search for records with this GTFS stop_id`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "served_by_onestop_ids",
						In:          "query",
						Description: `Search stops visited by a route or agency OnestopID. Accepts comma separated values.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "served_by_route_type",
						In:          "query",
						Description: `Search for stops served by a particular route (vehicle) type`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"integer"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=1c4721d4...", "url": "/stops?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
					},

					Ref: "#/components/parameters/sha1Param",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-c20-trimet", "url": "/stops?feed_onestop_id=f-c20-trimet"}},
					},

					Ref: "#/components/parameters/feedParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for stops geographically; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/stops?lon=-122.3&lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/stops?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/stop_times": {

		Get: &openapi3.Operation{
			Summary:     "Stop times",
			Description: `Search for stop times`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "trip_id",
						In:          "query",
						Description: `Stop times with this internal trip ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"integer"},
								Format: "int64",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "stop_id",
						In:          "query",
						Description: `Stop times with this internal stop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"integer"},
								Format: "int64",
							},
						},
					},
				},
			},
		},
	},

	"/routes": {

		Get: &openapi3.Operation{
			Summary:     "Routes",
			Description: `Search for routes`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "route_key",
						In:          "query",
						Description: `Route lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs route_id>' key, or a Onestop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/routes?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=png", "url": "/routes?format=png&feed_onestop_id=f-dr5r7-nycdotsiferry"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=daly+city", "url": "/routes?search=daly+city"}},
					},

					Ref: "#/components/parameters/searchParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=r-9q9j-l1", "url": "/routes?onestop_id=r-9q9j-l1"}},
					},

					Ref: "#/components/parameters/onestopParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "route_id",
						In:          "query",
						Description: `Search for records with this GTFS route_id`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "route_type",
						In:          "query",
						Description: `Search for routes with this GTFS route (vehicle) type`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"integer"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "operator_onestop_id",
						In:          "query",
						Description: `Search for records by operator OnestopID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include route geometry`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=041ffeec...", "url": "/routes?feed_version_sha1=041ffeec98316e560bc2b91960f7150ad329bd5f"}},
					},

					Ref: "#/components/parameters/sha1Param",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/routes?feed_onestop_id=f-sf~bay~area~rg"}},
					},

					Ref: "#/components/parameters/feedParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for routes geographically, based on stops at this location; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/routes?lon=-122.3&lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/routes?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/operators": {

		Get: &openapi3.Operation{
			Summary:     "Operators",
			Description: `Search for operators`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=bart", "url": "/operators?search=caltrain"}},
					},

					Ref: "#/components/parameters/searchParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=o-9q9-caltrain", "url": "/operators?onestop_id=o-9q9-caltrain"}},
					},

					Ref: "#/components/parameters/onestopParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/operators?feed_onestop_id=f-sf~bay~area~rg"}},
					},

					Ref: "#/components/parameters/feedParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "tag_key",
						In:          "query",
						Description: `Search for operators with a tag. Combine with tag_value also query for the value of the tag.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "tag_value",
						In:          "query",
						Description: `Search for feeds tagged with a given value. Must be combined with tag_key.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_name=Mexico", "url": "/operators?adm0_name=Mexico"}},
					},

					Ref: "#/components/parameters/adm0NameParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_iso=US", "url": "/operators?adm0_iso=US"}},
					},

					Ref: "#/components/parameters/adm0IsoParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_name=California", "url": "/operators?adm1_name=California"}},
					},

					Ref: "#/components/parameters/adm1NameParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_iso=US-CA", "url": "/operators?adm1_iso=US-CA"}},
					},

					Ref: "#/components/parameters/adm1IsoParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "city_name=Oakland", "url": "/operators?city_name=Oakland"}},
					},

					Ref: "#/components/parameters/cityNameParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for operators geographically, based on stops at this location; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/operators?lon=-122.3&lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/operators?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/feeds": {

		Get: &openapi3.Operation{
			Summary:     "Feeds",
			Description: `Search for feeds`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_key",
						In:          "query",
						Description: `Feed lookup key; can be an integer ID or a Onestop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/feeds?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/feeds?format=geojson"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=caltrain", "url": "/feeds?search=caltrain"}},
					},

					Ref: "#/components/parameters/searchParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=f-sf~bay~area~rg", "url": "/feeds?onestop_id=f-sf~bay~area~rg"}},
					},

					Ref: "#/components/parameters/onestopParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "spec",
						In:          "query",
						Description: `Type of data contained in this feed`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"gtfs", "gtfs-rt", "gbfs", "mds"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "fetch_error",
						In:          "query",
						Description: `Search for feeds with or without a fetch error`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "tag_key",
						In:          "query",
						Description: `Search for feeds with a tag. Combine with tag_value also query for the value of the tag.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "tag_value",
						In:          "query",
						Description: `Search for feeds tagged with a given value. Must be combined with tag_key.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for feeds geographically; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/feeds?lon=-122.3?lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feeds?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/feed_versions": {

		Get: &openapi3.Operation{
			Summary:     "Feed Versions",
			Description: `Search for feed versions`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_version_key",
						In:          "query",
						Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_key",
						In:          "query",
						Description: `Feed lookup key; can be an integer ID or Onestop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/feed_versions?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/feed_versions?format=geojson"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "sha1",
						In:          "query",
						Description: `Feed version SHA1`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_onestop_id",
						In:          "query",
						Description: `Feed OnestopID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "fetched_before",
						In:          "query",
						Description: `Filter for feed versions fetched earlier than given date time in UTC`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "datetime",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "fetched_after",
						In:          "query",
						Description: `Filter for feed versions fetched since given date time in UTC`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "datetime",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for feed versions geographically; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/feed_versions?lon=-122.3&lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feed_versions?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
			},
		},
	},

	"/agencies": {

		Get: &openapi3.Operation{
			Summary:     "Agencies",
			Description: ``,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "agency_key",
						In:          "query",
						Description: `Agency lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs agency_id>' key, or a Onestop ID`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/agencies?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/agencies?format=geojson"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "search=bart", "url": "/agencies?search=bart"}},
					},

					Ref: "#/components/parameters/searchParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "onestop_id=o-9q9-caltrain", "url": "/agencies?onestop_id=o-9q9-caltrain"}},
					},

					Ref: "#/components/parameters/onestopParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=1c4721d4...", "url": "/agencies?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
					},

					Ref: "#/components/parameters/sha1Param",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/agencies?feed_onestop_id=f-sf~bay~area~rg"}},
					},

					Ref: "#/components/parameters/feedParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "agency_id",
						In:          "query",
						Description: `Search for records with this GTFS agency_id (string)`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "agency_name",
						In:          "query",
						Description: `Search for records with this GTFS agency_name`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-description": "Search for agencies geographically, based on stops at this location; radius is in meters, requires lon and lat",

						"x-example-requests": []interface{}{map[string]interface{}{"description": "lon=-122&lat=37&radius=1000", "url": "/agencies?lon=-122.3&lat=37.8&radius=1000"}},
					},

					Ref: "#/components/parameters/radiusParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/agencies?bbox=-122.269,37.807,-122.267,37.808"}},
					},

					Ref: "#/components/parameters/bboxParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_name=Mexico", "url": "/agencies?adm0_name=Mexico"}},
					},

					Ref: "#/components/parameters/adm0NameParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm0_iso=US", "url": "/agencies?adm0_iso=US"}},
					},

					Ref: "#/components/parameters/adm0IsoParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_name=California", "url": "/agencies?adm1_name=California"}},
					},

					Ref: "#/components/parameters/adm1NameParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "adm1_iso=US-CA", "url": "/agencies?adm1_iso=US-CA"}},
					},

					Ref: "#/components/parameters/adm1IsoParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "city_name=Oakland", "url": "/agencies?city_name=Oakland"}},
					},

					Ref: "#/components/parameters/cityNameParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/stops/{stop_key}/departures": {

		Get: &openapi3.Operation{
			Summary:     "Stop departures",
			Description: `Departures from a given stop based on static and real-time data`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "stop_key",
						In:          "path",
						Description: `Stop lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs stop_id'> key, a Onestop ID`,
						Required:    true,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "service_date",
						In:          "query",
						Description: `Search for departures on a specified GTFS service calendar date, in YYYY-MM-DD format`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "date",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "date",
						In:          "query",
						Description: `Search for departures on a specified calendar date, in YYYY-MM-DD format`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "date",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "relative_date=NEXT_MONDAY", "url": "/stops/f-sf~bay~area~rg:LAKE/departures?relative_date=NEXT_MONDAY"}},
					},

					Ref: "#/components/parameters/relativeDateParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "next",
						In:          "query",
						Description: `Search for departures leaving within the next specified number of seconds in local time`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"integer"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "start_time",
						In:          "query",
						Description: `Search for departures leaving after a specified local time, in HH:MM:SS format`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "end_time",
						In:          "query",
						Description: `Search for departures leaving before a specified local time, in HH:MM:SS format`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include route geometry`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "use_service_window",
						In:          "query",
						Description: `Use a fall-back service date if the requested service_date is outside the active service period of the feed version. The fall-back date is selected as the matching day-of-week in the week which provides the best level of scheduled service in the feed version. This value defaults to true.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
			},
		},
	},

	"/routes/{route_key}/trips": {

		Get: &openapi3.Operation{
			Summary:     "Trips",
			Description: `Search for trips`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Ref: "#/components/parameters/includeAlertsParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/idParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "route_key",
						In:          "path",
						Description: `Route lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs route_id>' key, or a Onestop ID`,
						Required:    true,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/afterParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "limit=1", "url": "/routes/r-9q9j-l1/trips?limit=1"}},
					},

					Ref: "#/components/parameters/limitParam",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "format=geojson", "url": "/routes/r-9q9j-l1/trips?limit=10&format=geojson"}},
					},

					Ref: "#/components/parameters/formatParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "service_date",
						In:          "query",
						Description: `Search for trips active on this date`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "date",
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "/routes/r-9q9j-l1/trips?relative_date=NEXT_MONDAY", "url": "/routes/r-9q9j-l1/trips?relative_date=NEXT_MONDAY"}},
					},

					Ref: "#/components/parameters/relativeDateParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "trip_id",
						In:          "query",
						Description: `Search for records with this GTFS trip_id`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "include_geometry",
						In:          "query",
						Description: `Include shape geometry`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_version_sha1=041ffeec...", "url": "/routes/r-9q9j-l1/trips?feed_version_sha1=041ffeec98316e560bc2b91960f7150ad329bd5f"}},
					},

					Ref: "#/components/parameters/sha1Param",
				},
				&openapi3.ParameterRef{

					Extensions: map[string]any{

						"x-example-requests": []interface{}{map[string]interface{}{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/routes/r-9q9j-l1/trips?feed_onestop_id=f-sf~bay~area~rg"}},
					},

					Ref: "#/components/parameters/feedParam",
				},
				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "use_service_window",
						In:          "query",
						Description: `Use a fall-back service date if the requested service_date is outside the active service period of the feed version. The fall-back date is selected as the matching day-of-week in the week which provides the best level of scheduled service in the feed version. This value defaults to true.`,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},

								Enum: []interface{}{"true", "false"},
							},
						},
					},
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/latParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/lonParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseCreateDerivedProductParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseRedistributionAllowedParam",
				},
				&openapi3.ParameterRef{

					Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
				},
			},
		},
	},

	"/feeds/{feed_key}/download_latest_feed_version": {

		Get: &openapi3.Operation{
			Summary:     "",
			Description: `Download latest feed version for this feed`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_key",
						In:          "path",
						Description: `Feed lookup key; can be an integer ID or a Onestop ID`,
						Required:    true,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
			},
		},
	},

	"/feed_versions/{feed_version_key}/download": {

		Get: &openapi3.Operation{
			Summary:     "",
			Description: `Download this feed version`,

			Parameters: openapi3.Parameters{

				&openapi3.ParameterRef{

					Value: &openapi3.Parameter{
						Name:        "feed_version_key",
						In:          "path",
						Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,
						Required:    true,

						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
			},
		},
	},
}
