package main

import (
	"flag"
	"fmt"
	"golang.org/x/term"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/user"
	"strings"
)

type Number interface {
	~int | ~int64 | float64
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, _ := user.Current()
		return strings.Replace(path, "~", usr.HomeDir, 1)
	}
	return path
}

func getAspectRatio[T Number](w, h T) float64 {
	return float64(w) / float64(h)
}

func drawAscii(r, g, b int) string {
	chars := "@%#*+=-:."
	luminance := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	index := int((luminance / 255.0) * float64(len(chars)-1))
	return string(chars[index])
}

func drawRGB(r, g, b int) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm \x1b[0m", r, g, b)
}

func main() {
	// define a flag to get the image path
	imagePath := flag.String("image", "", "path to the image file")
	imagePathShort := flag.String("i", "", "alias for -image")
	flag.Parse()

	path := *imagePath
	if path == "" {
		path = *imagePathShort
	}

	// if no image path was provided, show usage and exit
	if path == "" {
		fmt.Println("usage: go run main.go -image=/path/to/image.jpg")
		os.Exit(1)
	}

	path = expandPath(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("file does not exist: %s\n", path)
		os.Exit(1)
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// try to decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("the file isn't a valid image: %v\n", err)
		os.Exit(1)
	}

	// fmt.Printf("✅ image loaded successfully: %s\n", path)
	// fmt.Printf("🖼️ image size: %v\n", img.Bounds())
	termW, termH, err := term.GetSize(0)
	if err != nil {
		fmt.Println("could not get terminal size.")
		os.Exit(1)
	}
	termH -= 1
	// fmt.Printf("📏 terminal size: %d columns x %d termH\n", termW, termH)
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	tAspect := getAspectRatio(termH, termW)
	pAspect := getAspectRatio(imgW, imgH)
	imgWAdjusted := termW
	imgHAdjusted := int(float64(imgWAdjusted) / pAspect)
	if tAspect <= pAspect {
		imgHAdjusted = termH
		imgWAdjusted = 2 * int(float64(imgHAdjusted)*pAspect)
	}
	scaleX := float64(imgW) / float64(imgWAdjusted)
	scaleY := float64(imgH) / float64(imgHAdjusted)

	for y := 0; y < imgHAdjusted; y++ {
		for x := 0; x < imgWAdjusted; x++ {
			imgX := int(float64(x) * scaleX)
			imgY := int(float64(y) * scaleY)
			imgYb := imgY + int(scaleY)/2
			if imgX >= imgW || imgY >= imgH {
				continue
			}

			r, g, b, _ := img.At(bounds.Min.X+imgX, bounds.Min.Y+imgY).RGBA()
			r8 := r >> 8
			g8 := g >> 8
			b8 := b >> 8

			if imgYb >= imgH {
				continue
			}

			bg_r, bg_g, bg_b, _ := img.At(bounds.Min.X+imgX, bounds.Min.Y+imgYb).RGBA()
			bg_r8 := bg_r >> 8
			bg_g8 := bg_g >> 8
			bg_b8 := bg_b >> 8

			char := '▀'
			fmt.Printf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%c", r8, g8, b8, bg_r8, bg_g8, bg_b8, char)
		}
		fmt.Print("\x1b[0m\n")
	}
}
