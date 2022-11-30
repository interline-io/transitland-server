package find

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func licenseFilter(license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	return licenseFilterTable("current_feeds", license, qView)
}

func licenseFilterTable(t string, license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	if license == nil {
		return qView
	}
	licenseField := func(t string, col string) string {
		return fmt.Sprintf("%s.license->>'%s'", az09(t), az09(col))
	}
	if v := license.CommercialUseAllowed; v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{licenseField(t, "commercial_use_allowed"): "yes"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{licenseField(t, "commercial_use_allowed"): "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.NotEq{licenseField(t, "commercial_use_allowed"): "no"})
	}

	if v := license.ShareAlikeOptional; v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{licenseField(t, "share_alike_optional"): "yes"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{licenseField(t, "share_alike_optional"): "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.NotEq{licenseField(t, "share_alike_optional"): "no"})
	}

	if v := license.CreateDerivedProduct; v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{licenseField(t, "create_derived_product"): "yes"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{licenseField(t, "create_derived_product"): "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.NotEq{licenseField(t, "create_derived_product"): "no"})
	}

	if v := license.RedistributionAllowed; v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{licenseField(t, "redistribution_allowed"): "yes"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{licenseField(t, "redistribution_allowed"): "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.NotEq{licenseField(t, "redistribution_allowed"): "no"})
	}

	if v := license.UseWithoutAttribution; v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{licenseField(t, "use_without_attribution"): "yes"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{licenseField(t, "use_without_attribution"): "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.NotEq{licenseField(t, "use_without_attribution"): "no"})
	}
	return qView
}
