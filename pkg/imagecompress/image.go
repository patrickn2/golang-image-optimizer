package imagecompress

type PkgImgCompressInterface interface {
	CompressImage(image []byte, quality, width int) ([]byte, error)
}
