package main

import (
	"github.com/rivo/tview"
	"strconv"
)

func main() {
	app := tview.NewApplication()

	form := tview.NewForm().
		AddInputField("Server Address", "", 20, nil, nil).
		AddInputField("Port", "22", 5, nil, nil). // Default SSH port is 22
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("Connect", func() {
			serverAddress := form.GetFormItem(0).(*tview.InputField).GetText()
			portStr := form.GetFormItem(1).(*tview.InputField).GetText()
			username := form.GetFormItem(2).(*tview.InputField).GetText()
			password := form.GetFormItem(3).(*tview.InputField).GetText()

			// Convert port from string to int
			port, err := strconv.Atoi(portStr)
			if err != nil {
				port = 22 // Default to 22 if there's an error
			}

			// Here, you would typically pass these values to the Server and User entities
			// and proceed with connecting to the SSH server
			// For example:
			// connectToServer(serverAddress, port, username, password)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Enter SSH Details").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
