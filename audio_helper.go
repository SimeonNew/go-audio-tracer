package main

import "math"

func distance(a, b Vector) float64 {
	dx := a.x - b.x
	dy := a.y - b.y
	return math.Sqrt(dx*dx + dy*dy)
}

func normalize(v Vector) Vector {
	length := math.Sqrt(v.x*v.x + v.y*v.y)
	return Vector{v.x / length, v.y / length}
}

func reflect(incident, normal Vector) Vector {
	dot := incident.x*normal.x + incident.y*normal.y
	return Vector{incident.x - 2*dot*normal.x, incident.y - 2*dot*normal.y}
}

func rayWallIntersection(ray Ray, wall Wall) Vector {
	x1, y1 := wall.start.x, wall.start.y
	x2, y2 := wall.end.x, wall.end.y
	x3, y3 := ray.origin.x, ray.origin.y
	x4, y4 := ray.origin.x + ray.direction.x*1000, ray.origin.y + ray.direction.y*1000

	den := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if math.Abs(den) < 1e-8 {
		return Vector{math.Inf(1), math.Inf(1)}
	}

	t := ((x1-x3)*(y3-y4) - (y1-y3)*(x3-x4)) / den
	u := -((x1-x2)*(y1-y3) - (y1-y2)*(x1-x3)) / den

	if t >= 0 && t <= 1 && u >= 0 {
		return Vector{x1 + t*(x2-x1), y1 + t*(y2-y1)}
	}

	return Vector{math.Inf(1), math.Inf(1)}
}