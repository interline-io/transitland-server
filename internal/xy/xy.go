package xy

import "math"

type Point struct {
	Lon float64
	Lat float64
}

var earthRadiusMetres float64 = 6371008

func deg2rad(v float64) float64 {
	return v * math.Pi / 180
}

func DistanceHaversine(lon1, lat1, lon2, lat2 float64) float64 {
	lon1 = deg2rad(lon1)
	lat1 = deg2rad(lat1)
	lon2 = deg2rad(lon2)
	lat2 = deg2rad(lat2)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	d := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Asin(math.Sqrt(d))
	return earthRadiusMetres * c
}

// PointRadiusBounds is a "close enough" bbox for a point with radius.
func PointRadiusBounds(p Point, d float64) (Point, Point, error) {
	latM := 111320.0
	lonM := 40075000.0 * (math.Cos(p.Lat*(math.Pi/180)) / 360)
	dlon := (d / lonM)
	dlat := (d / latM)
	a := Point{p.Lon - dlon, p.Lat - dlat}
	b := Point{p.Lon + dlon, p.Lat + dlat}
	return a, b, nil
}
