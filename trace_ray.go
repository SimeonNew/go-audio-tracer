package main

import "math"

func (g *Game) traceRay(ray Ray, intensity float64, bounces int) {
	if bounces == 0 || intensity < 0.01 {
		return
	}

	closestIntersection := Vector{math.Inf(1), math.Inf(1)}
	closestWall := -1
	minDist := math.Inf(1)

	for i, wall := range g.walls {
		intersection := rayWallIntersection(ray, wall)
		if intersection.x != math.Inf(1) && intersection.y != math.Inf(1) {
			dist := distance(ray.origin, intersection)
			if dist < minDist {
				minDist = dist
				closestIntersection = intersection
				closestWall = i
			}
		}
	}

	if closestWall == -1 {
		return
	}

	wall := g.walls[closestWall]
	g.intersections = append(g.intersections, closestIntersection)

	// Add path to left and right ears
	leftDelay := distance(closestIntersection, g.listener.leftEar) / speedOfSound
	rightDelay := distance(closestIntersection, g.listener.rightEar) / speedOfSound

	g.leftPaths = append(g.leftPaths, AudioPath{
		source:    g.audioSource,
		delay:     leftDelay,
		amplitude: intensity,
		direction: normalize(Vector{g.listener.leftEar.x - closestIntersection.x, g.listener.leftEar.y - closestIntersection.y}),
	})

	g.rightPaths = append(g.rightPaths, AudioPath{
		source:    g.audioSource,
		delay:     rightDelay,
		amplitude: intensity,
		direction: normalize(Vector{g.listener.rightEar.x - closestIntersection.x, g.listener.rightEar.y - closestIntersection.y}),
	})

	reflectedIntensity := intensity * (1.0 - wall.properties.transparency) * (1.0 - wall.properties.absorption)
	if reflectedIntensity > 0.01 {
		wallDirection := Vector{wall.end.x - wall.start.x, wall.end.y - wall.start.y}
		wallNormal := Vector{-wallDirection.y, wallDirection.x}
		wallNormal = normalize(wallNormal)

		reflectedDirection := reflect(ray.direction, wallNormal)
		reflectedRay := Ray{closestIntersection, reflectedDirection}
		g.traceRay(reflectedRay, reflectedIntensity, bounces-1)
	}

	transmittedIntensity := intensity * wall.properties.transparency
	if transmittedIntensity > 0.01 {
		transmittedRay := Ray{closestIntersection, ray.direction}
		g.traceRay(transmittedRay, transmittedIntensity, bounces-1)
	}
}