package model

// Some enum helpers

var specTypeMap = map[string]FeedSpecTypes{
	"gtfs":    FeedSpecTypesGtfs,
	"gtfs-rt": FeedSpecTypesGtfsRt,
	"gfbs":    FeedSpecTypesGbfs,
	"mds":     FeedSpecTypesMds,
}

func (f FeedSpecTypes) ToDBString() string {
	for k, v := range specTypeMap {
		if f == v {
			return k
		}
	}
	return ""
}

func (f FeedSpecTypes) FromDBString(s string) FeedSpecTypes {
	a, ok := specTypeMap[s]
	if !ok {
		panic("cannot convert")
	}
	return a
}
