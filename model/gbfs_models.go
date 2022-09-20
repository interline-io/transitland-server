package model

import (
	"github.com/interline-io/transitland-server/internal/gbfs"
)

type GbfsAlertTime struct {
	gbfs.AlertTime
}

type GbfsBrandAsset struct {
	gbfs.BrandAsset
}

type GbfsFeed struct {
	gbfs.GbfsFeed
}

func (g *GbfsFeed) SystemInformation() *GbfsSystemInformation     { return nil }
func (g *GbfsFeed) StationInformation() []*GbfsStationInformation { return nil }
func (g *GbfsFeed) StationStatus() []*GbfsStationStatus           { return nil }
func (g *GbfsFeed) Versions() []*GbfsSystemVersion                { return nil }
func (g *GbfsFeed) VehicleTypes() []*GbfsVehicleType              { return nil }
func (g *GbfsFeed) Bikes() []*GbfsFreeBikeStatus                  { return nil }
func (g *GbfsFeed) Regions() []*GbfsSystemRegion                  { return nil }
func (g *GbfsFeed) RentalHours() []*GbfsSystemHour                { return nil }
func (g *GbfsFeed) Calendars() []*GbfsSystemCalendar              { return nil }
func (g *GbfsFeed) Plans() []*GbfsSystemPricingPlan               { return nil }
func (g *GbfsFeed) Alerts() []*GbfsSystemAlert                    { return nil }
func (g *GbfsFeed) GeofencingZones() []*GbfsGeofenceZone          { return nil }

type GbfsFreeBikeStatus struct {
	gbfs.FreeBikeStatus
}

type GbfsGeofenceFeature struct {
	gbfs.GeofenceFeature
}

func (g *GbfsGeofenceFeature) Properties() []*GbfsGeofenceProperty {
	return nil
}

type GbfsGeofenceProperty struct {
	gbfs.GeofenceProperty
}

func (g *GbfsGeofenceProperty) Rules() []*GbfsGeofenceRule {
	return nil
}

type GbfsGeofenceRule struct {
	gbfs.GeofenceRule
}

type GbfsGeofenceZone struct {
	gbfs.GeofenceZone
}

func (g *GbfsGeofenceZone) Features() []*GbfsGeofenceFeature {
	return nil
}

type GbfsPlanPrice struct {
	gbfs.PlanPrice
}

type GbfsRentalApp struct {
	gbfs.RentalApp
}

type GbfsStationInformation struct {
	gbfs.StationInformation
}

type GbfsStationStatus struct {
	gbfs.StationStatus
}

func (g *GbfsStationStatus) VehicleTypesAvailable() []*GbfsVehicleTypeAvailable {
	return nil
}

func (g *GbfsStationStatus) VehicleDocksAvailable() []*GbfsVehicleDockAvailable {
	return nil
}

type GbfsSystemAlert struct {
	gbfs.SystemAlert
}

func (g *GbfsSystemAlert) Times() []*GbfsAlertTime {
	return nil
}

type GbfsSystemCalendar struct {
	gbfs.SystemCalendar
}

type GbfsSystemHour struct {
	gbfs.SystemHour
}

type GbfsSystemInformation struct {
	gbfs.SystemInformation
}

func (g *GbfsSystemInformation) BrandAssets() *GbfsBrandAsset {
	return nil
}

type GbfsSystemPricingPlan struct {
	gbfs.SystemPricingPlan
}

func (g *GbfsSystemPricingPlan) PerKmPricing() []*GbfsPlanPrice {
	return nil
}

func (g *GbfsSystemPricingPlan) PerMinPricing() []*GbfsPlanPrice {
	return nil
}

type GbfsSystemRegion struct {
	gbfs.SystemRegion
}

type GbfsSystemVersion struct {
	gbfs.SystemVersion
}

type GbfsVehicleDockAvailable struct {
	gbfs.VehicleDockAvailable
}

type GbfsVehicleType struct {
	gbfs.VehicleType
}

type GbfsVehicleTypeAvailable struct {
	gbfs.VehicleTypeAvailable
}
