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
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Tab: Next  Shift+Tab: Back  Enter: Show Directory  B: Back Directory  Space: Select/Unselect File  L: Show File  Q: Close File")

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
			currentDir := "/home/" + username
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

	selectedFiles := make(map[int]struct{})
	list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	for _, file := range files {
		list.AddItem(file, "", 0, nil)
	}

	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if _, ok := selectedFiles[index]; ok {
			delete(selectedFiles, index)
			list.SetItemText(index, files[index], "")
		} else {
			selectedFiles[index] = struct{}{}
			list.SetItemText(index, "[blue]"+files[index]+"[white]", "")
		}
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			selectedItem, _ := list.GetItemText(list.GetCurrentItem())
			selectedPath := filepath.Join(path, selectedItem)
			isDir, err := isValidDirectory(username, password, server, selectedPath)
			if err != nil || !isDir {
				return nil
			}
			navigateDir(selectedPath, username, password, server, app, rootFlex, form)
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'l' || event.Rune() == 'L' {
				selectedItem, _ := list.GetItemText(list.GetCurrentItem())
				selectedPath := filepath.Join(path, selectedItem)
				showFilePreview(username, password, server, selectedPath, app, rootFlex, form)
				return nil
			}
		}
		return event
	})

	rootFlex.AddItem(list, 0, 1, true)
	app.SetFocus(list)
}

func showFilePreview(username, password, server, path string, app *tview.Application, rootFlex *tview.Flex, form *tview.Form) {
	content := getFileContent(username, password, server, path)
	if content == "" {
		return
	}

	previewPane := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(content)

	// Adding the preview pane to the right side of the rootFlex
	rootFlex.AddItem(previewPane, 0, 1, true)
	app.SetFocus(previewPane)
}

func getFileContent(username, password, server, path string) string {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return ""
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return ""
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("cat \"" + path + "\""); err != nil {
		return ""
	}

	return b.String()
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

func isValidDirectory(username, password, server, path string) (bool, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 本番環境では適切なホストキーの検証を行うこと
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return false, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	cmd := "ls -ld \"" + path + "\""
	if err := session.Run(cmd); err != nil {
		return false, err
	}

	output := b.String()
	if len(output) > 0 && output[0] == 'd' {
		return true, nil // ディレクトリである
	}
	return false, nil // ディレクトリではない
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
