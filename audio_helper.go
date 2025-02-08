package main

import (
	"math"
)

func distance(a, b Vector) float64 {
	dx := a.x - b.x
	dy := a.y - b.y
	return math.Sqrt(dx*dx + dy*dy)
}
func cross(a, b Vector) Vector {
    return Vector{x: a.x * b.y - a.y * b.x, y: a.y * b.x - a.x * b.y}
}

func (v Vector) normalize () Vector {
	length := math.Sqrt(v.x*v.x + v.y*v.y)
	if length == 0 {
		return Vector{0, 0}
	}
	return Vector{v.x / length, v.y / length}
}

func (v Vector) length () float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}
func reflect(incident, normal Vector) Vector {
	dot := incident.x*normal.x + incident.y*normal.y
	return Vector{incident.x - 2*dot*normal.x, incident.y - 2*dot*normal.y}
}
func dot(a, b Vector) float64 {
	return a.x*b.x + a.y*b.y
}

// rayWallIntersection computes the intersection of a ray with a wall, given the
// ray's origin and direction, the wall's start and end points, and the last
// intersection point of the ray with the wall. If the ray doesn't intersect
// with the wall, or if the intersection is too close to the last intersection
// point, returns a vector with Inf values. Otherwise, returns the intersection
// point.
func rayWallIntersection(ray Ray, wall Wall, lastIntersection Vector) Vector {
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
		intersection := Vector{x1 + t*(x2-x1), y1 + t*(y2-y1)}
		// Check if the intersection is too close to the last intersection
		if distance(intersection, lastIntersection) < 0.01 { // small tolerance to avoid repeated bounces
			return Vector{math.Inf(1), math.Inf(1)} // No valid intersection
		}
		return intersection
	}

	return Vector{math.Inf(1), math.Inf(1)}
}

func extendRayToScreenEdge(ray Ray) Vector {
    // Calculate intersections with screen edges
    tTop := -ray.origin.y / ray.direction.y
    tBottom := (screenHeight - ray.origin.y) / ray.direction.y
    tLeft := -ray.origin.x / ray.direction.x
    tRight := (screenWidth - ray.origin.x) / ray.direction.x

    // Find the smallest positive t
    t := math.Inf(1)
    if tTop > 0 && tTop < t {
        t = tTop
    }
    if tBottom > 0 && tBottom < t {
        t = tBottom
    }
    if tLeft > 0 && tLeft < t {
        t = tLeft
    }
    if tRight > 0 && tRight < t {
        t = tRight
    }

    // Calculate the intersection point
    return Vector{
        x: ray.origin.x + t*ray.direction.x,
        y: ray.origin.y + t*ray.direction.y,
    }
}


func distanceFromPointToLine(ray Ray, point Vector) (float64, float64) {
    lineDir := ray.direction

    // Vector from ray origin to the point
    v1 := Vector{point.x - ray.origin.x, point.y - ray.origin.y}

    // Perpendicular distance from point to line (same as before)
    numerator := math.Abs(v1.x*lineDir.y - v1.y*lineDir.x)
    denominator := math.Sqrt(lineDir.x*lineDir.x + lineDir.y*lineDir.y)
    perpendicularDist := numerator / denominator

    // Projection of v1 onto the ray direction to get the distance along the ray
    dotProduct := dot(v1, lineDir)

    if dotProduct < 0 {
        // If dotProduct < 0, the closest point is behind the ray origin
        return perpendicularDist, -1
    }

    closestPoint := Vector{
        ray.origin.x + dotProduct*lineDir.x,
        ray.origin.y + dotProduct*lineDir.y,
    }

    // Distance from the origin to the closest point on the ray
    distToClosestPoint := math.Sqrt(math.Pow(closestPoint.x-ray.origin.x, 2) + math.Pow(closestPoint.y-ray.origin.y, 2))

    return perpendicularDist, distToClosestPoint
}



