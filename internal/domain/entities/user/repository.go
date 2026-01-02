package user

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sebaactis/powermix-back-mobile/internal/security/oauth"
	"gorm.io/gorm"
)

// isDuplicateKeyError verifica si el error es de clave duplicada (constraint violation)
// PostgreSQL error code 23505: unique_violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// Intentar obtener el error específico de PostgreSQL
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}

	// Fallback para otros casos
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint")
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// WithTx returns a new Repository that uses the given transaction
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB exposes the underlying db connection for transaction management
func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Create(ctx context.Context, user *User) error {

	var existing User


	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		insertErr := tx.Create(user).Error

		if insertErr != nil {

			if isDuplicateKeyError(insertErr) {

				if err := tx.Where("email = ?", user.Email).First(&existing).Error; err != nil {
					return err
				}


				if strings.TrimSpace(existing.Password) != "" {
					return ErrDuplicateEmail
				}

				existing.Password = user.Password
				return tx.Save(&existing).Error
			}
			return insertErr
		}

		return nil
	})

	return err
}

func (r *Repository) CreateWithOAuth(ctx context.Context, info *oauth.OAuthUserInfo) (*User, error) {
	var newUser User


	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		newUser = User{
			Name:          info.Name,
			Email:         info.Email,
			OAuthProvider: info.Provider,
			OAuthID:       info.ProviderID,
			StampsCounter: 0,
		}

		insertErr := tx.Create(&newUser).Error

		if insertErr != nil {

			if isDuplicateKeyError(insertErr) {

				if err := tx.Where("email = ?", info.Email).First(&newUser).Error; err != nil {
					return err
				}

				if newUser.OAuthProvider != "" {
					log.Printf("ℹ️ Usuario ya existe con OAuth: %+v", newUser)
					return nil
				}

				newUser.OAuthProvider = info.Provider
				newUser.OAuthID = info.ProviderID

				if err := tx.Save(&newUser).Error; err != nil {
					return err
				}

				log.Printf("🔁 Usuario existente actualizado con OAuth: %+v", newUser)
				return nil
			}
			return insertErr
		}

		log.Printf("✅ Usuario nuevo con OAuth creado: %+v", newUser)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
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

func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*User, error) {
	result := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("Usuario no encontrado")
	}

	user, err := r.FindByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) (*User, error) {
	var user User

	err := r.db.WithContext(ctx).Raw(`
		UPDATE users
		SET password = ?, login_attempt = 0, locked_until = NULL
		WHERE id = ?
		RETURNING *
	`, hashedPassword, id).Scan(&user).Error

	if err != nil {
		return nil, err
	}

	if r.db.RowsAffected == 0 {
		return nil, errors.New("Usuario no encontrado")
	}

	return &user, nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *Repository) IncrementStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	result := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("stamps_counter", gorm.Expr("stamps_counter + ?", 1))

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected == 0 {
		return 0, errors.New("user not found")
	}

	var user User

	if err := r.db.WithContext(ctx).Select("stamps_counter").First(&user, id).Error; err != nil {
		return 0, err
	}

	return user.StampsCounter, nil
}

func (r *Repository) ResetStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	result := r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("stamps_counter", 0)

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected == 0 {
		return 0, errors.New("user not found")
	}

	var user User

	if err := r.db.WithContext(ctx).Select("stamps_counter").First(&user, id).Error; err != nil {
		return 0, err
	}

	return user.StampsCounter, nil
}

func (r *Repository) IncrementLoginAttempt(ctx context.Context, id uuid.UUID) (int, error) {

	var newAttemptCount int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.Raw(`
			UPDATE users
			SET login_attempt = login_attempt + 1
			WHERE id = ?
			RETURNING login_attempt
		`, id).Scan(&newAttemptCount).Error

		if err != nil {
			return err
		}

		if newAttemptCount == 0 {
			return errors.New("No existe el usuario proporcionado")
		}

		if newAttemptCount >= 5 {
			now := time.Now()
			err := tx.Model(&User{}).
				Where("id = ?", id).
				Update("locked_until", now.Add(15*time.Minute)).Error

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return newAttemptCount, nil
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

func (r *Repository) UnlockUser(ctx context.Context, id uuid.UUID) error {

	updates := map[string]interface{}{
		"locked_until":  nil,
		"login_attempt": 0,
	}

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

var ErrDuplicateEmail = errors.New("el email ya esta en uso")
