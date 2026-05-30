package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/security/oauth"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrSameName = errors.New("el nombre no puede ser igual al actual")

type Service struct {
	repository   *Repository
	tokenService *token.Service
	validator    validations.StructValidator
	mailer       mailer.Mailer
	db           *gorm.DB
}

func NewService(repository *Repository, tokenService *token.Service, v validations.StructValidator, mailer mailer.Mailer) *Service {
	return &Service{repository: repository, tokenService: tokenService, db: repository.db, validator: v, mailer: mailer}
}

func (s *Service) WithTx(tx *gorm.DB) *Service {
	return &Service{
		repository:   s.repository.WithTx(tx),
		tokenService: s.tokenService,
		validator:    s.validator,
		mailer:       s.mailer,
		db:           tx,
	}
}

func (s *Service) Create(ctx context.Context, user *UserCreate) (*User, error) {
	if fields, ok := s.validator.ValidateStruct(user); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(user.Password)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	newUser := &User{
		Name:     name,
		Email:    email,
		Password: string(passwordHash),
	}

	if err := s.repository.Create(ctx, newUser); err != nil {
		return nil, wrapServiceErr("create", err)
	}

	return newUser, nil
}

func (s *Service) FindOrCreateFromOAuth(ctx context.Context, info *oauth.OAuthUserInfo) (*User, error) {
	u, err := s.repository.CreateWithOAuth(ctx, info)
	if err != nil {
		return nil, wrapServiceErr("find or create oauth", err)
	}
	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, wrapServiceErr("get by id", err)
	}
	return u, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		return nil, wrapServiceErr("get by email", err)
	}
	return u, nil
}

func (s *Service) IncrementLoginAttempt(ctx context.Context, id uuid.UUID) (int, error) {
	n, err := s.repository.IncrementLoginAttempt(ctx, id)
	if err != nil {
		return 0, wrapServiceErr("increment login attempt", err)
	}
	return n, nil
}

func (s *Service) IncrementStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	n, err := s.repository.IncrementStampsCounter(ctx, id)
	if err != nil {
		return 0, wrapServiceErr("increment stamps counter", err)
	}
	return n, nil
}

func (s *Service) ResetStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	n, err := s.repository.ResetStampsCounter(ctx, id)
	if err != nil {
		return 0, wrapServiceErr("reset stamps counter", err)
	}
	return n, nil
}

func (s *Service) UnlockUser(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.UnlockUser(ctx, id); err != nil {
		return wrapServiceErr("unlock user", err)
	}
	return nil
}

func (s *Service) CheckAndUnlockIfExpired(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return false, wrapServiceErr("check and unlock find user", err)
	}

	if user.LockedUntil.IsZero() || user.LockedUntil.Before(time.Now()) {

		if !user.LockedUntil.IsZero() {

			err := s.repository.UnlockUser(ctx, user.ID)
			if err != nil {
				return false, wrapServiceErr("check and unlock", err)
			}
			return true, nil
		}
		return false, nil
	}

	return false, nil
}

func (s *Service) UpdatePasswordByRecovery(ctx context.Context, req UserRecoveryPassword) (*User, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var updatedUser *User

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		txUserRepo := s.repository.WithTx(tx)
		txTokenService := s.tokenService.WithTx(tx)

		user, err := txUserRepo.FindByID(ctx, req.UserID)
		if err != nil {
			return wrapServiceErr("update password by recovery find user", err)
		}

		updatedUser, err = txUserRepo.UpdatePassword(ctx, user.ID, string(passwordHash))
		if err != nil {
			return wrapServiceErr("update password by recovery", err)
		}

		_, err = txTokenService.ValidateAndRevokeResetPasswordToken(ctx, req.Token)
		if err != nil {
			return wrapServiceErr("update password by recovery revoke token", err)
		}

		return nil
	})

	if err != nil {
		return nil, wrapServiceErr("update password by recovery transaction", err)
	}

	return updatedUser, nil
}

func (s *Service) UpdatePassword(ctx context.Context, userId uuid.UUID, req UserChangePassword) (*User, error) {
	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	_, err := s.repository.FindByID(ctx, userId)
	if err != nil {
		return nil, wrapServiceErr("update password find user", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.NewPassword)), bcrypt.DefaultCost)

	if err != nil {
		return nil, wrapServiceErr("update password hash", err)
	}

	u, err := s.repository.UpdatePassword(ctx, userId, string(passwordHash))
	if err != nil {
		return nil, wrapServiceErr("update password", err)
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, userId uuid.UUID, req UserUpdate) (*User, error) {
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		return nil, &validations.ValidationError{Fields: map[string]string{"name": "El nombre es requerido"}}
	}

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	user, err := s.GetByID(ctx, userId)
	if err != nil {
		return nil, wrapServiceErr("update find user", err)
	}

	if strings.EqualFold(user.Name, *req.Name) {
		return nil, ErrSameName
	}

	updates := map[string]interface{}{"name": *req.Name}

	u, err := s.repository.Update(ctx, userId, updates)
	if err != nil {
		return nil, wrapServiceErr("update", err)
	}
	return u, nil
}

func wrapServiceErr(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("user service: %s: %w", action, err)
}

func (s *Service) SendEmailContact(ctx context.Context, req mailer.ContactRequest) (*mailer.ContactResponse, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	if err := s.mailer.SendEmailContact(ctx, &req); err != nil {
		return nil, wrapServiceErr("send email contact", err)
	}

	newContactResponse := mailer.ContactResponse{
		Name:       req.Name,
		Email:      req.Email,
		Category:   req.Category,
		Message:    req.Message,
		ApiMessage: "Consulta enviada correctamente!",
	}

	return &newContactResponse, nil

}
