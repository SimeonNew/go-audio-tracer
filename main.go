package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/oto/v2"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
	numRays      = 360
	maxBounces   = 2
	speedOfSound = 343.0 // m/s
	sineFreq     = 200   // Frequency of sine wave in Hz
	sampleRate   = 44100 // Sample rate for audio
	proximityThreshold = 5.0
	volume = 1000
)

func (g *Game) Update() error {
    g.frame++

    // Get current mouse position
    x, y := ebiten.CursorPosition()
    mousePosition := Vector{float64(x), float64(y)}

    // Check mouse button state
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        // If the mouse button was just pressed, set the listener position
        g.listener.position = mousePosition
        g.isDragging = true // Start dragging
    }

    if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
        // If the mouse button was released, stop dragging
        g.isDragging = false
    }

    if g.isDragging {
        // Update the listener's position to follow the mouse while dragging
        g.listener.position = mousePosition
    }

    g.rays = make([]Ray, numRays)
    g.leftPaths = make([]AudioPath, 0)
    g.rightPaths = make([]AudioPath, 0)
    g.rayPathPoints = make([][]RayPathPoint, numRays)
    initialIntensity := 1.0

    for i := 0; i < numRays; i++ {
        angle := float64(i) * 2 * math.Pi / float64(numRays)
        direction := Vector{math.Cos(angle), math.Sin(angle)}
        g.rays[i] = Ray{g.listener.position, direction}
        g.rayPathPoints[i] = []RayPathPoint{{g.listener.position, initialIntensity}}
        g.traceRay(g.rays[i], initialIntensity, maxBounces, i)
    }

    // Generate audio data and append it to the ring buffer
    g.generateAudio()

    return nil
}




func calculateILD(direction Vector, isLeft bool) float64 {
	const minAttenuation = 0.3

	angle := math.Atan2(direction.y, direction.x) - math.Pi/2
	shadowEffect := 0.5 * (1.0 + math.Abs(math.Sin(angle)))
	if isLeft {
		shadowEffect = 1.0 - shadowEffect
	}
	return minAttenuation + (1.0-minAttenuation)*shadowEffect
}

func (g *Game) generateAudio() {
	var wg sync.WaitGroup

	leftChannel := make(chan float64, len(g.buffer)/4)
	rightChannel := make(chan float64, len(g.buffer)/4)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i, path := range g.leftPaths {
			g.leftPaths[i].ild = calculateILD(path.direction, true)
		}

		for i := 0; i < len(g.buffer); i += 4 {
			var sampleLeft float64 = 0
			currentTime := float64(g.totalSamples+int(i/4)) / float64(sampleRate)

			for _, path := range g.leftPaths {
				adjustedAmplitude := path.amplitude * path.ild
				baseSignal := 2 * math.Pi * path.source.frequency * (currentTime - path.delay)
				sampleLeft += adjustedAmplitude * volume * math.Sin(baseSignal)
			}

			leftChannel <- sampleLeft
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i, path := range g.rightPaths {
			g.rightPaths[i].ild = calculateILD(path.direction, false)
		}

		for i := 0; i < len(g.buffer); i += 4 {
			var sampleRight float64 = 0
			currentTime := float64(g.totalSamples+int(i/4)) / float64(sampleRate)

			for _, path := range g.rightPaths {
				adjustedAmplitude := path.amplitude * path.ild
				baseSignal := 2 * math.Pi * path.source.frequency * (currentTime - path.delay)
				sampleRight += adjustedAmplitude * volume * math.Sin(baseSignal)
			}

			rightChannel <- sampleRight
		}
	}()

	wg.Wait()
	close(leftChannel)
	close(rightChannel)

	for i := 0; i < len(g.buffer); i += 4 {
		sampleLeft := int16(<-leftChannel)
		sampleRight := int16(<-rightChannel)
		g.buffer[i] = byte(sampleLeft & 0xFF)
		g.buffer[i+1] = byte((sampleLeft >> 8) & 0xFF)
		g.buffer[i+2] = byte(sampleRight & 0xFF)
		g.buffer[i+3] = byte((sampleRight >> 8) & 0xFF)
		g.totalSamples++
	}
}



// Draw implements ebiten.Game's Draw function. It draws the game's walls, the
// audio source and listener, and the intersection points of the rays with the
// walls.
func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.Black)


    for _, path := range g.rayPathPoints {
        for i := 0; i < len(path)-1; i++ {
            // Get start intensity
            startIntensity := path[i].intensity

            // Get start and end positions
            startX := float32(path[i].position.x)
            startY := float32(path[i].position.y)
            endX := float32(path[i+1].position.x)
            endY := float32(path[i+1].position.y)

            // Calculate the total segment length
            totalLength := float32(math.Sqrt(float64(
                (endX-startX)*(endX-startX) + 
                (endY-startY)*(endY-startY))))

            // Draw multiple small segments to create gradient
            numSegments := 50 // Increase for smoother gradient
            for j := 0; j < numSegments; j++ {
                progress := float64(j) / float64(numSegments)
                
                // Interpolate positions
                x1 := startX + float32(progress)*(endX-startX)
                y1 := startY + float32(progress)*(endY-startY)
                x2 := startX + float32(progress + 1.0/float64(numSegments))*(endX-startX)
                y2 := startY + float32(progress + 1.0/float64(numSegments))*(endY-startY)

                // Calculate distance from start for this segment
                distanceFromStart := totalLength * float32(progress)
                
                // Calculate intensity using inverse square law
                // Add a small offset to prevent division by zero
                segmentIntensity := startIntensity / (1 + float64(distanceFromStart*distanceFromStart)/10000)

                // Calculate color for this segment
                r := uint8(150 * segmentIntensity)
                g := uint8(0)
                b := uint8(10 - 10*segmentIntensity)
                a := uint8(20*segmentIntensity)

                rayColor := color.RGBA{r, g, b, a}

                // Draw the small segment
                vector.StrokeLine(screen, x1, y1, x2, y2, 5, rayColor, true)
            }
        }
    }
	// Draw walls
    for _, wall := range g.walls {
        vector.StrokeLine(screen, float32(wall.start.x), float32(wall.start.y), float32(wall.end.x), float32(wall.end.y), 1, color.RGBA{255, 255, 255, 255}, true)
    }
    // Draw audio source
    vector.DrawFilledCircle(screen, float32(g.audioSource.position.x), float32(g.audioSource.position.y), 5, color.RGBA{255, 255, 255, 255}, true)

    // Draw listener
    vector.DrawFilledCircle(screen, float32(g.listener.position.x), float32(g.listener.position.y), 5, color.RGBA{0, 0, 255, 100}, true)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// main initializes the game and starts the game loop. It creates an oto audio
