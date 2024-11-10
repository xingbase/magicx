package main

import "github.com/xingbase/magicx"

func init() {
	parser.AddCommand("resize",
		"Resize to image",
		"The resize command-line",
		&reiszeCommand{},
	)
}

type reiszeCommand struct {
	Path    string  `short:"p" long:"path" description:"Full path" default:"data"`
	Width   int     `short:"w" long:"width" description:"Limit width" default:"2266"`
	Size    int64   `short:"s" long:"size" description:"Limit size (kb)" default:"30720"`
	Percent float64 `long:"percent" description:"Resize percentages" default:"95.0"`
}

func (c *reiszeCommand) Execute(args []string) error {
	files := magicx.Load(c.Path)

	magicx.Save(
		magicx.Resize(
			magicx.Decode(files),
			c.Width,
			c.Size,
			c.Percent,
		),
	)

	return nil
}
