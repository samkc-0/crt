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

	fmt.Printf("✅ image loaded successfully: %s\n", path)
	fmt.Printf("🖼️ image size: %v\n", img.Bounds())
	termW, termH, err := term.GetSize(0)
	termH -= 5
	if err != nil {
		fmt.Println("could not get terminal size.")
		os.Exit(1)
	}

	fmt.Printf("📏 terminal size: %d columns x %d termH\n", termW, termH)
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	// use terminal width as number of columns
	// aspectRatio := 0.5 // tweak based on terminal
	// aspect = h / w
	// if the aspect > 1 then it's portrait.
	// if the aspect 0 <= x <= 1 then it is landscape
	// so the aspect is basically "how potrait is this image"
	// so when the image is more portrait than the terminal, pad the sides.
	// when it's less portrait than the terminal, padd the top and bottom.
	tAspect := getAspectRatio(termH, termW)
	pAspect := getAspectRatio(imgW, imgH)
	// pad left and right
	imgWAdjusted := termW
	imgHAdjusted := int(float64(imgWAdjusted) / pAspect)
	// if the image is more potrait than the terminal, stretch/shrink img to height
	if tAspect <= pAspect {
		imgHAdjusted = termH
		imgWAdjusted = int(float64(imgHAdjusted) * pAspect)
	}
	imgWAdjusted *= 2
	actualAspect := float64(imgWAdjusted) / float64(imgHAdjusted)
	scaleX := float64(imgW) / float64(imgWAdjusted)
	scaleY := float64(imgH) / float64(imgHAdjusted)
	fmt.Printf("📐 resizing %dx%d (aspect %f) image to %dx%d (aspect %f)...", imgW, imgH, pAspect, imgHAdjusted, imgWAdjusted, actualAspect)
	for y := 0; y < imgHAdjusted; y++ {
		for x := 0; x < imgWAdjusted; x++ {
			imgX := int(float64(x) * scaleX)
			imgY := int(float64(y) * scaleY)

			if imgX >= imgW || imgY >= imgH {
				continue
			}

			r, g, b, _ := img.At(bounds.Min.X+imgX, bounds.Min.Y+imgY).RGBA()
			r8 := r >> 8
			g8 := g >> 8
			b8 := b >> 8

			fmt.Printf("\x1b[48;2;%d;%d;%dm ", r8, g8, b8)
		}
		fmt.Print("\x1b[0m\n")
	}
}
