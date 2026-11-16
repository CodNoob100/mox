package store

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMsgreader(t *testing.T) {
	_, err := io.Copy(io.Discard, &MsgReader{prefix: []byte("hello"), path: "bogus.txt", size: int64(len("hello"))})
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected error for non-existing file, got %s", err)
	}

	if err := os.WriteFile("emptyfile_test.txt", []byte{}, 0660); err != nil {
		t.Fatalf("writing emptyfile_test.txt: %s", err)
	}
	defer os.Remove("emptyfile_test.txt")
	mr := &MsgReader{prefix: []byte("hello"), path: "emptyfile_test.txt", size: int64(len("hello"))}
	defer mr.Close()
	if buf, err := io.ReadAll(mr); err != nil {
		t.Fatalf("readall: %s", err)
	} else if string(buf) != "hello" {
		t.Fatalf("got %q, expected %q", buf, "hello")
	}

	if err := os.WriteFile("msgreader_test.txt", []byte(" world"), 0660); err != nil {
		t.Fatalf("writing msgreader_test.txt: %s", err)
	}
	defer os.Remove("msgreader_test.txt")
	mr = &MsgReader{prefix: []byte("hello"), path: "msgreader_test.txt", size: int64(len("hello world"))}
	defer mr.Close()
	if buf, err := io.ReadAll(mr); err != nil {
		t.Fatalf("readall: %s", err)
	} else if string(buf) != "hello world" {
		t.Fatalf("got %q, expected %q", buf, "hello world")
	}

	mr.Reset()
	buf := make([]byte, 32)
	if n, err := mr.ReadAt(buf, 1); err != nil && err != io.EOF {
		t.Fatalf("readat: n %d, s %q, err %s", n, buf[:n], err)
	} else if n != len("ello world") || string(buf[:n]) != "ello world" {
		t.Fatalf("readat: got %d bytes (%q), expected %d (%q)", n, buf, int64(len("ello world")), "ello world")
	}

	// Read with 1 byte at a time to exercise the offset/buffer-length calculations.
	buf = make([]byte, 1)
	var result []byte
	mr = &MsgReader{prefix: []byte("hello"), path: "msgreader_test.txt", size: int64(len("hello world"))}
	for {
		n, err := mr.Read(buf)
		if n > 0 {
			result = append(result, buf...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read: %s", err)
		}
	}
	if string(result) != "hello world" {
		t.Fatalf("got %q, want %q", result, "hello world")
	}

	if err := mr.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	f, err := os.Open("msgreader_test.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	mr = FileMsgReader([]byte("hello"), f)

	if mr.Size() != int64(len("hello world")) {
		t.Fatalf("size, got %d, expect %d", mr.Size(), len("hello world"))
	}

	if err := mr.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

type testSourceCloser struct{ closed bool }

func (c *testSourceCloser) Close() error { c.closed = true; return nil }

func TestSourceMsgReader(t *testing.T) {
	data := []byte("Subject: source\r\n\r\nbody")
	closer := &testSourceCloser{}
	mr := SourceMsgReader([]byte("prefix"), bytes.NewReader(data), int64(len(data)), closer)
	got, err := io.ReadAll(mr)
	if err != nil || string(got) != "prefixSubject: source\r\n\r\nbody" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if err := mr.Close(); err != nil || !closer.closed {
		t.Fatalf("close err=%v closed=%v", err, closer.closed)
	}
}

func TestMessageReaderHookFailsClosed(t *testing.T) {
	previous := MessageSourceHook
	defer func() { MessageSourceHook = previous }()
	MessageSourceHook = func(string, Message) (io.ReaderAt, int64, io.Closer, error) {
		return nil, 0, nil, errors.New("object unavailable")
	}
	a := &Account{Name: "alice"}
	mr := a.MessageReader(Message{ID: 1, MsgPrefix: []byte("prefix"), Size: 6})
	if _, err := io.ReadAll(mr); err == nil || !strings.Contains(err.Error(), "object unavailable") {
		t.Fatalf("read err=%v", err)
	}
}
