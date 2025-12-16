package filters

import (
	"bytes"
	"context"
	"log"
	"net/mail"
	"path"
	"strings"
)

// DispatchRule describes how to override routing based on sender patterns.
type DispatchRule struct {
	Match      string `yaml:"match"`
	QueueName  string `yaml:"queue"`
	QueueID    int    `yaml:"queue_id"`
	PriorityID int    `yaml:"priority_id"`
}

// DispatchRuleProvider supplies per-account rule lists.
type DispatchRuleProvider interface {
	RulesFor(accountID int) []DispatchRule
}

// DispatchFromMapFilter maps sender addresses to routing overrides for FROM dispatch accounts.
type DispatchFromMapFilter struct {
	provider DispatchRuleProvider
	logger   *log.Logger
}

// NewDispatchFromMapFilter constructs the filter.
func NewDispatchFromMapFilter(provider DispatchRuleProvider, logger *log.Logger) *DispatchFromMapFilter {
	return &DispatchFromMapFilter{provider: provider, logger: logger}
}

// ID implements Filter.
func (f *DispatchFromMapFilter) ID() string { return "dispatch_from_map" }

// Apply inspects the message sender and applies the first matching rule.
func (f *DispatchFromMapFilter) Apply(ctx context.Context, m *MessageContext) error {
	if f == nil || f.provider == nil || m == nil || m.Message == nil || len(m.Message.Raw) == 0 {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(m.Account.DispatchingMode), "from") {
		return nil
	}
	from := f.senderAddress(m.Message.Raw)
	if from == "" {
		return nil
	}
	rules := f.provider.RulesFor(m.Account.ID)
	for _, rule := range rules {
		if rule.matches(from) {
			f.applyRule(m, rule, from)
			break
		}
	}
	return nil
}

func (f *DispatchFromMapFilter) senderAddress(raw []byte) string {
	reader, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		f.logf("dispatch_from_map: parse message failed: %v", err)
		return ""
	}
	value := strings.TrimSpace(reader.Header.Get("From"))
	if value == "" {
		return ""
	}
	if addrs, err := mail.ParseAddressList(value); err == nil && len(addrs) > 0 {
		return strings.ToLower(strings.TrimSpace(addrs[0].Address))
	}
	if addr, err := mail.ParseAddress(value); err == nil {
		return strings.ToLower(strings.TrimSpace(addr.Address))
	}
	return strings.ToLower(value)
}

func (f *DispatchFromMapFilter) applyRule(m *MessageContext, rule DispatchRule, sender string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]any)
	}
	if rule.QueueID > 0 {
		m.Annotations[AnnotationQueueIDOverride] = rule.QueueID
	}
	if strings.TrimSpace(rule.QueueName) != "" {
		m.Annotations[AnnotationQueueNameOverride] = strings.TrimSpace(rule.QueueName)
	}
	if rule.PriorityID > 0 {
		m.Annotations[AnnotationPriorityIDOverride] = rule.PriorityID
	}
	f.logf("dispatch_from_map: account %d matched %s for %s", m.Account.ID, rule.Match, sender)
}

func (f *DispatchFromMapFilter) logf(format string, args ...any) {
	if f == nil || f.logger == nil {
		return
	}
	f.logger.Printf(format, args...)
}

func (r DispatchRule) matches(addr string) bool {
	pattern := strings.TrimSpace(strings.ToLower(r.Match))
	if pattern == "" || pattern == "*" {
		return true
	}
	addr = strings.ToLower(strings.TrimSpace(addr))
	match, err := path.Match(pattern, addr)
	if err != nil {
		return false
	}
	return match
}
