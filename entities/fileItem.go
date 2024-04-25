package entities

type FileItem struct {
	Name        string
	IsDirectory bool
	Size        int64 // Size in bytes for files
}
