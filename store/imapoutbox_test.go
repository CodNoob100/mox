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
		return QueueIMAPAppend(tx, "account", Mailbox{ID: 7}, Message{ID: 9, UID: 3, ModSeq: 4, Size: 10, Received: received, Keywords: []string{"$label"}}, "/messages/9")
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
