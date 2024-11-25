package pipeline

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

var fileExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

var ContentTypeByLimitInfo = map[string]LimitSizeInfo{
	"comic":          {Width: 1600, Size: 10240},
	"magazine_comic": {Width: 2266, Size: 30720},
}

type LimitSizeInfo struct {
	Width int
	Size  int64
}

type FileInfo struct {
	Full        string
	Path        string
	Name        string
	Ext         string
	Size        int64
	IsMissmatch bool
}

type ImageInfo struct {
	Full        string
	Path        string
	Name        string
	Size        int64
	Image       image.Image
	Format      string
	IsStandard  bool
	IsMissmatch bool
}

func Load(dir string) <-chan map[string][]FileInfo {
	out := make(chan map[string][]FileInfo)

	go func() {
		defer close(out)

		files := make(map[string][]FileInfo)

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if fileExtensions[ext] {
					relPath, _ := filepath.Rel(dir, filepath.Dir(path))

					parts := strings.Split(relPath, "/")
					lastFolder := parts[len(parts)-1]

					folderN, _ := extractChapterFromFolderName(lastFolder)
					fileN, _ := extractChapterFromFileName(info.Name())

					log.Println("relPath: ", relPath, ", FileName: ", info.Name(), ", FolderNum: ", folderN, ", FileNum: ", fileN)

					fileInfo := FileInfo{
						Full:        path,
						Path:        lastFolder,
						Name:        info.Name(),
						Ext:         ext,
						Size:        info.Size(),
						IsMissmatch: folderN != fileN,
					}
					files[relPath] = append(files[relPath], fileInfo)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Println("Error walking through directory:", err)
		}

		out <- files
	}()

	return out
}

func Rename(in <-chan map[string][]FileInfo, n int) <-chan map[string][]FileInfo {
	out := make(chan map[string][]FileInfo)

	go func() {
		defer close(out)

		for fileMap := range in {
			newFileMap := make(map[string][]FileInfo)

			for dir, files := range fileMap {
				newFiles := make([]FileInfo, 0, len(files))

				for _, file := range files {
					parts := strings.Split(file.Full, "_")
					if len(parts) > 0 {
						num := strings.TrimSuffix(parts[len(parts)-1], file.Ext)

						if len(num) < n {
							newNum := fmt.Sprintf("%0*s", n, num)
							newName := strings.TrimSuffix(file.Name, file.Ext) + newNum + file.Ext
							newFile := strings.Replace(file.Full, num+file.Ext, newNum+file.Ext, 1)

							err := os.Rename(file.Full, newFile)
							if err != nil {
								fmt.Printf("Error rename file %s: %v\n", file.Name, err)
								newFiles = append(newFiles, file) // Keep original file info if rename fails
								continue
							}

							file.Name = newName
							file.Full = newFile
						}
					}
					newFiles = append(newFiles, file)
				}

				newFileMap[dir] = newFiles
			}

			out <- newFileMap
		}
	}()

	return out
}

func Decode(in <-chan map[string][]FileInfo) <-chan map[string][]ImageInfo {
	out := make(chan map[string][]ImageInfo)

	go func() {
		defer close(out)
		for dirMap := range in {
			resultMap := make(map[string][]ImageInfo)

			for dir, files := range dirMap {
				var imageInfos []ImageInfo

				for _, fileInfo := range files {
					// skip thumbnail file
					if strings.Contains(fileInfo.Name, "tmb_") {
						continue
					}

					file, err := os.Open(fileInfo.Full)
					if err != nil {
						fmt.Printf("Error opening file %s: %v\n", fileInfo.Full, err)
						continue
					}

					img, format, err := image.Decode(file)
					file.Close()
					if err != nil {
						fmt.Printf("Error decoding image %s: %v\n", fileInfo.Full, err)
						continue
					}

					imageInfos = append(imageInfos, ImageInfo{
						Full:        fileInfo.Full,
						Path:        fileInfo.Path,
						Name:        fileInfo.Name,
						Size:        fileInfo.Size,
						Image:       img,
						Format:      format,
						IsMissmatch: fileInfo.IsMissmatch,
					})
				}

				if len(imageInfos) > 0 {
					resultMap[dir] = imageInfos
				}
			}

			if len(resultMap) > 0 {
				out <- resultMap
			}
		}
	}()

	return out
}

func CheckImageSize(in <-chan map[string][]ImageInfo, info LimitSizeInfo) <-chan ImageInfo {
	out := make(chan ImageInfo)

	go func() {
		defer close(out)

		for dirs := range in {
			for _, images := range dirs {
				groupedImages := make(map[int][]ImageInfo)
				widthCounts := make(map[int]int)
				maxCount := 0
				standardWidth := 0

				// First pass: group images and count widths
				for _, img := range images {
					width := img.Image.Bounds().Dx()
					groupedImages[width] = append(groupedImages[width], img)
					widthCounts[width]++

					if widthCounts[width] > maxCount {
						maxCount = widthCounts[width]
						standardWidth = width
					}
				}

				// Second pass: process and send images
				for width, imgs := range groupedImages {
					isStandard := (width == standardWidth)
					for _, img := range imgs {
						processedImg := img
						processedImg.IsStandard = isStandard

						// size check
						if img.Size > info.Size*1024 {
							processedImg.IsStandard = true
						}

						out <- processedImg
					}
				}
			}
		}
	}()

	return out
}

func Resize(in <-chan ImageInfo, info LimitSizeInfo, initialPercent float64) <-chan ImageInfo {
	out := make(chan ImageInfo)

	go func() {
		defer close(out)
		for img := range in {
			bounds := img.Image.Bounds()
			width, height := bounds.Dx(), bounds.Dy()
			// fmt.Printf("Original - filename: %s  width:%d  height:%d  size:%d\n", filepath.Base(img.Path), width, height, img.Size)

			if width > info.Width || img.Size > info.Size*1024 {
				percent := initialPercent
				for percent >= 50 { // Don't go below 50% of original size
					newWidth := int(float64(width) * percent / 100)
					newHeight := int(float64(height) * float64(newWidth) / float64(width))

					dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
					draw.ApproxBiLinear.Scale(dst, dst.Rect, img.Image, img.Image.Bounds(), draw.Over, nil)

					// Encode to JPEG to check file size
					var buf bytes.Buffer
					switch img.Format {
					case "jpeg":
						jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85})
					case "png":
						png.Encode(&buf, dst)
					case "gif":
						gif.Encode(&buf, dst, nil)
					default:
						fmt.Printf("Unsupported image format: %s\n", img.Format)
						return
					}

					if buf.Len() <= int(info.Size*1024) {
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

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// extractChapterFromFolderName extracts the chapter number from the folder name.
func extractChapterFromFolderName(s string) (int, error) {
	re := regexp.MustCompile(`^(\d+)_`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid folder name format")
	}

	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return n, nil
}

// extractChapterFromFileName extracts the chapter number from the file name.
func extractChapterFromFileName(s string) (int, error) {
	re := regexp.MustCompile(`_(\d{4})_`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid file name format")
	}

	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return n, nil
}
