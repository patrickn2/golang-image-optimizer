package imagecompress

import "github.com/h2non/bimg"

type PkgImgCompress struct {
}

func NewImageCompress() *PkgImgCompress {
	return &PkgImgCompress{}
}

func (ic *PkgImgCompress) CompressImage(image []byte, quality, width int) ([]byte, error) {
	options := bimg.Options{
		Quality: quality,
		Width:   width,
		Type:    bimg.WEBP,
	}
	return bimg.NewImage(image).Process(options)
}
