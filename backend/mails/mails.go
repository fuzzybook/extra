// Service mails handles transactional email delivery.
package mails

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

const outboxDir = "mails/outbox"

type TransactionalMailParams struct {
	To       string            `json:"to"`
	Subject  string            `json:"subject"`
	TextBody string            `json:"text_body"`
	HTMLBody string            `json:"html_body"`
	Metadata map[string]string `json:"metadata"`
}

type TransactionalMail struct {
	ID        uuid.UUID         `json:"id"`
	To        string            `json:"to"`
	Subject   string            `json:"subject"`
	TextBody  string            `json:"text_body"`
	HTMLBody  string            `json:"html_body"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

// SendTransactional records a transactional email in the local outbox.
// TODO: replace the file-backed outbox with an SMTP sender.
//
//encore:api private method=POST path=/mails/transactional
func SendTransactional(ctx context.Context, p *TransactionalMailParams) (*TransactionalMail, error) {
	if strings.TrimSpace(p.To) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "recipient email is required"}
	}
	if strings.TrimSpace(p.Subject) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "email subject is required"}
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate mail id")
	}

	mail := TransactionalMail{
		ID:        id,
		To:        p.To,
		Subject:   p.Subject,
		TextBody:  p.TextBody,
		HTMLBody:  p.HTMLBody,
		Metadata:  p.Metadata,
		CreatedAt: time.Now().UTC(),
	}
	if mail.Metadata == nil {
		mail.Metadata = map[string]string{}
	}

	if err := saveToOutbox(ctx, &mail); err != nil {
		return nil, err
	}
	return &mail, nil
}

func saveToOutbox(ctx context.Context, mail *TransactionalMail) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(outboxDir, 0o755); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to create mail outbox")
	}

	payload, err := json.MarshalIndent(mail, "", "  ")
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to encode mail")
	}

	baseFilename := mail.CreatedAt.Format("20060102T150405Z") + "_" + mail.ID.String()
	jsonPath := filepath.Join(outboxDir, baseFilename+".json")
	if err := os.WriteFile(jsonPath, payload, 0o644); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to write mail to outbox")
	}

	if strings.TrimSpace(mail.HTMLBody) != "" {
		htmlPath := filepath.Join(outboxDir, baseFilename+".html")
		if err := os.WriteFile(htmlPath, []byte(mail.HTMLBody), 0o644); err != nil {
			return errs.WrapCode(err, errs.Internal, "failed to write mail html to outbox")
		}
	}
	return nil
}
