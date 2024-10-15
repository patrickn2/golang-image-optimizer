package imagecompress

import (
	"github.com/h2non/bimg"
)

type PkgImgCompress struct {
}

func NewImageCompress() *PkgImgCompress {
	return &PkgImgCompress{}
}

// func (ic *PkgImgCompress) GetImageType(image []byte) string {
// 	return bimg.NewImage(image).Type()
// }

func (ic *PkgImgCompress) CompressImage(c *CompressImageRequest) ([]byte, error) {
	var bimgType bimg.ImageType
	switch c.Type {
	case "image/png":
		bimgType = bimg.PNG
	case "image/jpeg":
		bimgType = bimg.JPEG
	case "image/gif":
		bimgType = bimg.GIF
	case "image/avif":
		bimgType = bimg.AVIF
	default:
		bimgType = bimg.WEBP
	}
	options := bimg.Options{
		Quality: c.Quality,
		Width:   c.Width,
		Height:  c.Height,
		Type:    bimgType,
	}
	newImage := bimg.NewImage(c.ImageData)
	newImageSize, err := newImage.Size()
	if err != nil {
		return nil, err
	}
	if c.Height != 0 && c.Height > newImageSize.Height {
		options.Height = newImageSize.Height
	}
	if c.Width > newImageSize.Width {
		options.Width = newImageSize.Width
	}
	return newImage.Process(options)
}
