package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"time"
)

type Scene struct {
	displayWidth  int
	displayHeight int
	pixels        []rune
	depth         []float64
	bgChar        rune
}

type Transform struct {
	rotateX, rotateY, rotateZ float64
}

type Cube struct {
	size     float64
	position float64
	chars    []rune
}

const (
	DISPLAY_WIDTH  = 160
	DISPLAY_HEIGHT = 44
	CAMERA_DIST    = 100
	PERSPECTIVE    = 40
	STEP_SIZE      = 0.6
)

func NewScene() *Scene {
	s := &Scene{
		displayWidth:  DISPLAY_WIDTH,
		displayHeight: DISPLAY_HEIGHT,
		pixels:        make([]rune, DISPLAY_WIDTH*DISPLAY_HEIGHT),
		depth:         make([]float64, DISPLAY_WIDTH*DISPLAY_HEIGHT),
		bgChar:        ' ',
	}
	return s
}

func (s *Scene) clear() {
	for i := range s.pixels {
		s.pixels[i] = s.bgChar
		s.depth[i] = 0
	}
}

func (s *Scene) draw() {
	fmt.Print("\x1b[H")
	for i := 0; i < len(s.pixels); i++ {
		if i%s.displayWidth == 0 && i > 0 {
			fmt.Print("\n")
		}
		fmt.Printf("%c", s.pixels[i])
	}
}

func project3Dto2D(x, y, z float64, t *Transform) (float64, float64, float64) {
	// Rotation matrices combined
	newX := y*math.Sin(t.rotateX)*math.Sin(t.rotateY)*math.Cos(t.rotateZ) -
		z*math.Cos(t.rotateX)*math.Sin(t.rotateY)*math.Cos(t.rotateZ) +
		y*math.Cos(t.rotateX)*math.Sin(t.rotateZ) +
		z*math.Sin(t.rotateX)*math.Sin(t.rotateZ) +
		x*math.Cos(t.rotateY)*math.Cos(t.rotateZ)

	newY := y*math.Cos(t.rotateX)*math.Cos(t.rotateZ) +
		z*math.Sin(t.rotateX)*math.Cos(t.rotateZ) -
		y*math.Sin(t.rotateX)*math.Sin(t.rotateY)*math.Sin(t.rotateZ) +
		z*math.Cos(t.rotateX)*math.Sin(t.rotateY)*math.Sin(t.rotateZ) -
		x*math.Cos(t.rotateY)*math.Sin(t.rotateZ)

	newZ := z*math.Cos(t.rotateX)*math.Cos(t.rotateY) -
		y*math.Sin(t.rotateX)*math.Cos(t.rotateY) +
		x*math.Sin(t.rotateY)

	return newX, newY, newZ + CAMERA_DIST
}

func (s *Scene) plotPoint(x, y, z float64, ch rune, offset float64) {
	invZ := 1.0 / z
	screenX := int(float64(s.displayWidth)/2.0 + offset + PERSPECTIVE*invZ*x*2.0)
	screenY := int(float64(s.displayHeight)/2.0 + PERSPECTIVE*invZ*y)

	idx := screenX + screenY*s.displayWidth
	if idx >= 0 && idx < len(s.pixels) {
		if invZ > s.depth[idx] {
			s.depth[idx] = invZ
			s.pixels[idx] = ch
		}
	}
}

func (s *Scene) renderCube(cube Cube, t *Transform) {
	size := cube.size
	chars := cube.chars

	for x := -size; x < size; x += STEP_SIZE {
		for y := -size; y < size; y += STEP_SIZE {
			// Draw each face
			surfaces := [][3]float64{
				{x, y, -size},  // Front
				{size, y, x},   // Right
				{-size, y, -x}, // Left
				{-x, y, size},  // Back
				{x, -size, -y}, // Bottom
				{x, size, y},   // To
			}

			for i, surf := range surfaces {
				px, py, pz := project3Dto2D(surf[0], surf[1], surf[2], t)
				s.plotPoint(px, py, pz, chars[i], cube.position)
			}
		}
	}
}

func main() {
	fmt.Print("\x1b[2J\x1b[?25l")
	defer fmt.Print("\x1b[?25h")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Print("\x1b[?25h")
		os.Exit(0)
	}()

	scene := NewScene()
	transform := &Transform{}

	cubes := []Cube{
		{size: 20, position: -40, chars: []rune{'@', '$', '~', '#', ';', '+'}},
		{size: 10, position: 10, chars: []rune{'O', '=', '*', '%', '&', 'X'}},
		{size: 5, position: 40, chars: []rune{':', '.', ',', '|', '-', '+'}},
	}

	for {
		scene.clear()

		for _, cube := range cubes {
			scene.renderCube(cube, transform)
		}

		scene.draw()

		transform.rotateX += 0.05
		transform.rotateY += 0.05
		transform.rotateZ += 0.01

		time.Sleep(16 * time.Millisecond)
	}
}
