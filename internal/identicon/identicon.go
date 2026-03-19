package identicon

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/color"
	"image/png"
)

const (
	gridSize  = 5
	cellSize  = 50
	padding   = 3
	imageSize = gridSize*cellSize + 2*padding
)

// Generate creates a 256x256 identicon PNG from the given input string (typically a UUID).
// The output is deterministic: the same input always produces the same image.
func Generate(input string) ([]byte, error) {
	hash := sha256.Sum256([]byte(input))

	fg := color.NRGBA{R: hash[0], G: hash[1], B: hash[2], A: 255}
	bg := color.NRGBA{R: 240, G: 240, B: 240, A: 255}

	// Build a 5x5 grid with horizontal symmetry.
	// Only compute the left half + center column; mirror to the right.
	var grid [gridSize][gridSize]bool
	idx := 3
	for y := 0; y < gridSize; y++ {
		for x := 0; x <= gridSize/2; x++ {
			filled := hash[idx%len(hash)]%2 == 0
			grid[y][x] = filled
			grid[y][gridSize-1-x] = filled
			idx++
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, imageSize, imageSize))

	// Fill background.
	for y := 0; y < imageSize; y++ {
		for x := 0; x < imageSize; x++ {
			img.Set(x, y, bg)
		}
	}

	// Draw grid cells.
	for gy := 0; gy < gridSize; gy++ {
		for gx := 0; gx < gridSize; gx++ {
			if !grid[gy][gx] {
				continue
			}
			x0 := padding + gx*cellSize
			y0 := padding + gy*cellSize
			for dy := 0; dy < cellSize; dy++ {
				for dx := 0; dx < cellSize; dx++ {
					img.Set(x0+dx, y0+dy, fg)
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
