package main

import (
	"github.com/hajimehoshi/oto/v2"
)

type Vector struct {
	x, y float64
}



type Ray struct {
	origin, direction Vector
}

type WallProperties struct {
	absorption   float64
	transparency float64
	transmissionRoughness float64
	roughness   float64
}

type Wall struct {
	start, end Vector
	properties WallProperties
}

type AudioSource struct {
	position  Vector
	frequency float64
	amplitude float64
}

type Listener struct {
	position Vector
	leftEar  Vector
	rightEar Vector
}

type AudioPath struct {
	source    AudioSource
	delay     float64
	amplitude float64
	direction Vector
	ild       float64
}

type Game struct {
	walls         []Wall
	wallEdges     []WallEdge
    audioSource   AudioSource
    listener      Listener
    rays          []Ray
    leftPaths     []AudioPath
    rightPaths    []AudioPath
    audioContext  *oto.Context
    player        oto.Player
    buffer        []byte
    totalSamples  int
    frame         int
    rayPathPoints      [][]RayPathPoint 
	isDragging    bool
}

type RayPathPoint struct {
    position  Vector
    intensity float64
}

type WallEdge struct {
    position Vector
    normal1  Vector  // Normal of first wall
    normal2  Vector  // Normal of second wall (if corner)
    isCorner bool
}