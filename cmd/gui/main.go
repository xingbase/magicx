package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/xingbase/magicx/pipeline"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("MagicX")

	folderPathEntry := widget.NewEntry()
	folderPathEntry.SetPlaceHolder("Enter folder path")

	contentTypeSelect := widget.NewSelect([]string{"comic", "magazine_comic"}, func(value string) {
		fmt.Println("Content type selected:", value)
	})
	contentTypeSelect.SetSelected("comic")

	progress := widget.NewProgressBar()
	progress.Hide()

	resultTextArea := widget.NewMultiLineEntry()

	resultScroll := container.NewScroll(resultTextArea)
	resultScroll.SetMinSize(fyne.NewSize(800, 600))

	var runButton *widget.Button
	runButton = widget.NewButton("Run", func() {
		folderPath := folderPathEntry.Text
		contentType := contentTypeSelect.Selected
		if folderPath == "No folder selected" {
			dialog.ShowInformation("Error", "Please enter a folder path", myWindow)
			return
		}

		fmt.Printf("Processing folder: %s as %s\n", folderPath, contentType)

		runButton.Disable()
		resultTextArea.SetText("") // Clear previous results

		go func() {
			var renameSuffixN = 3 // suffix file with number
			limitInfo := pipeline.ContentTypeByLimitInfo[contentType]

			// start pipeline
			files := pipeline.Load(folderPath)

			var results strings.Builder

			out := pipeline.CheckImage(pipeline.Decode(pipeline.Rename(files, renameSuffixN)), limitInfo)

			mismatchFolders := make(map[string]struct{})
			images := make([]pipeline.ImageInfo, 0)
			thumbs := make([]pipeline.ImageInfo, 0)

			for img := range out {
				if img.IsStandard {
					continue
				}
				if img.IsMissmatch {
					mismatchFolders[img.Path] = struct{}{}
				}
				if img.IsThumbnail {
					thumbs = append(thumbs, img)
				} else {
					images = append(images, img)
				}
			}

			if len(thumbs) > 0 {
				results.WriteString("\n### THUMBNAIL \n")
			}
			for _, img := range thumbs {
				results.WriteString(fmt.Sprintf("%s  (width: %d height: %d size: %s)\n", img.Name, img.Image.Bounds().Dx(), img.Image.Bounds().Dy(), pipeline.FormatFileSize(img.Size)))
			}

			if len(images) > 0 {
				results.WriteString("\n### IMAGE \n")
			}
			for _, img := range images {
				results.WriteString(fmt.Sprintf("%s  (width: %d height: %d size: %s)\n", img.Name, img.Image.Bounds().Dx(), img.Image.Bounds().Dy(), pipeline.FormatFileSize(img.Size)))
			}

			// show missmatch folders
			if len(mismatchFolders) > 0 {
				results.WriteString("\n### FOLDER / FILE MISMATCH \n")
			}
			for folder := range mismatchFolders {
				results.WriteString(fmt.Sprintln(folder))
			}

			myWindow.Canvas().Content().Refresh()
			resultTextArea.SetText(results.String()) // Set the results in the textarea
			dialog.ShowInformation("Complete", "MagicX processing has been completed.", myWindow)
			runButton.Enable()

			myApp.SendNotification(&fyne.Notification{
				Title:   "Process Complete",
				Content: "MagicX processing has been completed.",
			})
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("Folder Path:"),
		folderPathEntry,
		widget.NewLabel("Content Type:"),
		contentTypeSelect,
		runButton,
		widget.NewLabel("Results:"),
		resultScroll,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
