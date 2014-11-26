package opts

import (
	"testing"
)

func TestList(t *testing.T) {
	l := List{}
	if l.String() != "[]" {
		t.Errorf("expected empty list")
	}

	l.Set("foo")
	if len(l.Get()) != 1 {
		t.Errorf("only one argument added, but %d found", len(l.Get()))
	}
	if l.String() != "[foo]" {
		t.Errorf("expected single item list")
	}

	l.Set("bar")
	if len(l.Get()) != 2 {
		t.Errorf("only two arguments added, but %d found", len(l.Get()))
	}
	if l.String() != "[foo, bar]" {
		t.Errorf("expected two item list")
	}
}
