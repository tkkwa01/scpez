package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// フォームを作成
	form := tview.NewForm().
		AddInputField("Username", "k", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("Connect", func() {
			// SSH接続ロジック
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Enter SSH Credentials").SetTitleAlign(tview.AlignLeft)

	// メインループを実行
	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
