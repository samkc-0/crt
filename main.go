package main

import (
  "flag"
  "fmt"
  "image"
  _ "image/gif"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "golang.org/x/term"
  "os/user"
  "strings"
)

func expandPath(path string) string {
  if strings.HasPrefix(path, "~") {
    usr, _ := user.Current()
    return strings.Replace(path, "~", usr.HomeDir, 1)
  }
  return path
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
    fmt.Println("Usage: go run main.go -image=/path/to/image.jpg")
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
  cols, rows, err := term.GetSize(0)
  if err != nil {
    fmt.Println("could not get terminal size.")
    os.Exit(1)
  }

  fmt.Printf("📏 terminal size: %d columns x %d rows\n", cols, rows)
  bounds := img.Bounds()
  imgW := bounds.Dx()
  imgH := bounds.Dy()

  // use terminal width as number of columns
  aspectRatio := 0.5 // tweak based on terminal 
  rowsAdjusted := int(float64(rows) * aspectRatio)

  scaleX := float64(imgW) / float64(cols)
  scaleY := float64(imgH) / float64(rowsAdjusted)

  for y := 0; y < rowsAdjusted; y++ {
    for x := 0; x < cols; x++ {
      imgX := int(float64(x) * scaleX)
      imgY := int(float64(y) * scaleY)


      if imgX >= imgW || imgY >= imgH {
        continue
      }

      r, g, b, _ := img.At(bounds.Min.X+imgX, bounds.Min.Y+imgY).RGBA()
      r8 := r>>8
      g8 := g>>8
      b8 := b>>8

      fmt.Printf("\x1b[48;2;%d;%d;%dm ", r8, g8, b8)
    }
    fmt.Print("\x1b[0m\n")
  }
}
