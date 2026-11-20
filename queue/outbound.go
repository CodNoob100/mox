package queue

import (
	"context"
	"os"

	"github.com/mjl-/mox/smtp"
)

// OutboundPolicyInput is the message metadata available to a host-owned
// policy before Mox opens a remote SMTP connection.
type OutboundPolicyInput struct {
	Sender      smtp.Path
	Recipient   smtp.Path
	MessagePath string
	MsgPrefix   []byte
	Size        int64
}

// OutboundPolicyDecision describes a policy rejection. A zero decision means
// the message may proceed. The hook may rewrite MessagePath in place.
type OutboundPolicyDecision struct {
	Code         int
	EnhancedCode string
	Message      string
	Permanent    bool
}

type OutboundPolicyHookFunc func(context.Context, OutboundPolicyInput) (OutboundPolicyDecision, error)

// OutboundPolicyHook is nil by default. Hosts may register a policy during
// startup; Mox calls it before remote delivery is attempted.
var OutboundPolicyHook OutboundPolicyHookFunc

// OpenMessage exposes the queue message body to a host policy without making
// the policy depend on Mox's internal message-reader implementation.
func OpenMessage(input OutboundPolicyInput) (*os.File, error) {
	return os.Open(input.MessagePath)
}
