// Package markdown provides an markdown bubble which can render
// markdown in a pretty manor.
package markdown

import (
	"errors"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/knipferrc/teacup/dirfs"
)

type renderMarkdownMsg string
type errorMsg error

// Constants used throughout.
const (
	Padding = 1
)

// Bubble represents the properties of a code bubble.
type Bubble struct {
	Viewport    viewport.Model
	BorderColor lipgloss.AdaptiveColor
	Borderless  bool
	FileName    string
	ImageString string
}

// RenderMarkdown renders the markdown content with glamour.
func RenderMarkdown(width int, content string) (string, error) {
	background := "light"

	if lipgloss.HasDarkBackground() {
		background = "dark"
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStandardStyle(background),
	)

	out, err := r.Render(content)
	if err != nil {
		return "", errors.Unwrap(err)
	}

	return out, nil
}

// renderMarkdownCmd renders text as pretty markdown.
func renderMarkdownCmd(width int, filename string) tea.Cmd {
	return func() tea.Msg {
		content, err := dirfs.ReadFileContent(filename)
		if err != nil {
			return errorMsg(err)
		}

		markdownContent, err := RenderMarkdown(width, content)
		if err != nil {
			return errorMsg(err)
		}

		return renderMarkdownMsg(markdownContent)
	}
}

// New creates a new instance of markdown.
func New(borderless bool, borderColor lipgloss.AdaptiveColor) Bubble {
	viewPort := viewport.New(0, 0)
	border := lipgloss.NormalBorder()

	if borderless {
		border = lipgloss.HiddenBorder()
	}

	viewPort.Style = lipgloss.NewStyle().
		PaddingLeft(Padding).
		PaddingRight(Padding).
		Border(border).
		BorderForeground(borderColor)

	return Bubble{
		Viewport:    viewPort,
		Borderless:  borderless,
		BorderColor: borderColor,
	}
}

// Init initializes the code bubble.
func (b Bubble) Init() tea.Cmd {
	return nil
}

// SetFileName sets current file to render, this
// returns a cmd which will render the text.
func (b *Bubble) SetFileName(filename string) tea.Cmd {
	b.FileName = filename

	return renderMarkdownCmd(b.Viewport.Width, filename)
}

// SetBorderColor sets the current color of the border.
func (b *Bubble) SetBorderColor(color lipgloss.AdaptiveColor) {
	b.BorderColor = color
}

// SetSize sets the size of the bubble.
func (b *Bubble) SetSize(w, h int) tea.Cmd {
	b.Viewport.Width = w - b.Viewport.Style.GetHorizontalFrameSize()
	b.Viewport.Height = h - b.Viewport.Style.GetVerticalFrameSize()

	border := lipgloss.NormalBorder()

	if b.Borderless {
		border = lipgloss.HiddenBorder()
	}

	b.Viewport.Style = lipgloss.NewStyle().
		PaddingLeft(Padding).
		PaddingRight(Padding).
		Border(border).
		BorderForeground(b.BorderColor)

	if b.FileName != "" {
		return renderMarkdownCmd(b.Viewport.Width, b.FileName)
	}

	return nil
}

// Update handles updating the UI of a code bubble.
func (b Bubble) Update(msg tea.Msg) (Bubble, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case renderMarkdownMsg:
		b.Viewport.SetContent(string(msg))

		return b, nil
	case errorMsg:
		b.FileName = ""
		b.Viewport.SetContent(msg.Error())

		return b, nil
	}

	b.Viewport, cmd = b.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return b, tea.Batch(cmds...)
}

// View returns a string representation of the markdown bubble.
func (b Bubble) View() string {
	border := lipgloss.NormalBorder()

	if b.Borderless {
		border = lipgloss.HiddenBorder()
	}

	b.Viewport.Style = lipgloss.NewStyle().
		PaddingLeft(Padding).
		PaddingRight(Padding).
		Border(border).
		BorderForeground(b.BorderColor)

	return b.Viewport.View()
}
