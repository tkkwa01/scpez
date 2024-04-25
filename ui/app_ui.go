package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"path/filepath"
	"scpez/entities"
	"scpez/interfaces"
	"scpez/usecases"
)

// AppUI manages the user interface for the application.
type AppUI struct {
	app              *tview.Application
	pages            *tview.Pages
	serverInteractor *usecases.ServerInteractor
	user             entities.User   // User information
	server           entities.Server // Server information
}

// NewAppUI creates a new UI instance.
func NewAppUI(client interfaces.SSHClient) *AppUI {
	interactor := usecases.NewServerInteractor(client)
	app := tview.NewApplication()
	pages := tview.NewPages()

	return &AppUI{
		app:              app,
		pages:            pages,
		serverInteractor: interactor,
	}
}

// Start runs the UI application.
func (ui *AppUI) Start() error {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Title
	title := tview.NewTextView().SetText("SSH File Manager").SetTextAlign(tview.AlignCenter)
	flex.AddItem(title, 1, 1, false)

	// Help text
	help := tview.NewTextView().SetText("Use arrow keys to navigate, Enter to select, 'q' to quit").SetDynamicColors(true)
	help.SetBackgroundColor(tcell.ColorDarkCyan)
	flex.AddItem(help, 1, 1, false)

	// Main content area
	mainView := tview.NewList().SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ui.updateDetail(mainText)
	})
	flex.AddItem(mainView, 0, 1, true)

	ui.pages.AddPage("main", flex, true, true)
	ui.app.SetRoot(ui.pages, true)

	// Run the application
	return ui.app.Run()
}

// updateDetail updates the details view based on the selected item.
func (ui *AppUI) updateDetail(itemName string) {
	// This function would interact with the ServerInteractor to get details about the item.
}

// Stop stops the UI application.
func (ui *AppUI) Stop() {
	ui.app.Stop()
}

// FetchAndDisplayFiles fetches the file list from the server and displays it in the main view.
func (ui *AppUI) FetchAndDisplayFiles(server entities.Server, user entities.User, path string) {
	files, err := ui.serverInteractor.ListFiles(user, server, path)
	if err != nil {
		ui.showModal("Error", fmt.Sprintf("Failed to list files: %v", err))
		return
	}

	mainView := tview.NewList().ShowSecondaryText(false)
	for _, file := range files {
		mainView.AddItem(file.Name, "", 0, nil)
	}

	mainView.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if files[index].IsDirectory {
			// Navigate into the directory
			newPath := filepath.Join(path, files[index].Name)
			ui.FetchAndDisplayFiles(server, user, newPath)
		} else {
			// Show file details or open file
			ui.showFileDetails(files[index])
		}
	})

	ui.pages.SwitchToPage("main")
	ui.app.SetRoot(mainView, true)
}

// showModal displays a modal dialog with a message.
func (ui *AppUI) showModal(title, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ui.pages.SwitchToPage("main")
		})
	ui.app.SetRoot(modal, false)
}

// showFileDetails displays the details of a file.
func (ui *AppUI) showFileDetails(file entities.FileItem) {
	// Assuming we have a function to fetch the content or details of the file
	content := "File content or metadata here"
	detailsView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(content)
	ui.pages.AddPage("details", detailsView, true, true)
	ui.app.SetRoot(detailsView, true)
}

// SetupUserAndServer configures the user and server information and fetches the initial file list.
func (ui *AppUI) SetupUserAndServer(user entities.User, server entities.Server) {
	ui.user = user
	ui.server = server
	// ここで初期ディレクトリ（例えばユーザーのホームディレクトリ）のファイルリストを取得して表示
	initialPath := "/home/" + user.Username
	ui.FetchAndDisplayFiles(server, user, initialPath)
}
