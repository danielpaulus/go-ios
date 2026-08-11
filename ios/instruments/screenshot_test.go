package instruments

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// makePNG builds a tiny valid PNG for the conversion tests.
func makePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 40), G: uint8(y * 40), B: 128, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return b.Bytes()
}

// TestPngToJPEG asserts the shared PNG->JPEG conversion step (used by both the
// built-in MJPEG server and external consumers) produces a decodable JPEG.
func TestPngToJPEG(t *testing.T) {
	pngBytes := makePNG(t)

	jpg, err := pngToJPEG(pngBytes, 80)
	if err != nil {
		t.Fatalf("pngToJPEG: %v", err)
	}
	if len(jpg) == 0 {
		t.Fatal("pngToJPEG returned no bytes")
	}
	img, err := jpeg.Decode(bytes.NewReader(jpg))
	if err != nil {
		t.Fatalf("output is not valid jpeg: %v", err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("unexpected jpeg bounds: %v", img.Bounds())
	}
}

// TestPngToJPEGRejectsGarbage asserts a non-PNG input errors instead of panicking.
func TestPngToJPEGRejectsGarbage(t *testing.T) {
	if _, err := pngToJPEG([]byte("not a png"), 80); err == nil {
		t.Fatal("expected error decoding non-png input")
	}
}
