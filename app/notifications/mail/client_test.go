package mail

import "testing"

func TestNewClient_MissingKey(t *testing.T) {
	t.Setenv("RESEND_KEY", "")
	t.Setenv("MAIL_FROM", "noreply@example.com")
	if _, err := NewClient(); err == nil {
		t.Error("expected error when RESEND_KEY is not set")
	}
}

func TestNewClient_MissingFrom(t *testing.T) {
	t.Setenv("RESEND_KEY", "re_test_key")
	t.Setenv("MAIL_FROM", "")
	if _, err := NewClient(); err == nil {
		t.Error("expected error when MAIL_FROM is not set")
	}
}

func TestNewClient_Success(t *testing.T) {
	t.Setenv("RESEND_KEY", "re_test_key")
	t.Setenv("MAIL_FROM", "noreply@example.com")
	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.from != "noreply@example.com" {
		t.Errorf("expected from noreply@example.com, got %s", c.from)
	}
}
