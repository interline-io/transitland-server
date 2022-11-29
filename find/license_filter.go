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
	if v := license.RedistributionAllowed; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'redistribution_allowed'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'redistribution_allowed'": "no"})
	}
	if v := license.UseWithoutAttribution; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'use_without_attribution'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"current_feeds.license->>'use_without_attribution'": "no"})
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
	if v := license.RedistributionAllowed; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"t.license->>'redistribution_allowed'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"t.license->>'redistribution_allowed'": "no"})
	}
	if v := license.UseWithoutAttribution; v == nil {
	} else if *v {
		qView = qView.Where(sq.Eq{"t.license->>'use_without_attribution'": "yes"})
	} else {
		qView = qView.Where(sq.Eq{"t.license->>'use_without_attribution'": "no"})
	}
	return qView
}
