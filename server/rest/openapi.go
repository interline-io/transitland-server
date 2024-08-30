package rest

import (
	oa "github.com/getkin/kin-openapi/openapi3"
)

type param = oa.Parameter
type pref = oa.ParameterRef
type sref = oa.SchemaRef

type RequestInfo struct {
	Path     string
	PathItem *oa.PathItem
}

func newSchema(st string, format string, enum []any) *oa.Schema {
	return &oa.Schema{
		Type:   &oa.Types{st},
		Format: format,
		Enum:   enum,
	}
}

var ParameterComponents = oa.ParametersMap{
	"adm0IsoParam": &pref{
		Value: &param{
			Name:        "adm0_iso",
			In:          "query",
			Description: `Search by country 2 letter ISO 3166 code`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm0NameParam": &pref{
		Value: &param{
			Name:        "adm0_name",
			In:          "query",
			Description: `Search by country name`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm1IsoParam": &pref{
		Value: &param{
			Name:        "adm1_iso",
			In:          "query",
			Description: `Search by state/province/division ISO 3166-2 code`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"adm1NameParam": &pref{
		Value: &param{
			Name:        "adm1_name",
			In:          "query",
			Description: `Search by state/province/division name`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"afterParam": &pref{
		Value: &param{
			Name:        "after",
			In:          "query",
			Description: `Pagination cursor value. This should be treated as an opaque value created by the server and returned as the link to the next result page, which may be empty. For historical reasons, this is based on the integer record ID values, but that should not be assumed to be the case in the future.`,
			Schema: &sref{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"bboxParam": &pref{
		Value: &param{
			Name:        "bbox",
			In:          "query",
			Description: `Geographic search using a bounding box, with coordinates in (min_lon, min_lat, max_lon, max_lat) order as a comma separated string`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"cityNameParam": &pref{
		Value: &param{
			Name:        "city_name",
			In:          "query",
			Description: `Search by city name`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"feedParam": &pref{
		Value: &param{
			Name:        "feed_onestop_id",
			In:          "query",
			Description: `Search for records in this feed`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"formatParam": &pref{
		Value: &param{
			Name:        "format",
			In:          "query",
			Description: `Response format`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"json", "geojson", "geojsonl", "png"}),
			},
		},
	},
	"idParam": &pref{
		Value: &param{
			Name:        "id",
			In:          "query",
			Description: `Search for a specific internal ID`,
			Schema: &sref{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"includeAlertsParam": &pref{
		Value: &param{
			Name:        "include_alerts",
			In:          "query",
			Description: `Include alerts from GTFS Realtime feeds`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"true", "false"}),
			},
		},
	},
	"latParam": &pref{
		Value: &param{
			Name:        "lat",
			In:          "query",
			Description: `Latitude`,
			Schema: &sref{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"licenseCommercialUseAllowedParam": &pref{
		Value: &param{
			Name:        "license_commercial_use_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'commercial_use_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseCreateDerivedProductParam": &pref{
		Value: &param{
			Name:        "license_create_derived_product",
			In:          "query",
			Description: `Filter entities by feed license 'create_derived_product' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseRedistributionAllowedParam": &pref{
		Value: &param{
			Name:        "license_redistribution_allowed",
			In:          "query",
			Description: `Filter entities by feed license 'redistribution_allowed' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseShareAlikeOptionalParam": &pref{
		Value: &param{
			Name:        "license_share_alike_optional",
			In:          "query",
			Description: `Filter entities by feed license 'share_alike_optional' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"licenseUseWithoutAttributionParam": &pref{
		Value: &param{
			Name:        "license_use_without_attribution",
			In:          "query",
			Description: `Filter entities by feed license 'use_without_attribution' value. Please see Source Feed concept for details on license values. 'exclude_no' is equivalent to 'yes' and 'unknown'.`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"yes", "no", "unknown", "exclude_no"}),
			},
		},
	},
	"limitParam": &pref{
		Value: &param{
			Name:        "limit",
			In:          "query",
			Description: `Maximum number of records to return`,
			Schema: &sref{
				Value: newSchema("integer", "int32", nil),
			},
		},
	},
	"lonParam": &pref{
		Value: &param{
			Name:        "lon",
			In:          "query",
			Description: `Longitude`,
			Schema: &sref{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"onestopParam": &pref{
		Value: &param{
			Name:        "onestop_id",
			In:          "query",
			Description: `Search for a specific Onestop ID`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"radiusParam": &pref{
		Value: &param{
			Name:        "radius",
			In:          "query",
			Description: `Search radius (meters); requires lat and lon`,
			Schema: &sref{
				Value: newSchema("number", "", nil),
			},
		},
	},
	"relativeDateParam": &pref{
		Value: &param{
			Name:        "relative_date",
			In:          "query",
			Description: `Search for departures on a relative date label, e.g. TODAY, TUESDAY, NEXT_WEDNESDAY`,
			Schema: &sref{
				Value: newSchema("string", "", []any{"TODAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY", "NEXT_MONDAY", "NEXT_TUESDAY", "NEXT_WEDNESDAY", "NEXT_THURSDAY", "NEXT_FRIDAY", "NEXT_SATURDAY", "NEXT_SUNDAY"}),
			},
		},
	},
	"searchParam": &pref{
		Value: &param{
			Name:        "search",
			In:          "query",
			Description: `Full text search`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
	"sha1Param": &pref{
		Value: &param{
			Name:        "feed_version_sha1",
			In:          "query",
			Description: `Search for records in this feed version`,
			Schema: &sref{
				Value: newSchema("string", "", nil),
			},
		},
	},
}

// var PathItems = map[string]*oa.PathItem{
// 	"/feeds/{feed_key}/download_latest_feed_version": {
// 		Get: &oa.Operation{
// 			Summary:     "",
// 			Description: `Download latest feed version for this feed`,
// 			Parameters: oa.Parameters{
// 				&pref{
// 					Value: &param{
// 						Name:        "feed_key",
// 						In:          "path",
// 						Description: `Feed lookup key; can be an integer ID or a Onestop ID`,
// 						Required:    true,
// 						Schema: &sref{
// 							Value: newSchema("string", "", nil),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	},
// 	"/feed_versions/{feed_version_key}/download": {
// 		Get: &oa.Operation{
// 			Summary:     "",
// 			Description: `Download this feed version`,
// 			Parameters: oa.Parameters{
// 				&pref{
// 					Value: &param{
// 						Name:        "feed_version_key",
// 						In:          "path",
// 						Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,
// 						Required:    true,
// 						Schema: &sref{
// 							Value: newSchema("string", "", nil),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	},
// }
