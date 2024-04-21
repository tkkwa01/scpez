package main

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
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
		SetText("Tab: Next  Shift+Tab: Back  Enter: Show Directory  B: Back Directory  Space: Select/Unselect  T: Transfer!  L: Show File  Q: Close File")

	helpText.SetBackgroundColor(tcell.ColorBlue)
	rootFlex.AddItem(helpText, 1, 1, false)

	if err := app.SetRoot(rootFlex, true).Run(); err != nil {
		panic(err)
	}
}

func createForm(app *tview.Application, rootFlex *tview.Flex) *tview.Form {
	form := tview.NewForm()

	servers := []string{"sh.edu.kutc.kansai-u.ac.jp", "Other (Specify)"}
	var server string

	serverDropdown := tview.NewDropDown().
		SetLabel("Server: ").
		SetOptions(servers, nil)

	otherServerField := tview.NewInputField().SetLabel("Specify Server: ").SetFieldWidth(20)

	form.AddFormItem(serverDropdown).
		AddFormItem(otherServerField).
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("Connect", func() {
			_, option := serverDropdown.GetCurrentOption()
			if option == "Other (Specify)" {
				server = otherServerField.GetText()
			} else {
				server = option
			}
			username := form.GetFormItem(2).(*tview.InputField).GetText()
			password := form.GetFormItem(3).(*tview.InputField).GetText()
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
	list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	for _, file := range files {
		list.AddItem(file, "", 0, nil)
	}
	if err != nil {
		showModal(app, "Connection Error: Either the server address, username, or password is incorrect. Please also check your network. Connecting to a VPN may resolve the issue.", rootFlex, form, list)
		return
	}

	selectedFiles := make(map[string]struct{})

	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		cleanText := strings.TrimSuffix(strings.TrimPrefix(mainText, "[blue]"), "[white]")

		if _, ok := selectedFiles[cleanText]; ok {
			delete(selectedFiles, cleanText)
			list.SetItemText(index, cleanText, "")
		} else {
			selectedFiles[cleanText] = struct{}{}
			list.SetItemText(index, "[blue]"+cleanText+"[white]", "")
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
			switch event.Rune() {
			case 'l', 'L':
				selectedItem, _ := list.GetItemText(list.GetCurrentItem())
				selectedPath := filepath.Join(path, selectedItem)
				showFilePreview(username, password, server, selectedPath, app, rootFlex, list)
				return nil
			case 'b', 'B':
				if rootFlex.GetItemCount() > 1 {
					rootFlex.RemoveItem(rootFlex.GetItem(rootFlex.GetItemCount() - 1))
					app.SetFocus(rootFlex.GetItem(rootFlex.GetItemCount() - 1))
				} else {
					app.SetFocus(form)
				}
				return nil
			case 't', 'T':
				for filePath := range selectedFiles {
					remotePath := filepath.Join(path, filePath)
					if err := transferFile(username, password, server, remotePath); err != nil {
						showModal(app, "Transfer Failed:"+err.Error(), rootFlex, form, list)
						return nil
					}
				}
				showModal(app, "Succeed Transfer!!", rootFlex, form, list)
				return nil
			}
		}
		return event
	})

	rootFlex.AddItem(list, 0, 1, true)
	app.SetFocus(list)
}

func transferFile(username, password, server, remotePath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return err
	}
	defer client.Close()

	// 新しいSSHセッションを開始
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// リモートファイルを開く
	srcFile, err := session.Output("cat " + "\"" + remotePath + "\"")
	if err != nil {
		return err
	}

	// ホームディレクトリの取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// ローカルファイルパスを決定
	localFilePath := filepath.Join(homeDir, filepath.Base(remotePath))
	dstFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// リモートファイルの内容をローカルファイルに書き込み
	if _, err = io.Copy(dstFile, bytes.NewReader(srcFile)); err != nil {
		return err
	}

	return nil
}

func showFilePreview(username, password, server, path string, app *tview.Application, rootFlex *tview.Flex, list *tview.List) {
	content := getFileContent(username, password, server, path)
	if content == "" {
		return
	}

	previewPane := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(content)
	previewPane.SetBorder(true)
	previewPane.SetTitle(filepath.Base(path))

	rootFlex.AddItem(previewPane, 0, 1, true)
	app.SetFocus(previewPane)

	previewPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q') {
			rootFlex.RemoveItem(previewPane)
			app.SetFocus(list)
			return nil
		}
		return event
	})
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

func showModal(app *tview.Application, message string, rootFlex *tview.Flex, form *tview.Form, list *tview.List) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(rootFlex, true)
			if list != nil {
				app.SetFocus(list)
			} else {
				app.SetFocus(form)
			}
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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
		return true, nil
	}
	return false, nil
}

func connectAndListFiles(username, password, server, dir string) ([]string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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
