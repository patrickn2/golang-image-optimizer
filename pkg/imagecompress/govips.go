package imagecompress

import (
	"github.com/davidbyttow/govips/v2/vips"
)

type PkgImgGoVips struct {
}

func NewImageGoVips() *PkgImgGoVips {
	vips.Startup(nil)
	return &PkgImgGoVips{}
}

func (ic *PkgImgGoVips) CloseVips() {
	vips.Shutdown()
}

func (ic *PkgImgGoVips) CompressImage(c *CompressImageRequest) ([]byte, error) {
	importParams := vips.NewImportParams()
	importParams.NumPages.Set(-1) // Load all Gif Frames
	img, err := vips.LoadImageFromBuffer(c.ImageData, importParams)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	var vScale float64 = -1
	if c.Height != 0 {
		vScale = float64(c.Height) / float64(img.Height())
	}

	hScale := float64(c.Width) / float64(img.Width())
	err = img.ResizeWithVScale(hScale, vScale, vips.KernelAuto)
	if err != nil {
		return nil, err
	}

	var vipsImageType vips.ImageType
	switch c.Type {
	case "image/png":
		vipsImageType = vips.ImageTypePNG
	case "image/jpeg":
		vipsImageType = vips.ImageTypeJPEG
	case "image/gif":
		vipsImageType = vips.ImageTypeGIF
	case "image/avif":
		vipsImageType = vips.ImageTypeAVIF
	default:
		vipsImageType = vips.ImageTypeWEBP
	}
	exportParams :=
		vips.ExportParams{
			Quality: c.Quality,
			Format:  vipsImageType,
			Effort:  4,
		}
	image, _, err := img.Export(&exportParams)
	return image, err
}
