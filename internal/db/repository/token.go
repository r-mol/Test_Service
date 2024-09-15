package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/r-mol/Test_Service/internal/db/model"
	"github.com/r-mol/Test_Service/internal/domain"
	"github.com/r-mol/Test_Service/pkg/pg"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type TokenRepo struct {
	pg *pg.Client
}

func NewTokenRepo(pg *pg.Client) *TokenRepo {
	return &TokenRepo{
		pg: pg,
	}
}

func (r *TokenRepo) GetByUserID(
	ctx context.Context, userID uuid.UUID,
) (*domain.Token, error) {
	db, err := r.pg.StandbyPreferredGormDB()
	if err != nil {
		return nil, fmt.Errorf("get standby db connection: %w", err)
	}

	entity := new(model.Token)
	err = db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		First(entity).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}
		return nil, fmt.Errorf("unable to get approval by id at db level: %w", err)
	}

	return &domain.Token{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Hash:      entity.Hash,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}, nil
}

func (r *TokenRepo) UpdateOrCreate(ctx context.Context, token *domain.Token) (int, error) {
	db, err := r.pg.PrimaryGormDB()
	if err != nil {
		return 0, fmt.Errorf("get primary db connection: %w", err)
	}

	// Create a new entity with UserID and initial hash
	entity := &model.Token{UserID: token.UserID, Hash: token.Hash}

	// Define upsert behavior
	err = db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}}, // Column which has unique index
		DoUpdates: clause.Assignments(map[string]interface{}{"hash": token.Hash}),
	}).Create(entity).Error

	if err != nil {
		return 0, fmt.Errorf("unable to upsert token: %w", err)
	}

	// If the entity is newly created, rowsAffected will be 1 and entity.ID will already be set.
	// If an update occurs, you might want to retrieve the ID if not known:
	if entity.ID == 0 {
		var existingEntity model.Token
		if err := db.WithContext(ctx).Model(&model.Token{}).Where("user_id = ?", token.UserID).First(&existingEntity).Error; err != nil {
			return 0, fmt.Errorf("failed to retrieve token post-upsert: %w", err)
		}
		entity.ID = existingEntity.ID
	}

	return entity.ID, nil
}
