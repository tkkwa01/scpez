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

	// ルートとなるFlexコンテナ
	rootFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	form := tview.NewForm()
	form.AddInputField("Server", "", 20, nil, nil).
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Connect", func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			username := form.GetFormItem(1).(*tview.InputField).GetText()
			password := form.GetFormItem(2).(*tview.InputField).GetText()
			currentDir := "/home/" + username // starting directory
			navigateDir(currentDir, username, password, server, app, rootFlex)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	rootFlex.AddItem(form, 0, 1, true)

	if err := app.SetRoot(rootFlex, true).Run(); err != nil {
		panic(err)
	}
}

func navigateDir(path, username, password, server string, app *tview.Application, rootFlex *tview.Flex) {
	files, err := connectAndListFiles(username, password, server, path)
	if err != nil {
		panic(err) // Proper error handling is needed
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
			navigateDir(newPath, username, password, server, app, rootFlex)
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'b' || event.Rune() == 'B' { // Listen for 'B' key
				if rootFlex.GetItemCount() > 1 {
					rootFlex.RemoveItem(rootFlex.GetItem(rootFlex.GetItemCount() - 1))
				}
				return nil
			}
		}
		// Pass other key events to the default handler
		return event
	})

	// 新しいリストを追加し、フォーカスをそのリストに移動
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
