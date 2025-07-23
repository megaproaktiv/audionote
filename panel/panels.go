package panel

import "fyne.io/fyne/v2/widget"

type Panel struct {
	// Define panel properties and methods here
	PromptLabel      *widget.Label
	PromptEditor     *widget.Entry
	ResultField      *widget.Entry
	SavePromptButton *widget.Button
}
