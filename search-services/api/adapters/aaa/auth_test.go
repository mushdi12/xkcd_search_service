package aaa

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func newTestAAA(t *testing.T) AAA {
	t.Helper()

	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "password")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	a, err := New(time.Minute, logger)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return a
}

func TestAAA_Login_SuccessAndVerify(t *testing.T) {
	a := newTestAAA(t)

	token, err := a.Login("admin", "password")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	if err := a.Verify(token); err != nil {
		t.Fatalf("Verify returned error for valid token: %v", err)
	}
}

func TestAAA_Login_InvalidCredentials(t *testing.T) {
	a := newTestAAA(t)

	if _, err := a.Login("admin", "wrong"); err == nil {
		t.Fatalf("expected error for wrong password")
	}
	if _, err := a.Login("unknown", "password"); err == nil {
		t.Fatalf("expected error for unknown user")
	}
}

func TestAAA_Verify_InvalidToken(t *testing.T) {
	a := newTestAAA(t)

	if err := a.Verify("not_a_jwt"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
