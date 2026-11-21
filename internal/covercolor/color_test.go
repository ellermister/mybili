package covercolor

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDominantRedAndSkipsBlank(t *testing.T) {
	red := writePNG(t, color.RGBA{R: 220, G: 30, B: 40, A: 255})
	got, ok := Dominant(red)
	if !ok || got != "#dc1e28" {
		t.Fatalf("red = %s ok=%v", got, ok)
	}
	white := writePNG(t, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if _, ok := Dominant(white); ok {
		t.Fatal("white should be ignored")
	}
	if _, ok := Dominant(filepath.Join(t.TempDir(), "missing.png")); ok {
		t.Fatal("missing file should be ignored")
	}
}

func writePNG(t *testing.T, c color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, c)
		}
	}
	path := filepath.Join(t.TempDir(), "color.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	return path
}
