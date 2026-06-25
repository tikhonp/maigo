package maigo

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
)

// Attachment is a binary payload (file or image) encoded for transport over the
// Medsenger API. It is used both for message attachments and for files attached
// to medical records.
type Attachment struct {
	Name   string `json:"name"`   // File name shown to users.
	Type   string `json:"type"`   // MIME type, e.g. "image/png".
	Base64 string `json:"base64"` // Base64-encoded file contents.
}

// MessageAttachment was previously an empty, unusable struct. It is now an alias
// of Attachment so existing references keep compiling.
//
// Deprecated: use [Attachment] instead.
type MessageAttachment = Attachment

// PrepareBinary builds an Attachment from in-memory data, detecting the MIME
// type from the content. It mirrors the Python SDK's prepare_binary helper.
func PrepareBinary(name string, data []byte) Attachment {
	return Attachment{
		Name:   name,
		Type:   http.DetectContentType(data),
		Base64: base64.StdEncoding.EncodeToString(data),
	}
}

// PrepareFile reads a file from disk and builds an Attachment from it, using the
// base name of the path as the attachment name. It mirrors the Python SDK's
// prepare_file helper.
func PrepareFile(path string) (Attachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Attachment{}, err
	}
	return PrepareBinary(filepath.Base(path), data), nil
}
