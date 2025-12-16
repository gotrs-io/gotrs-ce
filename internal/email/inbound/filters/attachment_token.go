package filters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"strings"

	gomail "github.com/emersion/go-message/mail"
)

// AttachmentTokenFilter inspects attachment metadata for ticket tokens.
type AttachmentTokenFilter struct {
	logger *log.Logger
}

// NewAttachmentTokenFilter builds the metadata-based follow-up detector.
func NewAttachmentTokenFilter(logger *log.Logger) *AttachmentTokenFilter {
	return &AttachmentTokenFilter{logger: logger}
}

// ID implements Filter.
func (f *AttachmentTokenFilter) ID() string { return "followup_attachment_token" }

// Apply scans attachment metadata for [Ticket#] markers.
func (f *AttachmentTokenFilter) Apply(ctx context.Context, m *MessageContext) error {
	if m == nil || m.Message == nil || len(m.Message.Raw) == 0 {
		return nil
	}
	if hasFollowUpAnnotation(m) {
		return nil
	}
	if token := f.detectToken(m.Message.Raw); token != "" {
		if m.Annotations == nil {
			m.Annotations = make(map[string]any)
		}
		m.Annotations[AnnotationFollowUpTicketNumber] = token
		f.logf("followup_attachment_token: detected ticket %s", token)
	}
	return nil
}

func (f *AttachmentTokenFilter) detectToken(raw []byte) string {
	reader, err := gomail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		f.logf("followup_attachment_token: parse failed: %v", err)
		return ""
	}
	for {
		part, perr := reader.NextPart()
		if errors.Is(perr, io.EOF) {
			break
		}
		if perr != nil {
			f.logf("followup_attachment_token: read part failed: %v", perr)
			break
		}
		attHeader, ok := part.Header.(*gomail.AttachmentHeader)
		if !ok {
			continue
		}
		for _, candidate := range f.attachmentCandidates(attHeader) {
			if candidate == "" {
				continue
			}
			if token := findTicketToken(candidate); token != "" {
				return token
			}
		}
	}
	return ""
}

func (f *AttachmentTokenFilter) attachmentCandidates(header *gomail.AttachmentHeader) []string {
	if header == nil {
		return nil
	}
	var candidates []string
	if name, err := header.Filename(); err == nil {
		candidates = append(candidates, name)
	}
	if disp, params, err := header.ContentDisposition(); err == nil {
		candidates = append(candidates, disp)
		for _, key := range []string{"filename", "name"} {
			if v := strings.TrimSpace(params[key]); v != "" {
				candidates = append(candidates, v)
			}
		}
	}
	if ctype, params, err := header.ContentType(); err == nil {
		candidates = append(candidates, ctype)
		if v := strings.TrimSpace(params["name"]); v != "" {
			candidates = append(candidates, v)
		}
	}
	for _, key := range []string{"Content-Description", "Content-ID"} {
		if v := strings.TrimSpace(header.Get(key)); v != "" {
			candidates = append(candidates, v)
		}
	}
	return candidates
}

func (f *AttachmentTokenFilter) logf(format string, args ...any) {
	if f == nil || f.logger == nil {
		return
	}
	f.logger.Printf(format, args...)
}

func hasFollowUpAnnotation(m *MessageContext) bool {
	if m == nil || m.Annotations == nil {
		return false
	}
	_, exists := m.Annotations[AnnotationFollowUpTicketNumber]
	return exists
}
