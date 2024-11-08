package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/xingbase/magicx"
)

func main() {
	path := "/Users/JP17278/Downloads/00022_sansyoku"
	// pipeline
	files := magicx.Load(path)
	magicx.Process(magicx.Resize(magicx.Decode(magicx.Rename(files))))
}
