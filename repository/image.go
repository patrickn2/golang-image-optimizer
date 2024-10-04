package repository

import (
	"context"

	"github.com/patrickn2/go-image-optimizator/pkg/database"
)

type ImageRepository struct {
	db database.PkgDatabaseInterface
}

func NewImageRepository(db database.PkgDatabaseInterface) *ImageRepository {
	return &ImageRepository{
		db: db,
	}
}

func (ir *ImageRepository) GetImage(ctx context.Context, imageName string) ([]byte, error) {
	return ir.db.Get(ctx, imageName)
}

func (ir *ImageRepository) SaveImage(ctx context.Context, imageName string, image []byte) error {
	return ir.db.Set(ctx, imageName, image)
}
