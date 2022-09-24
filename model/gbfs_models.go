package model

import (
	"github.com/interline-io/transitland-server/internal/gbfs"
)

type GbfsAlertTime struct {
	gbfs.AlertTime
}

type GbfsBrandAsset struct {
	*gbfs.BrandAsset
}

type GbfsFeed struct {
	*gbfs.GbfsFeed
}

func (g *GbfsFeed) SystemInformation() *GbfsSystemInformation {
	if g.GbfsFeed.SystemInformation == nil {
		return nil
	}
	return &GbfsSystemInformation{
		Feed:              g,
		SystemInformation: g.GbfsFeed.SystemInformation,
	}
}

func (g *GbfsFeed) StationInformation() []*GbfsStationInformation {
	var ret []*GbfsStationInformation
	for _, s := range g.GbfsFeed.StationInformation {
		if s == nil {
			continue
		}
		ret = append(ret, &GbfsStationInformation{
			Feed:               g,
			StationInformation: s,
		})
	}
	return ret
}

func (g *GbfsFeed) VehicleTypes() []*GbfsVehicleType {
	return nil
}

func (g *GbfsFeed) RentalHours() []*GbfsSystemHour {
	return nil
}

func (g *GbfsFeed) Calendars() []*GbfsSystemCalendar {
	return nil
}

func (g *GbfsFeed) GeofencingZones() []*GbfsGeofenceZone {
	return nil
}

func (g *GbfsFeed) Alerts() []*GbfsSystemAlert {
	return nil
}

type GbfsFreeBikeStatus struct {
	Feed *GbfsFeed
	*gbfs.FreeBikeStatus
}

func (g *GbfsFreeBikeStatus) Station() *GbfsStationInformation {
	if g.Feed != nil {
		for _, s := range g.Feed.StationInformation() {
			if s == nil {
				continue
			}
			if s.StationID.Val == g.StationID.Val {
				return s
			}
		}
	}
	return nil
}

func (g *GbfsFreeBikeStatus) HomeStation() *GbfsStationInformation {
	if g.Feed != nil {
		for _, s := range g.Feed.StationInformation() {
			if s == nil {
				continue
			}
			if s.StationID.Val == g.HomeStationID.Val {
				return s
			}
		}
	}
	return nil
}

func (g *GbfsFreeBikeStatus) PricingPlan() *GbfsSystemPricingPlan {
	if g.Feed != nil {
		for _, s := range g.Feed.Plans {
			if s == nil {
				continue
			}
			if s.PlanID.Val == g.PricingPlanID.Val {
				return &GbfsSystemPricingPlan{SystemPricingPlan: s}
			}
		}
	}
	return nil
}

func (g *GbfsFreeBikeStatus) VehicleType() *GbfsVehicleType {
	return nil
}

func (g *GbfsFreeBikeStatus) RentalUris() *GbfsRentalUris {
	if g.RentalURIs == nil {
		return nil
	}
	return &GbfsRentalUris{RentalURIs: g.RentalURIs}
}

type GbfsGeofenceFeature struct {
	*gbfs.GeofenceFeature
}

type GbfsVehicleAssets struct {
	*gbfs.VehicleAssets
}

func (g *GbfsGeofenceFeature) Properties() []*GbfsGeofenceProperty {
	return nil
}

type GbfsGeofenceProperty struct {
	*gbfs.GeofenceProperty
}

func (g *GbfsGeofenceProperty) Rules() []*GbfsGeofenceRule {
	return nil
}

type GbfsGeofenceRule struct {
	*gbfs.GeofenceRule
}

func (g *GbfsGeofenceRule) VehicleType() *GbfsVehicleType {
	return nil
}

type GbfsGeofenceZone struct {
	*gbfs.GeofenceZone
}

func (g *GbfsGeofenceZone) Features() []*GbfsGeofenceFeature {
	return nil
}

