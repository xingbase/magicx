package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/xingbase/magicx"
	"github.com/xingbase/magicx/file"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("MagicX v1.4.1")

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
			output := magicx.Reanme(magicx.Load(folderPath))

			limited := magicx.LimitedSizeInfoByContentType[contentType]

			folders := make(map[string]struct{}, 0)
			images := make(map[string]struct{}, 0)
			sizeOver := make(map[string]struct{}, 0)
			thumbs := make(map[string]struct{}, 0)
			mismatch := make(map[string]struct{}, 0)
			notFoundThumbs := make(map[string]struct{}, 0)
			notFoundImages := make(map[string]struct{}, 0)
			noNumberings := make(map[string]struct{}, 0)

			for folderInfos := range output {
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
							delete(notFoundThumbs, episodeName)
						} else {
							// First pass: Group images by width and find the most common width
							width := f.Width
							groupedImages[width] = append(groupedImages[width], f)
							widthCounts[width]++

							if widthCounts[width] > maxCount {
								maxCount = widthCounts[width]
								standardWidth = width
							}

							fileN, _ := file.ExtractFileExtNum(f.Name, f.Ext)

							imageFileNums = append(imageFileNums, fileN)
						}
					}

					if !file.IsConsecutive(imageFileNums) {
						noNumberings[episodeName] = struct{}{}
					}

					if len(imageFileNums) == 0 {
						notFoundImages[episodeName] = struct{}{}
					}

					// Second pass: Process grouped images and determine if they are standard
					for width, imgs := range groupedImages {
						isStandardWidth := (width == standardWidth)
						for _, img := range imgs {
							processedImg := img
							processedImg.IsStandard = isStandardWidth

							// Check size against limit
							if img.Size > limited.Image.Size {
								sizeOver[episodeName] = struct{}{}
							}

							if !processedImg.IsStandard {
								images[episodeName] = struct{}{}
							}
						}
					}
				}
			}

			myWindow.Canvas().Content().Refresh()
			resultTextArea.SetText(magicx.ConsoleLog(folders, images, sizeOver, thumbs, mismatch, notFoundThumbs, notFoundImages, noNumberings)) // Set the results in the textarea
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
