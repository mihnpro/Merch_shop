package rest

import (
	"io"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

const maxRequestSize = 6 * 1024 * 1024

type uploadResponse struct {
	PhotoKey string `json:"photo_key"`
}

func (s *Server) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	mr, err := r.MultipartReader()
	if err != nil {
		writeError(w, domain.ErrEmptyRequest)
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, domain.ErrEmptyRequest)
			return
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}

		res, err := s.uploader.Upload(r.Context(), dto.UploadInput{
			Filename:    part.FileName(),
			ContentType: part.Header.Get("Content-Type"),
			Size:        -1,
			Body:        part,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, uploadResponse{PhotoKey: res.PhotoKey})
		return
	}

	writeError(w, domain.ErrEmptyRequest)
}
