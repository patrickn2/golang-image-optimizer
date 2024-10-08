package imagecompress

type CompressImageRequest struct {
	ImageData []byte
	Quality   int
	Width     int
	Height    int
}

type PkgImgCompressInterface interface {
	CompressImage(*CompressImageRequest) ([]byte, error)
}
