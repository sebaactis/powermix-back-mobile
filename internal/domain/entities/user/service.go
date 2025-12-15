package user

import (
	"context"
	"errors"
	"strings"

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

func (s *Service) Create(ctx context.Context, user *UserCreate) (*User, error) {
	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(user.Password)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	if fields, ok := s.validator.ValidateStruct(user); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	newUser := &User{
		Name:     name,
		Email:    email,
		Password: string(passwordHash),
	}

	if err := s.repository.Create(ctx, newUser); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateEmail
		}

		return nil, err
	}

	return newUser, nil
}

func (s *Service) FindOrCreateFromOAuth(ctx context.Context, info *oauth.OAuthUserInfo) (*User, error) {
	return s.repository.CreateWithOAuth(ctx, info)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repository.FindByEmail(ctx, email)
}

func (s *Service) IncrementLoginAttempt(ctx context.Context, id uuid.UUID) (int, error) {
	return s.repository.IncrementLoginAttempt(ctx, id)
}

func (s *Service) IncrementStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	return s.repository.IncrementStampsCounter(ctx, id)
}

func (s *Service) ResetStampsCounter(ctx context.Context, id uuid.UUID) (int, error) {
	return s.repository.ResetStampsCounter(ctx, id)
}

func (s *Service) UnlockUser(ctx context.Context, id uint) error {
	return s.repository.UnlockUser(ctx, id)
}

func (s *Service) UpdatePasswordByRecovery(ctx context.Context, req UserRecoveryPassword) (*User, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	user, err := s.repository.FindByID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user, err = s.repository.UpdatePassword(ctx, user.ID, string(passwordHash))

	if err != nil {
		return nil, err
	}

	_, err = s.tokenService.ValidateAndRevokeResetPasswordToken(ctx, req.Token)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdatePassword(ctx context.Context, userId uuid.UUID, req UserChangePassword) (*User, error) {
	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	_, err := s.repository.FindByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.NewPassword)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	return s.repository.UpdatePassword(ctx, userId, string(passwordHash))

}

func (s *Service) Update(ctx context.Context, userId uuid.UUID, req UserUpdate) (*User, error) {
	updates := map[string]interface{}{}

	user, err := s.GetByID(ctx, userId)

	if err != nil {
		return nil, err
	}

	if strings.EqualFold(user.Name, *req.Name) {
		return nil, ErrSameName
	}

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if len(updates) == 0 {
		return nil, errors.New("no hay campos para actulizar")
	}

	return s.repository.Update(ctx, userId, updates)
}

func (s *Service) SendEmailContact(ctx context.Context, req mailer.ContactRequest) (*mailer.ContactResponse, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validations.ValidationError{Fields: fields}
	}

	if err := s.mailer.SendEmailContact(ctx, &req); err != nil {
		return nil, err
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
