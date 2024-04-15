package main

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"strings"
)

func main() {
	app := tview.NewApplication()
	rootFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	title := tview.NewTextView().
		SetText("SCP-Ez").
		SetTextAlign(tview.AlignCenter)
	rootFlex.AddItem(title, 3, 1, false)
	title.SetBorder(true)

	mainContentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	form := createForm(app, mainContentFlex)
	mainContentFlex.AddItem(form, 0, 1, true)
	rootFlex.AddItem(mainContentFlex, 0, 1, true)

	helpText := tview.NewTextView().
		SetDynamicColors(true). // ダイナミックカラーを有効にする
		SetTextAlign(tview.AlignCenter).
		SetText("Tab: Next  Shift+Tab: Back  Enter: Show Directory  B: Back Directory")

	helpText.SetBackgroundColor(tcell.ColorBlue)

	rootFlex.AddItem(helpText, 1, 1, false)

	if err := app.SetRoot(rootFlex, true).Run(); err != nil {
		panic(err)
	}
}

func createForm(app *tview.Application, rootFlex *tview.Flex) *tview.Form {
	form := tview.NewForm()
	form.AddInputField("Server", "", 20, nil, nil).
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Connect", func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			username := form.GetFormItem(1).(*tview.InputField).GetText()
			password := form.GetFormItem(2).(*tview.InputField).GetText()
			currentDir := "/home/" + username // starting directory
			navigateDir(currentDir, username, password, server, app, rootFlex, form)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})
	return form
}

func navigateDir(path, username, password, server string, app *tview.Application, rootFlex *tview.Flex, form *tview.Form) {
	files, err := connectAndListFiles(username, password, server, path)
	if err != nil {
		showModal(app, "Failed to connect: "+err.Error(), rootFlex, form)
		return
	}

	list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	for _, file := range files {
		list.AddItem(file, "", 0, nil)
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			selectedFile, _ := list.GetItemText(list.GetCurrentItem())
			newPath := filepath.Join(path, selectedFile)
			navigateDir(newPath, username, password, server, app, rootFlex, form)
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'b' || event.Rune() == 'B' {
				if rootFlex.GetItemCount() > 1 {
					rootFlex.RemoveItem(rootFlex.GetItem(rootFlex.GetItemCount() - 1))
					app.SetFocus(rootFlex.GetItem(rootFlex.GetItemCount() - 1))
				} else {
					app.SetFocus(form)
				}
				return nil
			}
		}
		return event
	})

	rootFlex.AddItem(list, 0, 1, true)
	app.SetFocus(list)
}

func connectAndListFiles(username, password, server, dir string) ([]string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Not safe, adjust appropriately
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("ls \"" + dir + "\""); err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(b.String()), "\n"), nil
}

func showModal(app *tview.Application, message string, rootFlex *tview.Flex, form *tview.Form) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(rootFlex, true)
			app.SetFocus(form)
		})
	app.SetRoot(modal, false)
	app.SetFocus(modal)
}
