package main

import (
	"math"
	"testing"
)

func TestDistanceFromPointToLine(t *testing.T) {
	tests := []struct {
		name     string
		ray      Ray
		point    Vector
		wantDist float64
		wantDistToSource float64
	}{
		{
			name: "point on the line",
			ray:  Ray{Vector{0, 0}, Vector{1, 0}},
			point:    Vector{2, 0},
			wantDist: 0,
			wantDistToSource: 2,
		},
		{
			name: "point above the line",
			ray:  Ray{Vector{0, 0}, Vector{1, 0}},
			point:    Vector{2, 1},
			wantDist: 1,
			wantDistToSource: 2,
		},
		{
			name: "point below the line",
			ray:  Ray{Vector{0, 0}, Vector{1, 0}},
			point:    Vector{2, -1},
			wantDist: 1,
			wantDistToSource: 2,
		},
		{
			name: "point away from the line",
			ray:  Ray{Vector{50, 300}, Vector{-1, 1.2246467991473515e-16}},
			point:    Vector{400, 300},
			wantDist: 0,
			wantDistToSource: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist, distToSource := distanceFromPointToLine(tt.ray, tt.point)
			if math.Abs(dist-tt.wantDist) > 1e-6 {
				t.Errorf("distanceFromPointToLine() got dist = %v, want %v", dist, tt.wantDist)
			}
			if math.Abs(distToSource-tt.wantDistToSource) > 1e-6 {
				t.Errorf("distanceFromPointToLine() got distToSource = %v, want %v", distToSource, tt.wantDistToSource)
			}
		})
	}
}