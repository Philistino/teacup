package filetree

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/philistino/teacup/dirfs"
)

const (
	yesKey   = "y"
	enterKey = "enter"
)

// Update handles updating the filetree.
func (b Bubble) Update(msg tea.Msg) (Bubble, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
	case getDirectoryListingMsg:
		if msg != nil {
			cmd = b.list.SetItems(msg)
			cmds = append(cmds, cmd)
		}
	case copyToClipboardMsg:
		return b, b.list.NewStatusMessage(statusMessageInfoStyle(string(msg)))
	case errorMsg:
		return b, b.list.NewStatusMessage(statusMessageErrorStyle(msg.Error()))
	case tea.KeyMsg:
		if b.IsFiltering() {
			break
		}

		if !b.active {
			return b, nil
		}
		switch {
		case key.Matches(msg, openDirectoryKey):
			if !b.input.Focused() {
				selectedDir := b.GetSelectedItem()
				cmds = append(cmds, getDirectoryListingCmd(selectedDir.fileName, b.showHidden, b.showIcons))
			}
		case key.Matches(msg, selectItemKey):
			if !b.input.Focused() {
				selectedDir := b.GetSelectedItem()
				cmds = append(cmds, getDirectoryListingCmd(selectedDir.fileName, b.showHidden, b.showIcons))
			}
		case key.Matches(msg, toggleHiddenKey):
			if !b.input.Focused() {
				b.showHidden = !b.showHidden
				cmds = append(cmds, getDirectoryListingCmd(dirfs.CurrentDirectory, b.showHidden, b.showIcons))
			}
		case key.Matches(msg, homeShortcutKey):
			if !b.input.Focused() {
				cmds = append(cmds, getDirectoryListingCmd(dirfs.HomeDirectory, b.showHidden, b.showIcons))
			}
		case key.Matches(msg, rootShortcutKey):
			if !b.input.Focused() {
				cmds = append(cmds, getDirectoryListingCmd(dirfs.RootDirectory, b.showHidden, b.showIcons))
			}

		case key.Matches(msg, escapeKey):
			b.state = idleState

			if b.input.Focused() {
				b.input.Reset()
				b.input.Blur()
			}
		case key.Matches(msg, selectItemKey):
			selectedItem := b.GetSelectedItem()

			switch b.state {
			case idleState, deleteItemState, moveItemState:
				return b, nil
			case createFileState:
				statusCmd := b.list.NewStatusMessage(
					statusMessageInfoStyle("Successfully created file"),
				)

				cmds = append(cmds, statusCmd, tea.Sequentially(
					createFileCmd(b.input.Value()),
					getDirectoryListingCmd(dirfs.CurrentDirectory, b.showHidden, b.showIcons),
				))
			case createDirectoryState:
				statusCmd := b.list.NewStatusMessage(
					statusMessageInfoStyle("Successfully created directory"),
				)

				cmds = append(cmds, statusCmd, tea.Sequentially(
					createDirectoryCmd(b.input.Value()),
					getDirectoryListingCmd(dirfs.CurrentDirectory, b.showHidden, b.showIcons),
				))
			case renameItemState:
				statusCmd := b.list.NewStatusMessage(
					statusMessageInfoStyle("Successfully renamed"),
				)

				cmds = append(cmds, statusCmd, tea.Sequentially(
					renameItemCmd(selectedItem.fileName, b.input.Value()),
					getDirectoryListingCmd(dirfs.CurrentDirectory, b.showHidden, b.showIcons),
				))
			}

			b.state = idleState
			b.input.Blur()
			b.input.Reset()
		}
	}

	if b.active {
		switch b.state {
		case idleState, moveItemState:
			b.list, cmd = b.list.Update(msg)
			cmds = append(cmds, cmd)
		case createFileState, createDirectoryState, renameItemState:
			b.input, cmd = b.input.Update(msg)
			cmds = append(cmds, cmd)
		case deleteItemState:
			return b, nil
		}
	}

	return b, tea.Batch(cmds...)
}
