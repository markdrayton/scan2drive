package translate

import (
	"context"
	"io"

	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/stapelberg/scan2drive/internal/jobqueue"
	"github.com/stapelberg/scan2drive/internal/user"
)

func TranslatePDF(ctx context.Context, u *user.Account, j *jobqueue.Job, rd io.Reader) ([]byte, error) {
	b, err := io.ReadAll(rd)
	req := &translatepb.TranslateDocumentRequest{
		Parent:             "projects/scan2drive-375315/locations/global",
		SourceLanguageCode: j.Language,
		TargetLanguageCode: u.Language,
		DocumentInputConfig: &translatepb.DocumentInputConfig{
			Source: &translatepb.DocumentInputConfig_Content{
				Content: b,
			},
			MimeType: "application/pdf",
		},
	}
	resp, err := u.Translate.TranslateDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.DocumentTranslation.ByteStreamOutputs[0], nil
}