type GbfsPlanPrice struct {
	*gbfs.PlanPrice
}

type GbfsRentalApps struct {
	*gbfs.RentalApps
}

func (g *GbfsRentalApps) Android() *GbfsRentalApp {
	if g.RentalApps == nil || g.RentalApps.Android == nil {
		return nil
	}
	return &GbfsRentalApp{RentalApp: g.RentalApps.Android}
}

func (g *GbfsRentalApps) Ios() *GbfsRentalApp {
	if g.RentalApps == nil || g.RentalApps.IOS == nil {
		return nil
	}
	return &GbfsRentalApp{RentalApp: g.RentalApps.IOS}
}

type GbfsRentalApp struct {
	*gbfs.RentalApp
}

type GbfsStationInformation struct {
	Feed *GbfsFeed
	*gbfs.StationInformation
}

func (g *GbfsStationInformation) Region() *GbfsSystemRegion {
	return nil
}

func (g *GbfsStationInformation) Status() *GbfsStationStatus {
	return nil
}

type GbfsStationStatus struct {
	Feed *GbfsFeed
	*gbfs.StationStatus
}

func (g *GbfsStationStatus) VehicleTypesAvailable() []*GbfsVehicleTypeAvailable {
	return nil
}

func (g *GbfsStationStatus) VehicleDocksAvailable() []*GbfsVehicleDockAvailable {
	return nil
}

type GbfsSystemAlert struct {
	*gbfs.SystemAlert
}

func (g *GbfsSystemAlert) Times() []*GbfsAlertTime {
	return nil
}

type GbfsSystemCalendar struct {
	*gbfs.SystemCalendar
}

type GbfsSystemHour struct {
	*gbfs.SystemHour
}

type GbfsSystemInformation struct {
	Feed *GbfsFeed
	*gbfs.SystemInformation
}

func (g *GbfsSystemInformation) BrandAssets() *GbfsBrandAsset {
	if g.SystemInformation == nil || g.SystemInformation.BrandAssets == nil {
		return nil
	}
	return &GbfsBrandAsset{BrandAsset: g.SystemInformation.BrandAssets}
}

func (g *GbfsSystemInformation) RentalApps() *GbfsRentalApps {
	if g.SystemInformation == nil || g.SystemInformation.RentalApps == nil {
		return nil
	}
	return &GbfsRentalApps{RentalApps: g.SystemInformation.RentalApps}
}

type GbfsSystemPricingPlan struct {
	*gbfs.SystemPricingPlan
}

func (g *GbfsSystemPricingPlan) PerKmPricing() []*GbfsPlanPrice {
	return nil
}

func (g *GbfsSystemPricingPlan) PerMinPricing() []*GbfsPlanPrice {
	return nil
}

type GbfsSystemRegion struct {
	*gbfs.SystemRegion
}

type GbfsSystemVersion struct {
	*gbfs.SystemVersion
}

type GbfsVehicleDockAvailable struct {
	*gbfs.VehicleDockAvailable
}

func (g *GbfsVehicleDockAvailable) VehicleTypes() []*GbfsVehicleType {
	return nil
}

type GbfsVehicleType struct {
	*gbfs.VehicleType
}

func (g *GbfsVehicleType) DefaultPricingPlan() *GbfsSystemPricingPlan {
	return nil
}

func (g *GbfsVehicleType) PricingPlans() []*GbfsSystemPricingPlan {
	return nil
}

func (g *GbfsVehicleType) RentalUris() *GbfsRentalUris {
	return nil
}

func (g *GbfsVehicleType) VehicleAssets() *GbfsVehicleAssets {
	return nil
}

type GbfsVehicleTypeAvailable struct {
	*gbfs.VehicleTypeAvailable
}

func (g *GbfsVehicleTypeAvailable) VehicleType() *GbfsVehicleType {
	return nil
}

type GbfsRentalUris struct {
	*gbfs.RentalURIs
}
