package main

import (
	"fmt"

	"github.com/xingbase/magicx"
)

func init() {
	parser.AddCommand("rename",
		"Rename to file",
		"The rename command-line fix the numbering file.",
		&renameCommand{})
}

type renameCommand struct {
	Path string `short:"p" long:"path" description:"Full path" default:"/Users/JP17278/Downloads/00022_sansyoku"`
	Num  int    `short:"n" long:"num" description:"Suffix number" default:"3"`
}

func (c *renameCommand) Execute(args []string) error {
	files := magicx.Load(c.Path)

	for file := range magicx.Rename(files, c.Num) {
		fmt.Println(file.Name)
	}

	return nil
}
