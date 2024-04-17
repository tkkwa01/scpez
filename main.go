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

	// サーバ名のプリセット
	servers := []string{"sh.edu.kutc.kansai-u.ac.jp", "server2.example.com", "server3.example.net", "Other (Specify)"}
	var server string // 選択されたまたは入力されたサーバ名

	serverDropdown := tview.NewDropDown().
		SetLabel("Server: ").
		SetOptions(servers, nil)

	// サーバ名が「Other (Specify)」の場合に手動入力させる入力フィールド
	otherServerField := tview.NewInputField().SetLabel("Specify Server: ").SetFieldWidth(20)

	form.AddFormItem(serverDropdown).
		AddFormItem(otherServerField).
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Connect", func() {
			_, option := serverDropdown.GetCurrentOption()
			if option == "Other (Specify)" {
				server = otherServerField.GetText() // 入力フィールドからサーバ名を取得
			} else {
				server = option // ドロップダウンから選択されたサーバ名を使用
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
			switch event.Rune() {
			case 'l', 'L':
				selectedItem, _ := list.GetItemText(list.GetCurrentItem())
				selectedPath := filepath.Join(path, selectedItem)
				showFilePreview(username, password, server, selectedPath, app, rootFlex, list)
				return nil
			case 'b', 'B': // 戻る処理
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
