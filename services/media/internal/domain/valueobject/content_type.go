package valueobject

import (
	"strings"

	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

type ContentType struct {
	mime string
	ext  string
}

var allowed = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

func NewContentType(raw string) (ContentType, error) {
	mime := strings.ToLower(strings.TrimSpace(raw))
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	ext, ok := allowed[mime]
	if !ok {
		return ContentType{}, domain.ErrInvalidContentType
	}
	return ContentType{mime: mime, ext: ext}, nil
}

func (c ContentType) String() string { return c.mime }

func (c ContentType) Ext() string { return c.ext }
