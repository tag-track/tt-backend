package thumbnail

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/jdeng/goheif"
	"github.com/rwcarlsen/goexif/exif"
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

func applyOrientation(img image.Image, exifOri int) image.Image {

	switch exifOri {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

func getOrientationFromExif(exifData []byte) (int, error) {
	x, err := exif.Decode(bytes.NewReader(exifData))
	if err != nil {
		return 0, err
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 0, err
	}

	return tag.Int(0)
}

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

		exifBytes, err := goheif.ExtractExif(bytes.NewReader(data))
		if err != nil {
			break
		}
		if len(exifBytes) == 0 {
			break
		}

		ori, err := getOrientationFromExif(exifBytes)
		if err != nil {
			break
		}

		img = applyOrientation(img, ori)

	default:
		return nil, fmt.Errorf("unsupported format")
	}

	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return img, nil
}

// HEIC signature check helper
func isHEIC(data []byte) bool {
	// Minimum required length check
	if len(data) < 12 {
		return false
	}

	// Check for "ftyp" file type box at byte 4 (ISO BMFF container)
	if !bytes.Equal(data[4:8], []byte("ftyp")) {
		return false
	}

	// Check HEIC compatible brand identifiers
	majorBrand := string(data[8:12])
	heicBrands := []string{
		"heic", // HEIC image (HEVC)
		"heix", // HEIC image (HEVC 10-bit)
		"heim", // HEIC image (HEVC monochrome)
		"heis", // HEIC image (HEVC sRGB)
		"hevc", // HEVC sequence
		"hevx", // HEVC encoded image
		"mif1", // High Efficiency Image Format (HEIF)
	}

	for _, brand := range heicBrands {
		if majorBrand == brand {
			return true
		}
	}
	return false
}

func detectImageTypeFromBytes(data []byte) (ImageType, error) {

	if isHEIC(data) {
		return ImageTypeHEIC, nil
	}

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
	default:
		return ImageTypeNotSupported, fmt.Errorf("image type not supported: %v", mimeType)
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
