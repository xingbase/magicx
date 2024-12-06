package file

import (
	"fmt"
	"image"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var Extensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

type Image struct {
	Format string
	Width  int
	Height int
}

func ExtractFolderNum(s string) (int, error) {
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

func ExtractFileNum(s string) (int, error) {
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

func FormatSize(bytes int64) string {
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

func HasThumbnail(s string) bool {
	return strings.HasPrefix(s, "tmb")
}

func HasMismatch(folder string, file string) bool {
	a, _ := ExtractFolderNum(folder)
	b, _ := ExtractFileNum(file)
	return a != b
}

func ParseImage(path string) (Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return Image{}, err
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return Image{}, err
	}

	return Image{
		Format: format,
		Width:  config.Width,
		Height: config.Height,
	}, nil
}
