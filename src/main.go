package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	frameRate = 30
	fontSize  = 1
)

var (
	scaleX     = 1 * fontSize
	scaleY     = 1 * fontSize
	frameDelay = time.Second / time.Duration(frameRate)
)

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func grayscale(c color.Color) int {
	r, g, b, _ := c.RGBA()
	return int(0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8))
}

func avgPixel(img image.Image, x, y, w, h int) int {
	cnt, sum := 0, 0
	max := img.Bounds().Max
	for i := x; i < x+w && i < max.X; i++ {
		for j := y; j < y+h && j < max.Y; j++ {
			sum += grayscale(img.At(i, j))
			cnt++
		}
	}
	return sum / cnt
}

func brightnessToASCII(brightness int) byte {

	asciiChars := []byte(" .:-=+*#%@")

	if brightness < 0 {
		brightness = 0
	} else if brightness > 255 {
		brightness = 255
	}

	index := brightness * (len(asciiChars) - 1) / 255

	if index < 0 {
		index = 0
	} else if index >= len(asciiChars) {
		index = len(asciiChars) - 1
	}
	return asciiChars[index]
}

func processFrame(img image.Image, ch chan<- string) {
	max := img.Bounds().Max
	var frameData string
	for y := 0; y < max.Y; y += scaleY {
		for x := 0; x < max.X; x += scaleX {
			c := avgPixel(img, x, y, scaleX, scaleY)
			// Convert brightness to ASCII character
			frameData += string(brightnessToASCII(c))
		}
		frameData += "\n"
	}
	ch <- frameData
}

func main() {
	picsDir := "src/pics"
	dir, err := os.ReadDir(picsDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var frames []image.Image
	for _, entry := range dir {
		if !entry.IsDir() {
			imgPath := filepath.Join(picsDir, entry.Name())
			img, err := loadImage(imgPath)
			if err != nil {
				continue
			}
			frames = append(frames, img)
		}
	}

	var wg sync.WaitGroup
	ch := make(chan string)

	go func() {
		for _, img := range frames {
			wg.Add(1)
			go func(img image.Image) {
				defer wg.Done()
				processFrame(img, ch)
			}(img)
			time.Sleep(frameDelay) // Introduce delay between frames
		}
		wg.Wait()
		close(ch)
	}()

	for frameData := range ch {
		fmt.Print(frameData)
	}
}
