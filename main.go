package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
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
		cleanText := strings.TrimSuffix(strings.TrimPrefix(mainText, "[red]"), "[white]")

		if _, ok := selectedFiles[cleanText]; ok {
			delete(selectedFiles, cleanText)
			list.SetItemText(index, cleanText, "")
		} else {
			selectedFiles[cleanText] = struct{}{}
			list.SetItemText(index, "[red]"+cleanText+"[white]", "")
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
				homeDir, err := os.UserHomeDir()
				if err != nil {
					showModal(app, "Failed to get home directory: "+err.Error(), rootFlex, form, list)
					return nil
				}

				scpEzDir := filepath.Join(homeDir, "SCP-EZ")
				if _, err := os.Stat(scpEzDir); os.IsNotExist(err) {
					os.Mkdir(scpEzDir, os.ModePerm)
				}

				currentDate := time.Now().Format("060102") // YYMMDD形式
				dateDir := filepath.Join(scpEzDir, currentDate)
				if _, err := os.Stat(dateDir); os.IsNotExist(err) {
					os.Mkdir(dateDir, os.ModePerm)
				}

				allSucceeded := true
				for filePath := range selectedFiles {
					remotePath := filepath.Join(path, filePath)
					if err := transferPath(username, password, server, remotePath, dateDir); err != nil {
						showModal(app, "Transfer Failed:"+err.Error(), rootFlex, form, list)
						allSucceeded = false
						break
					}
				}

				if allSucceeded {
					showModal(app, "Succeed Transfer!!", rootFlex, form, list)
				}
				return nil

			}
		}
		return event
	})

	rootFlex.AddItem(list, 0, 1, true)
	app.SetFocus(list)
}

func transferFile(username, password, server, remotePath, localBaseDir string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	session.Stderr = &stderr

	sanitizedPath := strings.Replace(remotePath, "*", "", -1)

	cmd := "cat \"" + sanitizedPath + "\""
	srcFile, err := session.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute command: '%s', stderr: %s, error: %v", cmd, stderr.String(), err)
	}

	localFilePath := filepath.Join(localBaseDir, filepath.Base(sanitizedPath))
	dstFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, bytes.NewReader(srcFile)); err != nil {
		return fmt.Errorf("failed to write to local file: %v", err)
	}

	return nil
}

func transferPath(username, password, server, remotePath string, localBaseDir string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer client.Close()

	isDir, err := isValidDirectory(username, password, server, remotePath)
	if err != nil {
		return fmt.Errorf("failed to determine if path is a directory: %v", err)
	}

	if isDir {
		files, err := connectAndListFiles(username, password, server, remotePath)
		if err != nil {
			return fmt.Errorf("failed to list files in directory: %v", err)
		}

		localDir := filepath.Join(localBaseDir, filepath.Base(remotePath))
		if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create local directory: %v", err)
		}

		for _, file := range files {
			remoteFilePath := filepath.Join(remotePath, file)
			if err := transferPath(username, password, server, remoteFilePath, localDir); err != nil {
				return err
			}
		}
	} else {
		if err := transferFile(username, password, server, remotePath, localBaseDir); err != nil {
			return fmt.Errorf("failed to transfer file: %v", err)
		}
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

	sanitizedPath := strings.Replace(path, "*", "", -1)

	if err := session.Run("cat \"" + sanitizedPath + "\""); err != nil {
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

	client, err := ssh.Dial("tcp", server+"22", config)
	if err != nil {
		return false, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to start session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	cmd := fmt.Sprintf("if [ -d \"%s\" ]; then echo true; else echo false; fi", path)
	if err := session.Run(cmd); err != nil {
		return false, fmt.Errorf("failed to execute command: %v, error: %v", cmd, err)
	}

	output := strings.TrimSpace(b.String())
	if output == "true" {
		return true, nil
	} else if output == "false" {
		return false, nil
	} else {
		return false, fmt.Errorf("unexpected output: %s", output)
	}
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
