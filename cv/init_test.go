package cv_test

import (
	"image"
	"testing"

	"github.com/HumXC/give-me-time/cv"
)

func TestFind(t *testing.T) {
	want := image.Point{
		X: 103,
		Y: 174,
	}
	p, err := cv.Find("../test/big.png", "../test/small.png")
	if err != nil {
		t.Fatal(err)
		return
	}
	if !p.Eq(want) {
		t.Errorf("want: %v, got: %v", want, p)
		return
	}
}

func BenchmarkFind(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := cv.Find("../test/big.png", "../test/small.png")
		if err != nil {
			b.Fatal(err)
		}
	}
}
