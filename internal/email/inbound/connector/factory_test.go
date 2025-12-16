package connector

import (
	"context"
	"testing"
)

type noopFetcher struct{}

func (noopFetcher) Name() string { return "noop" }

func (noopFetcher) Fetch(ctx context.Context, account Account, handler Handler) error { return nil }

func TestFactoryReturnsRegisteredFetcher(t *testing.T) {
	fetcher := noopFetcher{}
	factory := NewFactory(WithFetcher(fetcher, "Pop3"))

	connFetcher, err := factory.FetcherFor(Account{Type: "POP3"})
	if err != nil {
		t.Fatalf("expected fetcher, got error %v", err)
	}
	if connFetcher.Name() != "noop" {
		t.Fatalf("unexpected fetcher %s", connFetcher.Name())
	}
}
