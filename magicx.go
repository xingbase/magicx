package magicx

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

var fileExtentions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

type FileInfo struct {
	Name string
	Ext  string
	Size int64
}

type ImageInfo struct {
	Path   string
	Size   int64
	Image  image.Image
	Format string
}

func Load(dir string) <-chan FileInfo {
	out := make(chan FileInfo)

	go func() {
		defer close(out)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if fileExtentions[ext] {
					out <- FileInfo{Name: path, Ext: ext, Size: info.Size()}
				}

			}
			return nil
		})
	}()

	return out
}

func Rename(in <-chan FileInfo) <-chan FileInfo {
	out := make(chan FileInfo)

	go func() {
		defer close(out)

		for file := range in {
			parts := strings.Split(file.Name, "_")

			num := strings.TrimSuffix(parts[len(parts)-1], file.Ext)

			if len(num) < 3 {
				newNum := fmt.Sprintf("%03s", num)
				newFile := strings.Replace(file.Name, num+file.Ext, newNum+file.Ext, 1)

				err := os.Rename(file.Name, newFile)
				if err != nil {
					fmt.Printf("Error rename file %s: %v\n", file.Name, err)
					continue
				}
			}

			out <- file
		}
	}()

	return out
}

func Decode(in <-chan FileInfo) <-chan ImageInfo {
	out := make(chan ImageInfo)

	go func() {
		defer close(out)
		for path := range in {
			file, err := os.Open(path.Name)
			if err != nil {
				fmt.Printf("Error opening file %v\n", err)
				return
			}
			defer file.Close()

			img, format, err := image.Decode(file)
			if err != nil {
				fmt.Printf("Error decoding image %v\n", err)
				return
			}

			out <- ImageInfo{Path: path.Name, Size: path.Size, Image: img, Format: format}
		}
	}()

	return out
}

func Resize(in <-chan ImageInfo) <-chan ImageInfo {
	out := make(chan ImageInfo)

	go func() {
		defer close(out)
		for img := range in {
			bounds := img.Image.Bounds()
			width, height := bounds.Dx(), bounds.Dy()
			fmt.Printf("Original - filename: %s  width:%d  height:%d  size:%d\n", filepath.Base(img.Path), width, height, img.Size)

			// TODO: rule
			// 1. check width over
			// 2. check size over
			if width > 1600 {
				newWidth := 1600
				newHeight := int(float64(height) * float64(newWidth) / float64(width))
				dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
				draw.ApproxBiLinear.Scale(dst, dst.Rect, img.Image, img.Image.Bounds(), draw.Over, nil)
				img.Image = dst
				fmt.Printf("Resized  - filename: %s  width:%d  height:%d\n", filepath.Base(img.Path), newWidth, newHeight)

				out <- img
			}

		}
	}()

	return out
}

func Process(in <-chan ImageInfo) {
	var wg sync.WaitGroup
	for img := range in {
		wg.Add(1)
		go func(img ImageInfo) {
			defer wg.Done()
			outFile, err := os.Create(img.Path)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", img.Path, err)
				return
			}
			defer outFile.Close()

			var encodeErr error
			switch img.Format {
			case "jpeg":
				encodeErr = jpeg.Encode(outFile, img.Image, nil)
			case "png":
				encodeErr = png.Encode(outFile, img.Image)
			case "gif":
				encodeErr = gif.Encode(outFile, img.Image, nil)
			default:
				fmt.Printf("Unsupported image format: %s\n", img.Format)
				return
			}

			if encodeErr != nil {
				fmt.Printf("Error encoding image %s: %v\n", img.Path, encodeErr)
			}
		}(img)
	}
	wg.Wait()
}
