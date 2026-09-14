package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// applyWatermark menambahkan teks kredit di pojok kiri-bawah gambar.
// Mendukung JPEG/PNG/GIF; gambar di-decode & di-encode ulang.
// Mengembalikan bytes gambar yang sudah diberi watermark.
func applyWatermark(data []byte, contentType, credit string) ([]byte, error) {
	text := watermarkText(credit)
	if text == "" {
		return data, nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("watermark: decode: %w", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Teks watermark semi-transparan di kiri-bawah dengan latar gelap.
	drawWatermarkText(rgba, text)

	encoded, err := encodeImage(rgba, contentType)
	if err != nil {
		return nil, fmt.Errorf("watermark: encode: %w", err)
	}
	return encoded, nil
}

// drawWatermarkText menggambar teks kredit + latar semi-transparan.
func drawWatermarkText(img *image.RGBA, text string) {
	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.NRGBA{255, 255, 255, 230}),
		Face: face,
	}

	// Ukuran teks.
	textWidth := font.MeasureString(face, text).Ceil()
	pad := 4
	bounds := img.Bounds()
	x := bounds.Min.X + pad
	y := bounds.Max.Y - pad - face.Metrics().Descent.Ceil()

	// Latar semi-transparan hitam.
	bg := image.NewUniform(color.NRGBA{0, 0, 0, 160})
	bgRect := image.Rect(
		x-pad,
		y-pad-face.Metrics().Ascent.Ceil(),
		x+textWidth+pad,
		y+pad+face.Metrics().Descent.Ceil(),
	)
	draw.Draw(img, bgRect, bg, image.Point{}, draw.Over)

	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(text)
}

// encodeImage meng-encode image.RGBA ke format sesuai content type.
func encodeImage(img *image.RGBA, contentType string) ([]byte, error) {
	var buf bytes.Buffer
	switch contentType {
	case "image/png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	case "image/gif":
		err := gif.Encode(&buf, img, nil)
		return buf.Bytes(), err
	default: // jpeg & lainnya
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		return buf.Bytes(), err
	}
}
