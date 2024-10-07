package repository

import (
	"context"
	"time"

	"github.com/patrickn2/go-image-optimizer/pkg/database"
)

type ImageRepository struct {
	db database.PkgDatabaseInterface
}

func NewImageRepository(db database.PkgDatabaseInterface) *ImageRepository {
	return &ImageRepository{
		db: db,
	}
}

func (ir *ImageRepository) GetImage(ctx context.Context, imageName string) ([]byte, *time.Time, error) {
	return ir.db.Get(ctx, imageName)
}

func (ir *ImageRepository) SaveImage(ctx context.Context, imageName string, image []byte) error {
	return ir.db.Set(ctx, imageName, image)
}
