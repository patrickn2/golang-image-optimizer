package imagecompress

import "github.com/h2non/bimg"

type PkgImgCompress struct {
}

func NewImageCompress() *PkgImgCompress {
	return &PkgImgCompress{}
}

func (ic *PkgImgCompress) CompressImage(c *CompressImageRequest) ([]byte, error) {
	options := bimg.Options{
		Quality: c.Quality,
		Width:   c.Width,
		Height:  c.Height,
		Type:    bimg.WEBP,
	}
	return bimg.NewImage(c.ImageData).Process(options)
}
