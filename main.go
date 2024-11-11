package main

import (
	"fmt"
	"strconv"

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

	folderPathLabel := widget.NewLabel("No folder selected")

	selectFolderBtn := widget.NewButton("Select Folder", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if list == nil {
				return
			}
			folderPathLabel.SetText(list.Path())
		}, myWindow)
	})

	contentTypeSelect := widget.NewSelect([]string{"comic", "magazine_comic"}, func(value string) {
		fmt.Println("Content type selected:", value)
	})
	contentTypeSelect.SetSelected("comic")

	percentEntry := widget.NewEntry()
	percentEntry.SetText("98.0")
	percentEntry.SetPlaceHolder("Enter percent (1-100)")

	progress := widget.NewProgressBar()
	progress.Hide()

	var runButton *widget.Button
	runButton = widget.NewButton("Run", func() {
		folderPath := folderPathLabel.Text
		contentType := contentTypeSelect.Selected
		if folderPath == "No folder selected" {
			dialog.ShowInformation("Error", "Please select a folder first", myWindow)
			return
		}

		percent, err := strconv.ParseFloat(percentEntry.Text, 64)
		if err != nil || percent < 1 || percent > 100 {
			dialog.ShowInformation("Error", "Please enter a valid percent value (1-100)", myWindow)
			return
		}

		fmt.Printf("Processing folder: %s as %s with percent: %.2f%%\n", folderPath, contentType, percent)

		runButton.Disable()
		progress.Show()
		progress.SetValue(0)

		go func() {
			var renameSuffixN = 3 // suffix file with number
			limitInfo := pipeline.ContentTypeByLimitInfo[contentType]

			// calc file counting
			files := pipeline.Load(folderPath)
			var fileCount int
			for range files {
				fileCount++
			}

			// start pipeline
			files = pipeline.Load(folderPath)

			processedCount := 0
			updateProgress := func() {
				myWindow.Canvas().Content().Refresh()
				progress.SetValue(float64(processedCount) / float64(fileCount))
			}

			pipeline.Save( // save image file
				pipeline.Resize( // resize image file
					pipeline.Decode( // decode image file
						pipeline.Rename( // rename image file
							files,
							renameSuffixN,
						),
					),
					limitInfo.Width,
					limitInfo.Size,
					percent,
				),
			)

			for range files {
				processedCount++
				updateProgress()
			}

			myWindow.Canvas().Content().Refresh()
			dialog.ShowInformation("Complete", "MagicX processing has been completed.", myWindow)
			runButton.Enable()
			progress.Hide()

			myApp.SendNotification(&fyne.Notification{
				Title:   "Process Complete",
				Content: "MagicX processing has been completed.",
			})
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("Selected Folder:"),
		folderPathLabel,
		selectFolderBtn,
		widget.NewLabel("Content Type:"),
		contentTypeSelect,
		widget.NewLabel("Reduce size:"),
		percentEntry,
		runButton,
		progress,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
