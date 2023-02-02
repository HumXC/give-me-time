package ocr_test

import (
	"testing"

	"github.com/HumXC/give-me-time/ocr"
)

func TestText(t *testing.T) {
	want := "AilLWNNM1.230"
	text, err := ocr.Text("../test/AilLWnNM1.230.png")
	if err != nil {
		t.Fatal(err)
	}
	if want != text {
		t.Errorf("want: %s, got: %s", want, text)
	}
}

func BenchmarkText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ocr.Text("../test/AilLWnNM1.230.png")
		if err != nil {
			b.Fatal(err)
		}
	}
}
