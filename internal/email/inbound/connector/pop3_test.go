package connector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/knadh/go-pop3"
	"github.com/stretchr/testify/require"
)

func TestPOP3FetcherFetchesMessages(t *testing.T) {
	conn := &fakePOP3Conn{
		uidl: []pop3.MessageID{
			{ID: 1, UID: "uid-1", Size: 123},
			{ID: 2, UID: "uid-2", Size: 456},
		},
		raw: map[int][]byte{
			1: []byte("first"),
			2: []byte("second"),
		},
	}
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	h := &recordingHandler{}
	f := NewPOP3Fetcher(
		WithPOP3Clock(func() time.Time { return now }),
		withPOP3ConnFactory(func(Account) (pop3Connection, error) { return conn, nil }),
	)

	acc := Account{ID: 7, Type: "pop3s", Host: "mail.example", Port: 995, Username: "agent", Password: []byte("secret")}
	require.NoError(t, f.Fetch(context.Background(), acc, h))

	require.Equal(t, 2, len(h.messages))
	require.Equal(t, []int{1, 2}, conn.deleted)
	require.Equal(t, 1, conn.quitCalls)
	require.Equal(t, "uid-1", h.messages[0].UID)
	require.Equal(t, now, h.messages[0].ReceivedAt)
	require.Equal(t, []byte("first"), h.messages[0].Raw)
}

func TestPOP3FetcherStopsOnHandlerError(t *testing.T) {
	conn := &fakePOP3Conn{
		uidl: []pop3.MessageID{{ID: 1, UID: "uid-1"}, {ID: 2, UID: "uid-2"}},
		raw:  map[int][]byte{1: []byte("first"), 2: []byte("second")},
	}
	h := &recordingHandler{failUID: "uid-2"}
	f := NewPOP3Fetcher(
		WithPOP3Clock(func() time.Time { return time.Unix(0, 0) }),
		withPOP3ConnFactory(func(Account) (pop3Connection, error) { return conn, nil }),
	)

	acc := Account{ID: 7, Type: "pop3", Host: "mail.example", Username: "agent", Password: []byte("secret")}
	err := f.Fetch(context.Background(), acc, h)
	require.Error(t, err)
	require.Equal(t, []int{1}, conn.deleted)
	require.Equal(t, 1, len(h.messages))
}

func TestPOP3FetcherReturnsAuthError(t *testing.T) {
	conn := &fakePOP3Conn{authErr: errors.New("bad creds")}
	f := NewPOP3Fetcher(withPOP3ConnFactory(func(Account) (pop3Connection, error) { return conn, nil }))
	h := &recordingHandler{}
	acc := Account{ID: 7, Type: "pop3", Host: "mail.example", Username: "agent", Password: []byte("secret")}
	err := f.Fetch(context.Background(), acc, h)
	require.ErrorContains(t, err, "pop3 auth")
	require.Empty(t, h.messages)
}

type recordingHandler struct {
	messages []*FetchedMessage
	failUID  string
}

func (h *recordingHandler) Handle(_ context.Context, msg *FetchedMessage) error {
	if h.failUID == msg.UID {
		return fmt.Errorf("fail %s", msg.UID)
	}
	h.messages = append(h.messages, msg)
	return nil
}

type fakePOP3Conn struct {
	uidl      []pop3.MessageID
	raw       map[int][]byte
	deleted   []int
	quitCalls int

	authErr error
	uidlErr error
	retrErr map[int]error
	deleErr error
	quitErr error
}

func (f *fakePOP3Conn) Auth(_, _ string) error {
	return f.authErr
}

func (f *fakePOP3Conn) Quit() error {
	f.quitCalls++
	return f.quitErr
}

func (f *fakePOP3Conn) Uidl(_ int) ([]pop3.MessageID, error) {
	if f.uidlErr != nil {
		return nil, f.uidlErr
	}
	out := make([]pop3.MessageID, len(f.uidl))
	copy(out, f.uidl)
	return out, nil
}

func (f *fakePOP3Conn) RetrRaw(id int) (*bytes.Buffer, error) {
	if err, ok := f.retrErr[id]; ok {
		return nil, err
	}
	payload, ok := f.raw[id]
	if !ok {
		return nil, fmt.Errorf("unknown message %d", id)
	}
	return bytes.NewBuffer(payload), nil
}

func (f *fakePOP3Conn) Dele(ids ...int) error {
	if f.deleErr != nil {
		return f.deleErr
	}
	f.deleted = append(f.deleted, ids...)
	return nil
}
