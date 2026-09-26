package tests

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/diag"
	"github.com/siper92/akha/lang/token"
)

type ErrCase struct {
	Name string
	Src  string
	Code string
	Pos  token.Pos
}

type AstCase struct {
	Name string
	Src  string
	Want string
}

func AsDiag(t *testing.T, err error) *diag.Error {
	t.Helper()
	var de *diag.Error
	if !errors.As(err, &de) {
		t.Fatalf("expected *diag.Error, got %v", err)
	}

	return de
}

func CheckErr(t *testing.T, err error, kind error, c ErrCase) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %s at %v, got none", c.Code, c.Pos)
	}

	if !errors.Is(err, kind) {
		t.Fatalf("expected %v, got %v", kind, err)
	}
	de := AsDiag(t, err)

	if de.Code != c.Code {
		t.Fatalf("code\n got: %s\nwant: %s\n err: %v", de.Code, c.Code, err)
	}

	if de.Pos != c.Pos {
		t.Fatalf("pos\n got: %v\nwant: %v\n err: %v", de.Pos, c.Pos, err)
	}
}
