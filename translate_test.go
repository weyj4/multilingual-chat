package main

import "testing"

func TestTranslate(t *testing.T) {
	cases := []struct {
		in, want, inlang, outlang string
	}{
		{"Hello", "Hola ", "en", "es"},
	}
	for _, c := range cases {
		got := Translate(c.inlang, c.outlang, c.in)
		if got != c.want {
			t.Errorf("Translate(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
