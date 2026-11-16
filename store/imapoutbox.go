package store

import (
	"fmt"
	"time"

	"github.com/mjl-/bstore"
)

// IMAPMutationEvent is a durable, at-least-once handoff record for host-owned
// mailbox persistence. It is written in the same bstore transaction as the
// IMAP mutation, so a committed mutation cannot lose its handoff record.
type IMAPMutationEvent struct {
	ID          int64
	EventID     string `bstore:"nonzero,unique"`
	Account     string `bstore:"nonzero"`
	Operation   string `bstore:"nonzero"`
	MailboxID   int64  `bstore:"nonzero"`
	MessageID   int64  `bstore:"nonzero"`
	UID         UID    `bstore:"nonzero"`
	ModSeq      ModSeq `bstore:"nonzero"`
	Size        int64
	Received    time.Time
	Flags       Flags
	Keywords    []string
	MessagePath string `bstore:"nonzero"`
	MailboxName string `bstore:"nonzero"`
	MsgPrefix   []byte
}

// NewIMAPAppendEvent builds the handoff payload for a committed APPEND.
func NewIMAPAppendEvent(account string, mailbox Mailbox, message Message, messagePath string) IMAPMutationEvent {
	return IMAPMutationEvent{
		EventID: fmt.Sprintf("imap-append:%s:%d:%d", account, message.ID, message.UID),
		Account: account, Operation: "APPEND", MailboxID: mailbox.ID, MessageID: message.ID,
		UID: message.UID, ModSeq: message.ModSeq, Size: message.Size, Received: message.Received,
		Flags: message.Flags, Keywords: append([]string(nil), message.Keywords...), MessagePath: messagePath,
		MailboxName: mailbox.Name, MsgPrefix: append([]byte(nil), message.MsgPrefix...),
	}
}

// QueueIMAPAppend records a committed-message handoff while the caller still
// owns the enclosing account transaction.
func QueueIMAPAppend(tx *bstore.Tx, account string, mailbox Mailbox, message Message, messagePath string) error {
	if tx == nil {
		return fmt.Errorf("bstore transaction is required")
	}
	if account == "" || mailbox.ID == 0 || message.ID == 0 || message.UID == 0 {
		return fmt.Errorf("account, mailbox, message, and UID are required")
	}
	event := IMAPMutationEvent{
		EventID: fmt.Sprintf("imap-append:%s:%d:%d", account, message.ID, message.UID),
		Account: account, Operation: "APPEND", MailboxID: mailbox.ID, MessageID: message.ID,
		UID: message.UID, ModSeq: message.ModSeq, Size: message.Size, Received: message.Received,
		Flags: message.Flags, Keywords: append([]string(nil), message.Keywords...), MessagePath: messagePath, MailboxName: mailbox.Name, MsgPrefix: append([]byte(nil), message.MsgPrefix...),
	}
	return tx.Insert(&event)
}
