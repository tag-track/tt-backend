package thumbnail

import (
	"fmt"
	"github.com/lucsky/cuid"
	"mime/multipart"
	"sync"
)

type Size int

const (
	ExtraSmallSize Size = 128
	SmallSize      Size = 480
	MediumSize     Size = 640
	LargeSize      Size = 1024
	ExtraLargeSize Size = 2048
)

func (s Size) Px() int {
	return int(s)
}

type Thumbnails struct {
	EntityId      string
	Discriminator string

	ExtraSmall []byte // 128 Max len
	Small      []byte // 480 Max len
	Medium     []byte // 640 Max len
	Large      []byte // 1024 Max len
	ExtraLarge []byte // 2048 Max len
}

type NewThumbnailOption = func(th *Thumbnails)

func WithThumbnailSize(img []byte, size Size) NewThumbnailOption {
	return func(th *Thumbnails) {

		switch size {
		case ExtraSmallSize:
			th.ExtraSmall = img
			return
		case SmallSize:
			th.Small = img
			return
		case MediumSize:
			th.Medium = img
			return
		case LargeSize:
			th.Large = img
			return
		case ExtraLargeSize:
			th.ExtraLarge = img
			return
		}

	}
}

func WithThumbnailEntityId(id string) NewThumbnailOption {
	return func(th *Thumbnails) {
		th.EntityId = id
	}
}

func WithThumbnailDiscriminator() NewThumbnailOption {
	return func(th *Thumbnails) {
		th.Discriminator = cuid.New()
	}
}

func NewThumbnails(opts ...NewThumbnailOption) *Thumbnails {
	th := &Thumbnails{}
	for _, o := range opts {
		o(th)
	}
	return th
}

func NewThumbnailsFromMultipart(file multipart.File, entityId string) (*Thumbnails, error) {
	jpegImg, err := convertFileToJpeg(file)
	if err != nil {
		return nil, err
	}

	sizes := []Size{ExtraSmallSize, SmallSize, MediumSize, LargeSize, ExtraLargeSize}
	thumbnailOpt := make([]NewThumbnailOption, 0, len(sizes)) // Correct initialization

	var wg sync.WaitGroup
	wg.Add(len(sizes))

	var mut sync.Mutex
	var fnError error

	for _, s := range sizes {
		go func(s Size) { // Capture 's' as a parameter
			defer wg.Done()

			img := resizeImage(jpegImg, s)
			b, err := convertImageToByte(img)
			if err != nil {
				mut.Lock()
				if fnError == nil {
					fnError = err
				}
				mut.Unlock()
				return
			}

			opt := WithThumbnailSize(b, s)

			mut.Lock()
			thumbnailOpt = append(thumbnailOpt, opt)
			mut.Unlock()
		}(s) // Pass current 's' to the goroutine
	}

	wg.Wait()

	if fnError != nil {
		return nil, fnError
	}

	thumbnailOpt = append(thumbnailOpt, WithThumbnailEntityId(entityId))
	thumbnailOpt = append(thumbnailOpt, WithThumbnailDiscriminator())

	return NewThumbnails(thumbnailOpt...), nil
}

func (t *Thumbnails) GetImageBaseName() string {
	return fmt.Sprintf("%s_%s", t.EntityId, t.Discriminator)
}

func (t *Thumbnails) GetImageBaseNameWithExt() string {
	return fmt.Sprintf("%s.jpeg", t.GetImageBaseName())
}

func (t *Thumbnails) GetImageDataMap() *map[string][]byte {
	res := make(map[string][]byte)

	baseName := t.GetImageBaseName()

	res[fmt.Sprintf("%s_%s.jpeg", baseName, "xs")] = t.ExtraSmall
	res[fmt.Sprintf("%s_%s.jpeg", baseName, "s")] = t.Small
	res[fmt.Sprintf("%s_%s.jpeg", baseName, "m")] = t.Medium
	res[fmt.Sprintf("%s_%s.jpeg", baseName, "l")] = t.Large
	res[fmt.Sprintf("%s_%s.jpeg", baseName, "xl")] = t.ExtraLarge

	return &res
}
