package dto

import "io"

type UploadInput struct {
	Filename    string
	ContentType string
	Size        int64
	Body        io.Reader
}

type UploadResult struct {
	PhotoKey string
}

type Identity struct {
	UserID string
	Role   string
}

func (i Identity) IsAdmin() bool { return i.Role == "admin" }
