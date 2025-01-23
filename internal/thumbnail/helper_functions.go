package thumbnail

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jdeng/goheif"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"
)

type ImageType int

const (
	ImageTypeJPEG ImageType = iota
	ImageTypePNG
	ImageTypeHEIC
	ImageTypeNotSupported
)

func convertFileToJpeg(file multipart.File) (image.Image, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	imageType, errDet := detectImageTypeFromBytes(data)
	if errDet != nil {
		return nil, errDet
	}
	if imageType == ImageTypeNotSupported {
		return nil, errors.New("image type not supported")
	}

	var img image.Image
	switch imageType {
	case ImageTypeJPEG:
		img, err = jpeg.Decode(bytes.NewReader(data))
	case ImageTypePNG:
		img, err = png.Decode(bytes.NewReader(data))
	case ImageTypeHEIC:
		img, err = goheif.Decode(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unsupported format")
	}

	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return img, nil
}

func detectImageTypeFromBytes(data []byte) (ImageType, error) {
	headerBytes := data
	if len(data) > 512 {
		headerBytes = data[:512]
	}
	mimeType := http.DetectContentType(headerBytes)
	switch mimeType {
	case "image/jpeg":
		return ImageTypeJPEG, nil
	case "image/png":
		return ImageTypePNG, nil
	case "image/heic":
		return ImageTypeHEIC, nil
	default:
		return ImageTypeNotSupported, nil
	}
}

func resizeImage(img image.Image, size Size) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	newMaxSide := float64(size.Px())

	// Calculate scaling factor
	scale := math.Min(newMaxSide/float64(width), newMaxSide/float64(height))
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create a new image with the target size
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)

	return dst
}

func convertImageToByte(img image.Image) ([]byte, error) {
	// Create buffer to hold JPEG bytes
	var buf bytes.Buffer

	// Encode as JPEG with quality setting (1-100)
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}
