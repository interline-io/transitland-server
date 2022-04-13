package resolvers

import (
	"sort"

	"github.com/interline-io/transitland-server/model"
)

// sortStopsByOnestopID sorts stops according to the order in provided list
func sortStopsByOnestopID(ents []*model.Stop, osids []string) []*model.Stop {
	if len(osids) == 0 {
		return ents
	}
	instops := append([]*model.Stop{}, ents...)
	var out []*model.Stop
	for _, osid := range osids {
		for j, ent := range instops {
			if ent != nil && ent.OnestopID != nil && *ent.OnestopID == osid {
				out = append(out, ent)
				instops[j] = nil
			}
		}
	}
	// sort remaining stops
	var remain []*model.Stop
	var nosids []*model.Stop
	for _, ent := range instops {
		if ent == nil {
			continue
		}
		if ent.OnestopID == nil {
			nosids = append(nosids, ent)
		} else {
			remain = append(remain, ent)
		}
	}
	sort.Slice(remain, func(i, j int) bool {
		// these are non null and non-null onestopid
		return *remain[i].OnestopID < *remain[j].OnestopID
	})
	out = append(out, remain...)
	out = append(out, nosids...)
	return out
}
