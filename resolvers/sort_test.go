package resolvers

import (
	"fmt"
	"testing"

	"github.com/interline-io/transitland-server/model"
)

func Test_sortStopsByOnestopID(t *testing.T) {
	osids := []string{"x", "a", "q"}
	var stops []*model.Stop
	x := []string{"q", "x", "a", "a", "", "z", "e", ""}
	for i := 0; i < len(x); i++ {
		s := model.Stop{}
		if x[i] != "" {
			s.OnestopID = &x[i]
		}
		stops = append(stops, &s)
	}
	sorted := sortStopsByOnestopID(stops, osids)
	for _, v := range sorted {
		if v.OnestopID != nil {
			fmt.Println("v:", *v.OnestopID)
		} else {
			fmt.Println(nil)
		}
	}
}
