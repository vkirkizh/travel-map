package flights

import "math"

const earthRadiusKM = 6371.0

func DistanceKM(from Point, to Point) int {
	lat1 := degreesToRadians(from.Lat)
	lng1 := degreesToRadians(from.Lng)
	lat2 := degreesToRadians(to.Lat)
	lng2 := degreesToRadians(to.Lng)

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(math.Round(earthRadiusKM * c))
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
