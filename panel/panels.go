package panel

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Panel struct {
	// Define panel properties and methods here
	CurrentDir           string
	PromptLabel          *widget.Label
	PromptEditor         *widget.Entry
	ResultField          *widget.Entry
	SavePromptButton     *widget.Button
	CopyResultButton     *widget.Button
	OutputField          *widget.Entry
	OutputPathSelector   *widget.Button
	OutputDirectoryLabel *widget.Label
	Window               *fyne.Window
}
