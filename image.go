package giopdf

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"

	"github.com/andybalholm/giopdf/pdf"
)

func decodeImage(img pdf.Value) (image.Image, error) {
	if img.HasFilter("DCTDecode") {
		// It's a JPEG image.
		return jpeg.Decode(img.EncodedReader("DCTDecode"))
	}

	data, err := io.ReadAll(img.Reader())
	if err != nil {
		return nil, err
	}
	cs := img.Key("ColorSpace").Name()
	bits := img.Key("BitsPerComponent").Int()
	switch {
	case cs == "DeviceGray" && bits == 1:
		result := &bitmapImage{
			Width:      img.Key("Width").Int(),
			Height:     img.Key("Height").Int(),
			Data:       data,
			Foreground: color.White,
			Background: color.Black,
		}
		switch img.Key("Decode").String() {
		case "[0, 1]", "<nil>":
			// Default; leave values unchanged.
		case "[1, 0]":
			result.Background = color.White
			result.Foreground = color.Black
		default:
			return nil, fmt.Errorf("unsupported Decode array: %v", img.Key("Decode"))
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported image (ColorSpace: %v, BitsPerComponent: %d)", img.Key("ColorSpace"), bits)
}

type bitmapImage struct {
	Width      int
	Height     int
	Data       []byte
	Foreground color.Color
	Background color.Color
}

func (bi *bitmapImage) ColorModel() color.Model {
	return color.GrayModel
}

func (bi *bitmapImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, bi.Width, bi.Height)
}

func (bi *bitmapImage) At(x, y int) color.Color {
	stride := bi.Width / 8
	if bi.Width%8 != 0 {
		stride += 1
	}
	b := bi.Data[stride*y+x/8]
	b &= 1 << (7 - x%8)

	if b == 0 {
		return bi.Background
	}
	return bi.Foreground
}
