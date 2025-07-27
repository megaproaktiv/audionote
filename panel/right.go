package panel

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (p *Panel) RightPanel() *container.AppTabs {
	rightPanel := container.NewAppTabs(
		// Left tab: Prompt Editor
		container.NewTabItem("Prompt Editor",
			container.NewBorder(
				// Top: Just the label
				container.NewPadded(p.PromptLabel),
				// Bottom: Centered save button
				container.NewPadded(
					container.NewHBox(
						layout.NewSpacer(),
						p.SavePromptButton,
						layout.NewSpacer(),
					),
				),
				// Left, Right: nil
				nil, nil,
				// Center: Maximized scrollable editor
				container.NewScroll(p.PromptEditor),
			),
		),
		// Right tab: Result
		container.NewTabItem("Result",
			container.NewBorder(
				// Top: Result label
				container.NewPadded(widget.NewLabel("Processing Result")),
				// Bottom: nil
				nil,
				// Left, Right: nil
				nil, nil,
				// Center: Scrollable result field
				container.NewScroll(p.ResultField),
			),
		),
	)
	return rightPanel
}
