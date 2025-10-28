package routes

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/security/auth"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	ProofHandler   *proof.HTTPHandler
	TokenHandler   *token.HTTPHandler
	AuthHandler    *auth.HTTPHandler
	Validator      *validations.Validator
	RateLimiter    *middlewares.RateLimiter
	AuthMiddleware *middlewares.AuthMiddleware
}

func Router(d Deps) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middlewares.Logger(), middlewares.JSONContentType(), middlewares.Timeout(30*time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", d.UserHandler.Create)
		r.Post("/login", d.AuthHandler.Login)
		r.Get("/recoveryPassword", d.AuthHandler.RecoveryPasswordRequest)
		r.Post("/updatePasswordRecovery", d.AuthHandler.UpdatePasswordByRecovery)
		r.Post("/refreshToken", d.AuthHandler.RefreshToken)

		r.Group(func(pr chi.Router) {
			pr.Use(d.AuthMiddleware.RequireAuth())

			pr.Get("/user/{id}", d.UserHandler.GetByID)

			pr.Get("/me/proofs", d.ProofHandler.GetAllByUserId)
			pr.Get("/me/proofs/{id}", d.ProofHandler.GetById)	
			pr.Post("/proof", d.ProofHandler.Create)
		})
	})

	return r
}