// context, sets up the game state (walls, audio source, listener, and audio
// buffer), and creates an oto player for the game. It then sets up the Ebiten
// window and starts the game loop with ebiten.RunGame. If there's an error, it
// logs the error and exits.
func main() {
	otoCtx, readyChan, err := oto.NewContext(sampleRate, 2, 2)
	if err != nil {
		log.Fatal(err)
	}
	<-readyChan

	game := &Game{
		walls: []Wall{
			{Vector{240, 180}, Vector{1680, 180}, WallProperties{absorption: 0.2, transparency: 0.2, transmissionRoughness: 0.5, roughness: 0.5}},
			{Vector{1680, 180}, Vector{1680, 900}, WallProperties{absorption: 0.2, transparency: 0.2, transmissionRoughness: 0.5, roughness: 0.5}},
			{Vector{1680, 900}, Vector{240, 900}, WallProperties{absorption: 0.2, transparency: 0.2, transmissionRoughness: 0.5, roughness: 0.5}},
			{Vector{240, 900}, Vector{240, 180}, WallProperties{absorption: 0.2, transparency: 0.2, transmissionRoughness: 0.5, roughness: 0.5}},
            {Vector{400, 750}, Vector{400, 320}, WallProperties{absorption: 0.2, transparency: 0.5, transmissionRoughness: 0.5, roughness: 0.5}},

		},
		audioSource:  AudioSource{Vector{600, 535}, sineFreq, 0.5},
		listener:     Listener{Vector{800, 535}, Vector{795, 535}, Vector{805, 535}},
		audioContext: otoCtx,
		buffer:       make([]byte, 176400),
		totalSamples: 0,
	}
    game.getWallEdges()
    fmt.Println(game.wallEdges)

	game.player = otoCtx.NewPlayer(game)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("2D Audio Ray Tracing")
	ebiten.SetTPS(2)

	game.player.Play()

	if err := ebiten.RunGame(game); err != nil {    
		log.Fatal(err)
	}
}

// Returns all edges/corners for diffraction calculations
func (g *Game) getWallEdges() {
    edges := make([]WallEdge, 0)
    pointMap := make(map[Vector]bool)

    // Iterate through each wall
    for i, wall1 := range g.walls {
        // Calculate the wall's direction and normal
        wallDir := Vector{
            wall1.end.x - wall1.start.x,
            wall1.end.y - wall1.start.y,
        }.normalize()
        wallNormal := Vector{-wallDir.y, wallDir.x}

        // Add wall endpoints
        for _, point := range []Vector{wall1.start, wall1.end} {
            if !pointMap[point] {
                edges = append(edges, WallEdge{
                    position: point,
                    normal1:  wallNormal,
                    isCorner: false,
                })
                pointMap[point] = true
            }
        }

        // Check if the end of wall1 is a corner shared with another wall
        for j, wall2 := range g.walls {
            if i == j {
                continue // Skip if it's the same wall
            }
            wall2Dir := Vector{
                wall2.end.x - wall2.start.x,
                wall2.end.y - wall2.start.y,
            }.normalize()
            wall2Normal := Vector{-wall2Dir.y, wall2Dir.x}

            if wall1.start.x == wall2.end.x && wall1.start.y == wall2.end.y {
                // Mark the point as a corner
                for k, edge := range edges {
                    if edge.position == wall1.start {
                        edges[k].isCorner = true
                        edges[k].normal2 = wall2Normal
                        break
                    }
                }
            }
            if wall1.end.x == wall2.start.x && wall1.end.y == wall2.start.y {
                // Mark the point as a corner
                for k, edge := range edges {
                    if edge.position == wall1.end {
                        edges[k].isCorner = true
                        edges[k].normal2 = wall2Normal
                        break
                    }
                }
            }
        }
    }

    g.wallEdges = edges
}
// Read implements the io.Reader interface for the Game type. It fills the given buffer
// with audio samples from the ring buffer. If the ring buffer is empty, it will block
// until more audio data is generated.
func (g *Game) Read(buf []byte) (int, error) {
    // Ensure there are enough samples available in the buffer
    if len(g.buffer) < len(buf) {
        copy(buf, g.buffer)
    }
    
    // Copy generated audio data into the output buffer
    copy(buf, g.buffer[:len(buf)])

    return len(buf), nil
}


