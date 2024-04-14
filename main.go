package main

import (
	"bytes"
	"fmt"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
)

// SSH接続とコマンド実行の関数
func connectAndListFiles(username, password, server string) (string, error) {
	// SSHクライアント設定
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 本番環境では適切なホストキーチェックを行うこと
	}

	// SSHクライアント接続
	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// SSHセッションの作成
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// コマンド実行
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("ls -l"); err != nil {
		return "", err
	}

	return b.String(), nil
}

func main() {
	app := tview.NewApplication()
	form := tview.NewForm()

	var username, password, server string
	form.AddInputField("Server", "", 20, nil, func(text string) { server = text }).
		AddInputField("Username", "", 20, nil, func(text string) { username = text }).
		AddPasswordField("Password", "", 20, '*', func(text string) { password = text }).
		AddButton("Connect", func() {
			output, err := connectAndListFiles(username, password, server)
			if err != nil {
				fmt.Println("Error connecting to server:", err)
				return
			}
			fmt.Println("Server response:", output)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("SSH Login").SetTitleAlign(tview.AlignLeft)

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
