package infrastructure

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"scpez/entities"
	"strings"
)

// SSHClientImpl implements the SSHClient interface for real SSH operations.
type SSHClientImpl struct{}

// NewSSHClientImpl creates a new instance of SSHClientImpl.
func NewSSHClientImpl() *SSHClientImpl {
	return &SSHClientImpl{}
}

// ListFiles implements the SSHClient interface to list files on a remote server.
func (client *SSHClientImpl) ListFiles(user entities.User, server entities.Server, path string) ([]entities.FileItem, error) {
	config := &ssh.ClientConfig{
		User: user.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(user.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, you should use a more secure method.
	}
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Address, server.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("ls -l " + path); err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	return parseFileList(b.String()), nil
}

// TransferFile implements the SSHClient interface to transfer a file from a remote server.
func (client *SSHClientImpl) TransferFile(user entities.User, server entities.Server, filePath string, destinationPath string) error {
	config := &ssh.ClientConfig{
		User: user.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(user.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, you should use a more secure method.
	}
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Address, server.Port), config)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Set up the command to pipe the file content directly to a local file
	cmd := fmt.Sprintf("cat %s", filepath.Clean(filePath))
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	// Create the local file and write the output
	localFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	if _, err = localFile.Write(b.Bytes()); err != nil {
		return fmt.Errorf("failed to write to local file: %v", err)
	}

	return nil
}

// parseFileList parses the output of an 'ls -l' command into a slice of FileItem.
func parseFileList(output string) []entities.FileItem {
	lines := strings.Split(output, "\n")
	var files []entities.FileItem
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Simplified parsing: this would need to be more robust in a real application
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}
		name := parts[8]
		isDir := (parts[0][0] == 'd')
		files = append(files, entities.FileItem{Name: name, IsDirectory: isDir})
	}
	return files
}
