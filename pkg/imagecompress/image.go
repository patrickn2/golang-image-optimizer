package imagecompress

type CompressImageRequest struct {
	ImageData []byte
	ImageType string
	Quality   int
	Width     int
	Height    int
	NewType   string
}

type PkgImgCompressInterface interface {
	CompressImage(*CompressImageRequest) ([]byte, error)
}
