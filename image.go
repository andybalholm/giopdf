package giopdf

import (
	"fmt"
	"image"
	"image/color"

	"github.com/benoitkugler/pdf/model"
)

func decodeImage(img *model.XObjectImage) (image.Image, error) {
	data, err := img.Stream.Decode()
	if err != nil {
		return nil, err
	}
	switch {
	case img.ColorSpace == model.ColorSpaceName("DeviceGray") && img.BitsPerComponent == 1:
		result := &bitmapImage{
			Width:      img.Width,
			Height:     img.Height,
			Data:       data,
			Foreground: color.White,
			Background: color.Black,
		}
		if len(img.Decode) > 0 {
			switch img.Decode[0] {
			case [2]float32{0, 1}:
				// Default; leave values unchanged.
			case [2]float32{1, 0}:
				result.Background = color.White
				result.Foreground = color.Black
			default:
				return nil, fmt.Errorf("unsupported Decode array: %v", img.Decode)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported image (ColorSpace: %s, BitsPerComponent: %d)", img.ColorSpace, img.BitsPerComponent)
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
