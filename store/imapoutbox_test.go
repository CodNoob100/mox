package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/mjl-/bstore"
)

func TestQueueIMAPAppendPersistsAtomicEvent(t *testing.T) {
	if err := QueueIMAPAppend(nil, "account", Mailbox{ID: 1}, Message{ID: 2, UID: 3}, "message"); err == nil {
		t.Fatal("nil transaction accepted")
	}
	db, err := bstore.Open(context.Background(), filepath.Join(t.TempDir(), "index.db"), &bstore.Options{Perm: 0660}, DBTypes...)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	received := time.Unix(123, 0).UTC()
	err = db.Write(context.Background(), func(tx *bstore.Tx) error {
		return QueueIMAPAppend(tx, "account", Mailbox{ID: 7, Name: "INBOX"}, Message{ID: 9, UID: 3, ModSeq: 4, Size: 10, Received: received, Keywords: []string{""}}, "/messages/9")
	})
	if err != nil {
		t.Fatal(err)
	}
	event, err := bstore.QueryDB[IMAPMutationEvent](context.Background(), db).FilterEqual("EventID", "imap-append:account:9:3").Get()
	if err != nil {
		t.Fatal(err)
	}
	if event.Operation != "APPEND" || event.MailboxID != 7 || event.MessageID != 9 || event.UID != 3 || event.MessagePath != "/messages/9" || len(event.Keywords) != 1 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestNewIMAPStoreEvent(t *testing.T) {
	event := NewIMAPStoreEvent("account", Mailbox{ID: 7, Name: "INBOX"}, Message{ID: 9, UID: 3, ModSeq: 4, Flags: Flags{Seen: true}, Keywords: []string{"$label"}})
	if event.Operation != "STORE" || event.EventID != "imap-store:account:9:4" || event.MailboxName != "INBOX" || !event.Flags.Seen || len(event.Keywords) != 1 || event.MessagePath != "" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestNewIMAPExpungeEvent(t *testing.T) {
	event := NewIMAPExpungeEvent("account", Mailbox{ID: 7, Name: "INBOX"}, Message{ID: 9, UID: 3}, 4)
	if event.Operation != "EXPUNGE" || event.EventID != "imap-expunge:account:9:4" || event.MailboxName != "INBOX" || event.UID != 3 || event.ModSeq != 4 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestNewIMAPCopyEvent(t *testing.T) {
	event := NewIMAPCopyEvent("account", Mailbox{ID: 7, Name: "Archive"}, 2, 3, Message{ID: 9, UID: 11, ModSeq: 4}, "/messages/9")
	if event.Operation != "COPY" || event.EventID != "imap-copy:account:9:11" || event.SourceMessageID != 2 || event.SourceUID != 3 || event.MailboxName != "Archive" || event.MessagePath != "/messages/9" {
		t.Fatalf("unexpected event: %+v", event)
	}
}
