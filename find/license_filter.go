package find

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func licenseFilter(license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	return licenseFilterTable("current_feeds", license, qView)
}

func licenseCheck(t string, col string, v *model.LicenseValue, qView sq.SelectBuilder) sq.SelectBuilder {
	c := fmt.Sprintf("%s.license->>'%s'", az09(t), az09(col))
	if v == nil {
	} else if *v == model.LicenseValueYes {
		qView = qView.Where(sq.Eq{c: "yes"})
	} else if *v == model.LicenseValueUnknown {
		qView = qView.Where(sq.Eq{c: "unknown"})
	} else if *v == model.LicenseValueNo {
		qView = qView.Where(sq.Eq{c: "no"})
	} else if *v == model.LicenseValueExcludeNo {
		qView = qView.Where(sq.Or{sq.Eq{c: "yes"}, sq.Eq{c: "unknown"}, sq.Eq{c: nil}})
	}
	return qView
}

func licenseFilterTable(t string, license *model.LicenseFilter, qView sq.SelectBuilder) sq.SelectBuilder {
	if license == nil {
		return qView
	}
	qView = licenseCheck(t, "commercial_use_allowed", license.CommercialUseAllowed, qView)
	qView = licenseCheck(t, "share_alike_optional", license.ShareAlikeOptional, qView)
	qView = licenseCheck(t, "create_derived_product", license.CreateDerivedProduct, qView)
	qView = licenseCheck(t, "redistribution_allowed", license.RedistributionAllowed, qView)
	qView = licenseCheck(t, "use_without_attribution", license.UseWithoutAttribution, qView)
	return qView
}
