package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]*User, error) {
	users := []*User{}
	result := r.db.WithContext(ctx).Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User

	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	return r.Update(ctx, id, map[string]interface{}{
		"password": hashedPassword,
	})
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *Repository) IncrementLoginAttempt(ctx context.Context, id uuid.UUID) (int, error) {

	result := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("login_attempt", gorm.Expr("login_attempt + ?", 1))

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected == 0 {
		return 0, errors.New("user not found")
	}

	var user User

	if err := r.db.WithContext(ctx).Select("login_attempt").First(&user, id).Error; err != nil {
		return 0, err
	}

	if user.LoginAttempt >= 5 {
		r.LockedUser(ctx, id)
	}

	return user.LoginAttempt, nil
}

func (r *Repository) LockedUser(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("locked_until", now.Add(15*time.Minute)); err != nil {
		return err.Error
	}

	return nil
}

func (r *Repository) UnlockUser(ctx context.Context, id uint) error {

	if err := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("locked_until", nil).
		Update("login_attempt", 0); err != nil {
		return err.Error
	}

	return nil
}

var ErrDuplicateEmail = errors.New("email already in use")
