package xy

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointRadiusBounds(t *testing.T) {
	tcs := []struct {
		p      Point
		radius float64
	}{
		{Point{Lon: -122.431297, Lat: 37.773972}, 1000},
		{Point{Lon: -149.88725143506608, Lat: 61.21262379506115}, 1000},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			a, b, err := PointRadiusBounds(tc.p, tc.radius)
			_ = err
			d := DistanceHaversine(a.Lon, a.Lat, b.Lon, b.Lat)
			fmt.Println("d:", d)
			assert.InDelta(t, tc.radius*math.Sqrt2*2, d, 10)
			t.Log("a:", a, "b:", b)
		})
	}
}
