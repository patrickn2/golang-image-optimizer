package imagecompress

import (
	"github.com/davidbyttow/govips/v2/vips"
)

type PkgImgGoVips struct {
}

func NewImageGoVips() *PkgImgGoVips {
	vips.LoggingSettings(nil, vips.LogLevelError)
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
		h := c.Height
		if h > img.Height() {
			h = img.Height()
		}
		vScale = float64(h) / float64(img.Height())
	}
	w := c.Width
	if w > img.Width() {
		w = img.Width()
	}
	hScale := float64(w) / float64(img.Width())
	err = img.ResizeWithVScale(hScale, vScale, vips.KernelAuto)
	if err != nil {
		return nil, err
	}

	var exportParams *vips.ExportParams
	switch c.Type {
	case "image/png":
		exportParams = vips.NewDefaultPNGExportParams()
	case "image/jpeg":
		exportParams = vips.NewDefaultJPEGExportParams()
	case "image/gif":
		exportParams = vips.NewDefaultExportParams()
		exportParams.Format = vips.ImageTypeGIF
	case "image/avif":
		exportParams = vips.NewDefaultExportParams()
		exportParams.Format = vips.ImageTypeAVIF
	default:
		exportParams = vips.NewDefaultWEBPExportParams()
	}
	exportParams.Quality = c.Quality
	image, _, err := img.Export(exportParams)
	return image, err
}
