package info

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/megaproaktiv/audionote-config/panel"
)

func ShowAboutDialog(panel panel.Panel) {
	w := panel.Window
	aboutContent := widget.NewRichTextFromMarkdown(`# Audio Note LLM

A desktop application for configuring and processing audio notes using Large Language Models.

## Features
• **Audio File Support**: Process MP3 and M4A files
• **AI Processing Actions**: Choose from various processing templates
• **Prompt Editor**: Edit and customize AI prompt templates
• **Language Support**: Multiple language configurations
• **AWS Integration**: S3 bucket and profile configuration
• **Persistent Settings**: Automatic saving of user preferences

## Technical Details
• Built with **Go** programming language
• UI framework: **Fyne v2**
• Configuration: **Viper** with YAML storage
• Cross-platform compatibility

## Version
**1.0.0** - Initial Release

---
*Built with ❤️ for efficient audio note processing*`)

	aboutDialog := dialog.NewCustom("About Audio Note LLM", "Close", aboutContent, *w)
	aboutDialog.Resize(fyne.NewSize(500, 400))
	aboutDialog.Show()
}
