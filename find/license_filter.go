package find

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func licenseFilter(license *model.LicenseFilter, q sq.SelectBuilder) sq.SelectBuilder {
	return licenseFilterTable("current_feeds", license, q)
}

func licenseCheck(t string, col string, v *model.LicenseValue, q sq.SelectBuilder) sq.SelectBuilder {
	c := fmt.Sprintf("%s.license->>'%s'", az09(t), az09(col))
	if v == nil {
	} else if *v == model.LicenseValueYes {
		q = q.Where(sq.Eq{c: "yes"})
	} else if *v == model.LicenseValueUnknown {
		q = q.Where(sq.Eq{c: "unknown"})
	} else if *v == model.LicenseValueNo {
		q = q.Where(sq.Eq{c: "no"})
	} else if *v == model.LicenseValueExcludeNo {
		q = q.Where(sq.Or{sq.Eq{c: "yes"}, sq.Eq{c: "unknown"}, sq.Eq{c: nil}})
	}
	return q
}

func licenseFilterTable(t string, license *model.LicenseFilter, q sq.SelectBuilder) sq.SelectBuilder {
	if license == nil {
		return q
	}
	q = licenseCheck(t, "commercial_use_allowed", license.CommercialUseAllowed, q)
	q = licenseCheck(t, "share_alike_optional", license.ShareAlikeOptional, q)
	q = licenseCheck(t, "create_derived_product", license.CreateDerivedProduct, q)
	q = licenseCheck(t, "redistribution_allowed", license.RedistributionAllowed, q)
	q = licenseCheck(t, "use_without_attribution", license.UseWithoutAttribution, q)
	return q
}
