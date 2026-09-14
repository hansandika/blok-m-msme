package store

import (
	"testing"
)

func TestParseBBox(t *testing.T) {
	b, err := ParseBBox("106.79,-6.256,106.81,-6.234")
	if err != nil {
		t.Fatal(err)
	}
	if b.West != 106.79 || b.South != -6.256 || b.East != 106.81 || b.North != -6.234 {
		t.Fatalf("unexpected bbox %#v", b)
	}
	if _, err := ParseBBox("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseBool(t *testing.T) {
	v, err := ParseBool("true")
	if err != nil || v == nil || !*v {
		t.Fatalf("true: %v %v", v, err)
	}
	v, err = ParseBool("0")
	if err != nil || v == nil || *v {
		t.Fatalf("0: %v %v", v, err)
	}
	v, err = ParseBool("")
	if err != nil || v != nil {
		t.Fatalf("empty: %v %v", v, err)
	}
}
