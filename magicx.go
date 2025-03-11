package magicx

import (
	"fmt"
	_ "image/gif"  //   Import GIF decoder
	_ "image/jpeg" // Import JPEG decoder
	_ "image/png"  // Import PNG decoder
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/xingbase/magicx/file"
)

const (
	EN Language = iota
	JP
)

type Language int8

var LimitedSizeInfoByContentType = map[string]LimitedSizeInfo{
	"comic": {
		Image:     ImageSize{Width: 1600, Size: 10485760}, // 10MB
		Thumbnail: ThumbnailSize{Width: 500, Size: 51200}, // 50KB
		Folder:    62914560,                               // 60MB
	},
	"magazine_comic": {
		Image:     ImageSize{Width: 2266, Size: 31457280}, // 30MB
		Thumbnail: ThumbnailSize{Width: 500, Size: 51200}, // 50KB
		Folder:    62914560,                               // 60MB
	},
}

type LimitedSizeInfo struct {
	Folder    int64
	Image     ImageSize
	Thumbnail ThumbnailSize
}

type ImageSize struct {
	Width int
	Size  int64
}
type ThumbnailSize struct {
	Width int
	Size  int64
}

type FolderInfo struct {
	Name  string
	Size  int64
	Files []FileInfo
}

type FileInfo struct {
	Path        string
	Folder      string
	Name        string
	Ext         string
	Size        int64
	Width       int
	Height      int
	Format      string
	IsStandard  bool
	IsThumbnail bool
	IsMissmatch bool
}

func (f FileInfo) FullName() string {
	return f.Path + "/" + f.Name
}

func Load(dir string) <-chan []FolderInfo {
	out := make(chan []FolderInfo)

	go func() {
		defer close(out)

		files := make(map[string][]FileInfo)

		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))

				if file.Extensions[ext] {
					rPath, _ := filepath.Rel(dir, filepath.Dir(path))

					// extract episode folder name
					parts := strings.Split(rPath, "/")
					folder := parts[len(parts)-1]

					// parsing for image metadata
					img, err := file.ParseImage(path)
					if err != nil {
						fmt.Printf("Failed to parse image %s: %v\n", path, err)
					}

					fileInfo := FileInfo{
						Path:        filepath.Dir(path),
						Folder:      folder,
						Name:        info.Name(),
						Ext:         ext,
						Size:        info.Size(),
						Width:       img.Width,
						Height:      img.Height,
						Format:      img.Format,
						IsStandard:  true,
						IsThumbnail: file.HasThumbnail(info.Name()),
					}

					if !fileInfo.IsThumbnail {
						fileInfo.IsMissmatch = file.HasMismatch(folder, info.Name())
					}

					files[folder] = append(files[folder], fileInfo)
				}
			}

			return nil
		})
		if err != nil {
			fmt.Println("Error walking through directory: ", err)
		}

		data := make([]FolderInfo, 0)
		for folder, fileInfos := range files {
			var folderSize int64
			for _, file := range fileInfos {
				folderSize += file.Size
			}

			folderInfo := FolderInfo{
				Name:  folder,
				Size:  folderSize,
				Files: fileInfos,
			}
			data = append(data, folderInfo)
		}

		out <- data
	}()

	return out
}

func Reanme(in <-chan []FolderInfo) <-chan []FolderInfo {
	out := make(chan []FolderInfo)

	go func() {
		defer close(out)

		for folderInfos := range in {
			for i := range folderInfos {
				for j, file := range folderInfos[i].Files {
					// split file names with “_”
					parts := strings.Split(file.Name, "_")

					if len(parts) > 0 {
						// extract the last numbering part
						last := parts[len(parts)-1]
						num := strings.TrimSuffix(last, file.Ext)

						// add padding if the num is less then 3 digits
						if len(num) < 3 {
							newNum := fmt.Sprintf("%03s", num)
							newName := strings.Replace(file.Name, num+file.Ext, newNum+file.Ext, 1)
							newFile := file.Path + "/" + newName

							// try to rename the file
							err := os.Rename(file.FullName(), newFile)
							if err != nil {
								fmt.Printf("Failed to rename file %s: %v\n", file.Name, err)
							}

							// update to the new file name
							folderInfos[i].Files[j].Name = newName
						}
					}
				}
			}
			// send results to an output channel
			out <- folderInfos
		}
	}()

	return out
}

func EpisodeName(n int, lang Language) string {
	var name string

	switch lang {
	case EN:
		name = fmt.Sprintf("%d", n)
	case JP:
		name = fmt.Sprintf("%d話", n)
	}

	return name
}

func Println(title string, data map[string]struct{}) {
	if len(data) > 0 {
		fmt.Println(title)
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Println(strings.Join(keys, ", "))
		fmt.Println()
	}
}

func ConsoleLog(folders, images, thumbs, mismatch, notFoundThumbs, noNumberings map[string]struct{}) string {
	var results strings.Builder

	logging := func(title string, items map[string]struct{}) {
		if len(items) > 0 {
			results.WriteString(fmt.Sprintf("\n# %s\n", title))
			keys := make([]string, 0, len(items))
			for k := range items {
				keys = append(keys, k)
			}

			sort.Slice(keys, func(i, j int) bool {
				numI := extractNumber(keys[i])
				numJ := extractNumber(keys[j])
				return numI < numJ
			})

			results.WriteString(strings.Join(keys, ", "))
			results.WriteString("\n")
		}
	}

	logging("1話の容量が60MBを超えていた話", folders)
	logging("1話内で横幅が統一されていない話", images)
	logging("話サムネの容量が50KB以上になっていた話", thumbs)
	logging("フォルダ名とファイル名一致していない話", mismatch)
	logging("サムネがない話", notFoundThumbs)
	logging("ページ表記が順番でなってない話", noNumberings)
	return results.String()
}

func extractNumber(s string) int {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(s)
	if numStr == "" {
		return 0
	}
	// 문자열을 정수로 변환
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}
