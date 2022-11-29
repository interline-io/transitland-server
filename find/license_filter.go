package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func licenseFilter(license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	if license == nil {
		return qView
	}
	if v := license.CommercialUseAllowed; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'commercial_use_allowed'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'commercial_use_allowed'": "no"})
	}
	if v := license.ShareAlikeOptional; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'share_alike_optional'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'share_alike_optional'": "no"})
	}
	if v := license.CreateDerivedProduct; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'create_derived_product'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'create_derived_product'": "no"})
	}
	return qView
}

func licenseFilterT(license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	if license == nil {
		return qView
	}
	if v := license.CommercialUseAllowed; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"t.license->>'commercial_use_allowed'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"t.license->>'commercial_use_allowed'": "no"})
	}
	if v := license.ShareAlikeOptional; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"t.license->>'share_alike_optional'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"t.license->>'share_alike_optional'": "no"})
	}
	if v := license.CreateDerivedProduct; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"t.license->>'create_derived_product'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"t.license->>'create_derived_product'": "no"})
	}
	return qView
}
