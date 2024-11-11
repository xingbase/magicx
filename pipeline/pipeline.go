package pipeline

import (
	"bytes"
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

var ContentTypeByLimitInfo = map[string]LimitSizeInfo{
	"comic":          {Width: 1600, Size: 20480},
	"magazine_comic": {Width: 2266, Size: 30720},
}

type LimitSizeInfo struct {
	Width int
	Size  int64
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

func Rename(in <-chan FileInfo, n int) <-chan FileInfo {
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

				file.Name = newFile
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
			}

			img, format, err := image.Decode(file)
			file.Close()
			if err != nil {
				fmt.Printf("Error decoding image %v\n", err)
				continue
			}

			out <- ImageInfo{Path: path.Name, Size: path.Size, Image: img, Format: format}
		}
	}()

	return out
}

func Resize(in <-chan ImageInfo, limitWidth int, limitKb int64, initialPercent float64) <-chan ImageInfo {
	out := make(chan ImageInfo)

	go func() {
		defer close(out)
		for img := range in {
			bounds := img.Image.Bounds()
			width, height := bounds.Dx(), bounds.Dy()
			fmt.Printf("Original - filename: %s  width:%d  height:%d  size:%d\n", filepath.Base(img.Path), width, height, img.Size)

			if width > limitWidth || img.Size > limitKb*1024 {
				percent := initialPercent
				for percent >= 50 { // Don't go below 50% of original size
					newWidth := int(float64(width) * percent / 100)
					newHeight := int(float64(height) * float64(newWidth) / float64(width))

					dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
					draw.ApproxBiLinear.Scale(dst, dst.Rect, img.Image, img.Image.Bounds(), draw.Over, nil)

					// Encode to JPEG to check file size
					var buf bytes.Buffer
					jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85})

					if buf.Len() <= int(limitKb*1024) {
						img.Image = dst
						fmt.Printf("Resized  - filename: %s  width:%d  height:%d  size:%d\n", filepath.Base(img.Path), newWidth, newHeight, buf.Len())
						break
					}

					percent -= 5 // Reduce by 5% and try again
				}

				out <- img
			}
		}
	}()

	return out
}

func Save(in <-chan ImageInfo) {
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
