package main

import (
	"bytes"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
	"strings"
)

func main() {
	app := tview.NewApplication()

	// フォームとリストビューを作成
	form := tview.NewForm()
	list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)

	// フォームのフィールドを追加
	form.AddInputField("Server", "", 20, nil, nil)
	form.AddInputField("Username", "", 20, nil, nil)
	form.AddPasswordField("Password", "", 10, '*', nil)
	form.AddButton("Connect", func() {
		server := form.GetFormItemByLabel("Server").(*tview.InputField).GetText()
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

		// SSH接続とファイルリストの取得
		files, err := connectAndListFiles(username, password, server)
		if err != nil {
			panic(err) // エラーハンドリングは適切に行うべきです
		}

		// リストビューにファイルを追加
		list.Clear()
		for _, file := range files {
			if file != "" {
				list.AddItem(file, "", 0, nil)
			}
		}
		app.SetFocus(list)
	})
	form.AddButton("Quit", func() {
		app.Stop()
	})

	// レイアウトを設定
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 9, 1, true).
		AddItem(list, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

// SSH接続とコマンド実行の関数
func connectAndListFiles(username, password, server string) ([]string, error) {
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
	if err := session.Run("ls"); err != nil {
		return nil, err
	}

	return strings.Split(b.String(), "\n"), nil
}
