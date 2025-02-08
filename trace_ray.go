package main

import (
	"math"
)


// Add adds two vectors and returns the result.
func (v Vector) add(other Vector) Vector {
	return Vector{
		x: v.x + other.x,
		y: v.y + other.y,
	}
}


func (g *Game) traceRay(ray Ray, intensity float64, bounces int, rayIndex int) {
	if bounces == 0 || intensity < 0.01 {
		return
	}

	closestIntersection := Vector{math.Inf(1), math.Inf(1)}
	closestWall := -1
	minDist := math.Inf(1)
	lastIntersection := g.rayPathPoints[rayIndex][len(g.rayPathPoints[rayIndex])-1].position

	for i, wall := range g.walls {
		intersection := rayWallIntersection(ray, wall, lastIntersection)
		if intersection.x != math.Inf(1) && intersection.y != math.Inf(1) {
			dist := distance(ray.origin, intersection)
			if dist < minDist {
				minDist = dist
				closestIntersection = intersection
				closestWall = i
			}
		}
	}

	perpendicularDist, distanceToSource := distanceFromPointToLine(ray, g.audioSource.position)
	if perpendicularDist < proximityThreshold && distanceToSource != -1 && distanceToSource < minDist {
		g.addAudioPaths(ray, intensity)
	} else if closestWall == -1 {
		edgeIntersection := extendRayToScreenEdge(ray)
		g.rayPathPoints[rayIndex] = append(g.rayPathPoints[rayIndex], RayPathPoint{edgeIntersection, intensity})
		return
	}

	g.rayPathPoints[rayIndex] = append(g.rayPathPoints[rayIndex], RayPathPoint{closestIntersection, intensity})
	distanceOriginIntersection := distance(ray.origin, closestIntersection)
	intensity = intensity / (1 + distanceOriginIntersection*distanceOriginIntersection/10000)
	wall := g.walls[closestWall]

	for _, edge := range g.wallEdges {
		if !edge.isCorner {
			edgeDist := distance(closestIntersection, edge.position)
			if edgeDist < 10.0 {
				g.handleDiffraction(ray, wall, edge, closestIntersection, intensity, bounces)
				return
			}
		}
	}

	reflectedIntensity := intensity * (1.0 - wall.properties.transparency) * (1.0 - wall.properties.absorption)
	if reflectedIntensity > 0.01 {
		wallDirection := Vector{wall.end.x - wall.start.x, wall.end.y - wall.start.y}
		wallNormal := Vector{-wallDirection.y, wallDirection.x}.normalize()

		// Reflect the ray and add randomness
		reflectedDirection := reflect(ray.direction, wallNormal)
		reflectedRay := Ray{closestIntersection, reflectedDirection}

		newRayIndex := len(g.rayPathPoints)
		g.rayPathPoints = append(g.rayPathPoints, []RayPathPoint{{closestIntersection, reflectedIntensity}})
		g.traceRay(reflectedRay, reflectedIntensity, bounces-1, newRayIndex)
	}

	transmittedIntensity := intensity * wall.properties.transparency
	if transmittedIntensity > 0.01 {
		transmittedDirection := ray.direction
		transmittedRay := Ray{closestIntersection, transmittedDirection}

		newRayIndex := len(g.rayPathPoints)
		g.rayPathPoints = append(g.rayPathPoints, []RayPathPoint{{closestIntersection, transmittedIntensity}})
		g.traceRay(transmittedRay, transmittedIntensity, bounces-1, newRayIndex)
	}
}