func (g *Game) addAudioPaths(ray Ray, intensity float64) {
    leftDelay := distance(ray.origin, g.listener.leftEar) / speedOfSound
    rightDelay := distance(ray.origin, g.listener.rightEar) / speedOfSound

    g.leftPaths = append(g.leftPaths, AudioPath{
        source:    g.audioSource,
        delay:     leftDelay,
        amplitude: intensity,
        direction: ray.direction,
    })

    g.rightPaths = append(g.rightPaths, AudioPath{
        source:    g.audioSource,
        delay:     rightDelay,
        amplitude: intensity,
        direction: ray.direction,
    })
}
func (g *Game) handleDiffraction(ray Ray, wall Wall, edge WallEdge, hitPoint Vector, intensity float64, bounces int) {
    const (
        numDiffractedRays = 5  // Increased for smoother wave pattern
        baseIntensity      = 0.4  // Base intensity factor for diffracted rays
    )

    // Calculate distance to the wall edge
    distanceToEdge := Vector{hitPoint.x - edge.position.x, hitPoint.y - edge.position.y}.length()

    // Check if the wall allows sound to pass through
    if wall.properties.transparency > 0 {
        // Calculate the intensity reduction based on transparency
        intensity *= (1.0 - wall.properties.transparency)
        
        if intensity < 0.01 {
            return // If intensity is too low, stop further processing
        }

        // Handle diffraction for walls with transparency
        for i := 0; i < numDiffractedRays; i++ {
            t := float64(i) / float64(numDiffractedRays-1)
            angle := normalizeAngle(math.Atan2(ray.direction.y, ray.direction.x) + (t-0.5)*math.Pi)

            // Calculate intensity considering wall absorption
            diffractedIntensity := intensity * baseIntensity * (1.0 - wall.properties.absorption)

            // Add distance-based attenuation
            distanceAttenuation := 1.0 / (1.0 + math.Pow(distanceToEdge, 0.5))
            diffractedIntensity *= distanceAttenuation

            if diffractedIntensity > 0.01 {
                newDirection := Vector{math.Cos(angle), math.Sin(angle)}.normalize()
                diffractedRay := Ray{origin: hitPoint, direction: newDirection}
                newRayIndex := len(g.rayPathPoints)
                g.rayPathPoints = append(g.rayPathPoints, []RayPathPoint{{hitPoint, diffractedIntensity}})
                g.traceRay(diffractedRay, diffractedIntensity, bounces-1, newRayIndex)
            }
        }
        return // Exit early after handling transparency
    }

    // If the wall does not allow sound to pass through, handle edge diffraction
    if edge.isCorner {
        angle1 := normalizeAngle(math.Atan2(edge.normal1.y, edge.normal1.x))
        angle2 := normalizeAngle(math.Atan2(edge.normal2.y, edge.normal2.x))

        // Ensure proper angle ordering
        if normalizeAngle(angle2-angle1) < 0 {
            angle1, angle2 = angle2, angle1
        }

        // Generate wave-like diffraction pattern
        for i := 0; i < numDiffractedRays; i++ {
            t := float64(i) / float64(numDiffractedRays-1)
            phase := t * 2 * math.Pi

            // Calculate angle using wave function
            spreadFactor := 0.8 // Controls the spread of the wave pattern
            angle := normalizeAngle(angle1 + t*(normalizeAngle(angle2-angle1)))

            // Apply wave modulation to intensity
            waveIntensity := math.Cos(phase) * 0.5 + 0.5 // Normalize to 0-1 range
            diffractedIntensity := intensity * baseIntensity * waveIntensity * (1.0 - wall.properties.absorption)

            // Add distance-based attenuation
            distanceAttenuation := 1.0 / (1.0 + math.Pow(distanceToEdge, 0.5))
            diffractedIntensity *= distanceAttenuation

            if diffractedIntensity > 0.01 {
                waveOffset := math.Sin(phase) * spreadFactor
                finalAngle := normalizeAngle(angle + waveOffset)
                newDirection := Vector{math.Cos(finalAngle), math.Sin(finalAngle)}.normalize()

                diffractedRay := Ray{origin: edge.position, direction: newDirection}
                newRayIndex := len(g.rayPathPoints)
                g.rayPathPoints = append(g.rayPathPoints, []RayPathPoint{{edge.position, diffractedIntensity}})
                g.traceRay(diffractedRay, diffractedIntensity, bounces-1, newRayIndex)
            }
        }
    } else {
        // Handle edge diffraction for straight walls
        wallAngle := normalizeAngle(math.Atan2(edge.normal1.y, edge.normal1.x))

        // Create semicircular wave pattern
        for i := 0; i < numDiffractedRays; i++ {
            t := float64(i) / float64(numDiffractedRays-1)
            phase := t * math.Pi // Semicircular distribution

            // Calculate base angle with smooth distribution
            angle := normalizeAngle(wallAngle + (t-0.5)*math.Pi) // Full half-circle spread

            // Wave modulation
            waveIntensity := math.Pow(math.Cos(phase), 2) // Smoother falloff
            diffractedIntensity := intensity * baseIntensity * waveIntensity * (1.0 - wall.properties.absorption)

            // Enhanced distance-based attenuation
            distanceFactor := 1.0 / (1.0 + math.Pow(distanceToEdge, 0.7))
            diffractedIntensity *= distanceFactor

            if diffractedIntensity > 0.01 {
                waveOffset := math.Sin(phase) * 0.3 // Reduced wave amplitude for more natural spread
                finalAngle := normalizeAngle(angle + waveOffset)
                newDirection := Vector{math.Cos(finalAngle), math.Sin(finalAngle)}.normalize()

                diffractedRay := Ray{origin: edge.position, direction: newDirection}
                newRayIndex := len(g.rayPathPoints)
                g.rayPathPoints = append(g.rayPathPoints, []RayPathPoint{{edge.position, diffractedIntensity}})
                g.traceRay(diffractedRay, diffractedIntensity, bounces-1, newRayIndex)
            }
        }
    }
}


func normalizeAngle(angle float64) float64 {
    // First bring into range [0, 2π]
    angle = math.Mod(angle, 2*math.Pi)
    if angle < 0 {
        angle += 2 * math.Pi
    }
    // Then shift to [-π, π]
    if angle > math.Pi {
        angle -= 2 * math.Pi
    }
    return angle
}