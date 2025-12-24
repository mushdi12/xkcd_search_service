package db

import (
	"testing"
)

func TestStringArray_Scan_Nil(t *testing.T) {
	var a StringArray
	if err := a.Scan(nil); err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(a) != 0 {
		t.Fatalf("expected empty slice, got %v", a)
	}
}

func TestStringArray_Scan_EmptyString(t *testing.T) {
	var a StringArray
	if err := a.Scan("{}"); err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(a) != 0 {
		t.Fatalf("expected empty slice, got %v", a)
	}
}

func TestStringArray_Scan_WithValues(t *testing.T) {
	var a StringArray
	if err := a.Scan(`{"a","b","c"}`); err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(a) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(a))
	}
	if a[0] != "a" || a[1] != "b" || a[2] != "c" {
		t.Fatalf("unexpected values: %#v", a)
	}
}

func TestStringArray_Value_Empty(t *testing.T) {
	a := StringArray{}
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if v != "{}" {
		t.Fatalf("expected \"{}\", got %q", v)
	}
}

func TestStringArray_Value_WithValues(t *testing.T) {
	a := StringArray{"a", "b", "c"}
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if v != `{"a","b","c"}` {
		t.Fatalf("unexpected value: %q", v)
	}
}


