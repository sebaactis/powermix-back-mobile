package routes

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/security/auth"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	ProofHandler   *proof.HTTPHandler
	VoucherHandler *voucher.HTTPHandler
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
		// Auth
		r.Post("/register", d.UserHandler.Create)
		r.Post("/login", d.AuthHandler.Login)
		r.Post("/login-google", d.AuthHandler.OAuthGoogle)

		// User password
		r.Get("/recoveryPassword", d.AuthHandler.RecoveryPasswordRequest)
		r.Post("/updatePasswordRecovery", d.AuthHandler.UpdatePasswordByRecovery)

		// Token
		r.Post("/refreshToken", d.AuthHandler.RefreshToken)

		r.Group(func(pr chi.Router) {
			pr.Use(d.AuthMiddleware.RequireAuth())

			// User
			pr.Get("/user/{id}", d.UserHandler.GetByID)
			pr.Get("/user/me", d.UserHandler.Me)
			pr.Put("/user/update", d.UserHandler.Update)

			// Proof
			pr.Get("/proofs/me", d.ProofHandler.GetAllByUserId)
			pr.Get("/proofs/me/paginated", d.ProofHandler.GetAllByUserIdPaginated)
			pr.Get("/proofs/me/last3", d.ProofHandler.GetLastThreeByUserId)
			pr.Get("/proofs/me/{id}", d.ProofHandler.GetById)
			pr.Post("/proof", d.ProofHandler.Create)
			pr.Post("/proof/others", d.ProofHandler.CreateFromOthers)

		})
	})

	return r
}
