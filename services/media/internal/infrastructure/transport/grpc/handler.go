package grpc

import (
	"io"

	mediapb "github.com/mihnpro/Merch_shop/services/media/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

func (g *GRPCServer) UploadPhoto(stream mediapb.MediaService_UploadPhotoServer) error {
	first, err := stream.Recv()
	if err != nil {
		return NewGRPCError(domain.ErrEmptyRequest)
	}
	meta := first.GetMetadata()
	if meta == nil {
		return NewGRPCError(domain.ErrEmptyRequest)
	}

	pr, pw := io.Pipe()
	go func() {
		for {
			msg, recvErr := stream.Recv()
			if recvErr == io.EOF {
				_ = pw.Close()
				return
			}
			if recvErr != nil {
				_ = pw.CloseWithError(recvErr)
				return
			}
			if data := msg.GetData(); len(data) > 0 {
				if _, wErr := pw.Write(data); wErr != nil {
					return
				}
			}
		}
	}()

	res, err := g.uploader.Upload(stream.Context(), dto.UploadInput{
		Filename:    meta.GetFilename(),
		ContentType: meta.GetContentType(),
		Size:        -1,
		Body:        pr,
	})
	if err != nil {
		_ = pr.CloseWithError(err)
		return NewGRPCError(err)
	}

	return stream.SendAndClose(&mediapb.PhotoKey{PhotoKey: res.PhotoKey})
}
