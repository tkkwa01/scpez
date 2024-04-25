package interfaces

import "scpez/entities"

type SSHClient interface {
	ListFiles(user entities.User, server entities.Server, path string) ([]entities.FileItem, error)
	TransferFile(user entities.User, server entities.Server, filePath string, destinationPath string) error
}
