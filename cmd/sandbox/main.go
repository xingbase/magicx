package main

import (
	_ "image/gif"  //   Import GIF decoder
	_ "image/jpeg" // Import JPEG decoder
	_ "image/png"  // Import PNG decoder

	"github.com/xingbase/magicx"
	"github.com/xingbase/magicx/file"
)

func main() {
	dir := "/Users/JP17278/Downloads/data"

	result := magicx.Reanme(magicx.Load(dir))

	limited := magicx.LimitedSizeInfoByContentType["comic"]

	underImages := make(map[string]struct{}, 0)
	underThumbs := make(map[string]struct{}, 0)
	folders := make(map[string]struct{}, 0)
	images := make(map[string]struct{}, 0)
	thumbs := make(map[string]struct{}, 0)
	mismatch := make(map[string]struct{}, 0)
	notFoundThumbs := make(map[string]struct{}, 0)
	notNumberings := make(map[string]struct{}, 0)

	for folderInfos := range result {
		for i := range folderInfos {
			n, _ := file.ExtractFolderNum(folderInfos[i].Name)
			if n == 0 {
				continue
			}

			episodeName := magicx.EpisodeName(n, magicx.JP)

			notFoundThumbs[episodeName] = struct{}{}

			if folderInfos[i].Size > limited.Folder {
				folders[episodeName] = struct{}{}
			}

			groupedImages := make(map[int][]magicx.FileInfo)
			widthCounts := make(map[int]int)
			maxCount := 0
			standardWidth := 0

			imageFileNums := []int{}

			for _, f := range folderInfos[i].Files {
				if f.IsMissmatch {
					mismatch[episodeName] = struct{}{}
				}

				if f.IsThumbnail {
					if f.Size > limited.Thumbnail.Size {
						thumbs[episodeName] = struct{}{}
					}
					if f.Size < magicx.UnderImageSize {
						underThumbs[episodeName] = struct{}{}
					}

					delete(notFoundThumbs, episodeName)
				} else {
					width := f.Width
					groupedImages[width] = append(groupedImages[width], f)
					widthCounts[width]++

					if widthCounts[width] > maxCount {
						maxCount = widthCounts[width]
						standardWidth = width
					}

					fileN, _ := file.ExtractFileNum(f.Name)
					imageFileNums = append(imageFileNums, fileN)
				}
			}

			if !file.IsConsecutive(imageFileNums) {
				notNumberings[episodeName] = struct{}{}
			}

			// Second pass: Process grouped images and determine if they are standard
			for width, imgs := range groupedImages {
				isStandardWidth := (width == standardWidth)
				for _, img := range imgs {
					processedImg := img
					processedImg.IsStandard = isStandardWidth

					// Check size against limit
					if img.Size > limited.Image.Size {
						processedImg.IsStandard = false
					}
					if img.Size < magicx.UnderImageSize {
						underImages[episodeName] = struct{}{}
					}

					if !processedImg.IsStandard {
						images[episodeName] = struct{}{}
					}
				}
			}
		}
	}

	magicx.Println("1話の容量が60MBを超えていた話", folders)
	magicx.Println("1話内で横幅が統一されていない話", images)
	magicx.Println("話サムネの容量が50KB以上になっていた話", thumbs)
	magicx.Println("1話の容量が5KB以下になっていた話", underImages)
	magicx.Println("話サムネの容量が5KB以下になっていた話", underThumbs)
	magicx.Println("フォルダ名とファイル名一致していない話", mismatch)
	magicx.Println("サムネがない話", notFoundThumbs)
	magicx.Println("ページ表記が順番でなってない話", notNumberings)

}
