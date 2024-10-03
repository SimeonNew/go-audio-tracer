package main

import "github.com/hajimehoshi/oto/v2"

type Vector struct {
	x, y float64
}

type Ray struct {
	origin, direction Vector
}

type WallProperties struct {
	absorption   float64
	transparency float64
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
}

type Game struct {
	walls         []Wall
	audioSource   AudioSource
	listener      Listener
	rays          []Ray
	intersections []Vector
	leftPaths     []AudioPath
	rightPaths    []AudioPath
	audioContext  *oto.Context
	player        oto.Player
	buffer        []byte
	frame         int
	totalSamples  int
}