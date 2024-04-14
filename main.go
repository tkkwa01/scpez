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

	form := tview.NewForm()

	// ここで入力フィールドを設定
	form.AddInputField("Server", "", 20, nil, nil).
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Connect", func() {
			// ここで入力フィールドの値を取得
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			username := form.GetFormItem(1).(*tview.InputField).GetText()
			password := form.GetFormItem(2).(*tview.InputField).GetText()
			currentDir := "/home/" + username
			navigateDir(currentDir, username, password, server, app)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}

func navigateDir(path, username, password, server string, app *tview.Application) {
	files, err := connectAndListFiles(username, password, server, path)
	if err != nil {
		panic(err) // 適切なエラーハンドリングを行う
	}

	// ディレクトリ内容を表示するリスト
	list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	for _, file := range files {
		list.AddItem(file, "", 0, nil)
	}
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			selectedFile, _ := list.GetItemText(list.GetCurrentItem())
			newPath := filepath.Join(path, selectedFile)
			navigateDir(newPath, username, password, server, app)
			return nil
		}
		return event
	})

	// リストをアプリのルートに設定
	app.SetRoot(list, true)
}

func connectAndListFiles(username, password, server, dir string) ([]string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 安全でない例、実際には適切な設定が必要
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
