package usecases

import (
	"scpez/entities"
	"scpez/interfaces"
)

type ServerInteractor struct {
	SSHClient interfaces.SSHClient
}

func NewServerInteractor(client interfaces.SSHClient) *ServerInteractor {
	return &ServerInteractor{
		SSHClient: client,
	}
}

func (si *ServerInteractor) ListFiles(user entities.User, server entities.Server, path string) ([]entities.FileItem, error) {
	return si.SSHClient.ListFiles(user, server, path)
}

func (si *ServerInteractor) TransferFile(user entities.User, server entities.Server, filePath string, destinationPath string) error {
	return si.SSHClient.TransferFile(user, server, filePath, destinationPath)
}
