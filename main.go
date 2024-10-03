package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/oto/v2"
)

const (
	screenWidth  = 800
	screenHeight = 600
	numRays      = 360
	maxBounces   = 3
	speedOfSound = 343.0 // m/s
	sineFreq     = 200   // Frequency of sine wave in Hz
	sampleRate   = 44100 // Sample rate for audio
)

func (g *Game) Update() error {
	g.frame++
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.audioSource.position = Vector{float64(x), float64(y)}
	}

	g.rays = make([]Ray, numRays)
	g.intersections = make([]Vector, 0)
	g.leftPaths = make([]AudioPath, 0)
	g.rightPaths = make([]AudioPath, 0)
	initialIntensity := 1.0

	for i := 0; i < numRays; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numRays)
		direction := Vector{math.Cos(angle), math.Sin(angle)}
		ray := Ray{g.audioSource.position, direction}
		g.traceRay(ray, initialIntensity, maxBounces)
	}

	// Generate audio data and append it to the ring buffer
	g.generateAudio()

	return nil
}

// Generate audio data based on the current game state. This function fills the ring buffer
// with the appropriate audio samples for continuous playback.
func (g *Game) generateAudio() {
    for i := 0; i < len(g.buffer); i += 4 {
        // Generate audio samples as before
        sample := int16(math.Sin(2 * math.Pi * sineFreq * float64(g.totalSamples) / float64(sampleRate)) * 32767)
        g.buffer[i] = byte(sample & 0xFF)
        g.buffer[i+1] = byte((sample >> 8) & 0xFF)
        g.buffer[i+2] = byte(sample & 0xFF)
        g.buffer[i+3] = byte((sample >> 8) & 0xFF)
        g.totalSamples++
    }
}



// Draw implements ebiten.Game's Draw function. It draws the game's walls, the
// audio source and listener, and the intersection points of the rays with the
// walls.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	for _, wall := range g.walls {
		vector.StrokeLine(screen, float32(wall.start.x), float32(wall.start.y), float32(wall.end.x), float32(wall.end.y), 1, color.RGBA{255, 255, 255, 255}, true)
	}

	vector.DrawFilledCircle(screen, float32(g.audioSource.position.x), float32(g.audioSource.position.y), 5, color.RGBA{255, 0, 0, 255}, true)

	vector.DrawFilledCircle(screen, float32(g.listener.position.x), float32(g.listener.position.y), 5, color.RGBA{0, 255, 0, 255}, true)

	for _, intersection := range g.intersections {
		vector.DrawFilledCircle(screen, float32(intersection.x), float32(intersection.y), 2, color.RGBA{0, 0, 255, 255}, true)
	}
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
			{Vector{100, 100}, Vector{700, 100}, WallProperties{absorption: 0.2, transparency: 0.1}},
			{Vector{700, 100}, Vector{700, 500}, WallProperties{absorption: 0.2, transparency: 0.1}},
			{Vector{700, 500}, Vector{100, 500}, WallProperties{absorption: 0.2, transparency: 0.1}},
			{Vector{100, 500}, Vector{100, 100}, WallProperties{absorption: 0.2, transparency: 0.1}},
		},
		audioSource:  AudioSource{Vector{400, 300}, sineFreq, 0.5},
		listener:     Listener{Vector{200, 300}, Vector{195, 300}, Vector{205, 300}},
		audioContext: otoCtx,
		buffer:       make([]byte, 88200),
		totalSamples: 0,
	}

	game.player = otoCtx.NewPlayer(game)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("2D Audio Ray Tracing")

	game.player.Play()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
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


