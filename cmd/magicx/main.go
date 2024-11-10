package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct{}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stdout)
			parser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
	// path := "/Users/JP17278/Downloads/00022_sansyoku"
	// // pipeline
	// files := magicx.Load(path)
	// magicx.Save(magicx.Resize(magicx.Decode(magicx.Rename(files)), 30720, 95.0))
}
