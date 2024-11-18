package main

import (
	"fmt"

	"github.com/xingbase/magicx/pipeline"
)

func main() {
	// dir := "/Users/JP17278/Downloads/7153_冬の夜/00001_fuyu"
	dir := "/Users/JP17278/Downloads/00022_sansyoku"

	limitInfo := pipeline.ContentTypeByLimitInfo["comic"]
	images := pipeline.CheckImageSize(pipeline.Decode(pipeline.Rename(pipeline.Load(dir), 3)), limitInfo)

	for img := range images {
		if !img.IsStandard {
			fmt.Printf("%s width: %d, height: %d, size: %d\n", img.Path, img.Image.Bounds().Dx(), img.Image.Bounds().Dy(), img.Size)
		}
	}

}
