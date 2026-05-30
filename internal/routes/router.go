package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/prode"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/config"
	"github.com/sebaactis/powermix-back-mobile/internal/security/auth"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	ProofHandler   *proof.HTTPHandler
	VoucherHandler *voucher.HTTPHandler
	TokenHandler   *token.HTTPHandler
	AuthHandler    *auth.HTTPHandler
	ProdeHandler   *prode.HTTPHandler
	Config         config.Config
	Validator      *validations.Validator
	RateLimiter    *middlewares.RateLimiter
	AuthMiddleware *middlewares.AuthMiddleware
}

func Router(d Deps) *chi.Mux {

	r := chi.NewRouter()

	r.Use(
		middlewares.RequestID(),
		middlewares.RequestLogger(slog.Default()),
		middlewares.Recoverer(slog.Default()),
		middlewares.JSONContentType(),
		d.RateLimiter.Middleware(),
		middlewares.Timeout(30*time.Second),
	)

	r.Route("/api/v1", func(r chi.Router) {
		// Autenticación
		r.Post("/register", d.UserHandler.Create)
		r.Post("/login", d.AuthHandler.Login)
		r.Post("/login-google", d.AuthHandler.OAuthGoogle)

		// Password de usuario
		r.Post("/recoveryPassword", d.AuthHandler.RecoveryPasswordRequest)
		r.Post("/updatePasswordRecovery", d.AuthHandler.UpdatePasswordByRecovery)

		// Token
		r.Post("/refreshToken", d.AuthHandler.RefreshToken)

		r.Group(func(pr chi.Router) {
			pr.Use(d.AuthMiddleware.RequireAuth())

			// User
			pr.Get("/user/{id}", d.UserHandler.GetByID)
			pr.Get("/user/me", d.UserHandler.Me)
			pr.Put("/user/update", d.UserHandler.Update)
			pr.Put("/user/change-password", d.UserHandler.UpdatePassword)
			pr.Post("/user/contact", d.UserHandler.SendEmailContact)

			// Proof
			pr.Get("/proofs/me", d.ProofHandler.GetAllByUserID)
			pr.Get("/proofs/me/paginated", d.ProofHandler.GetAllByUserIDPaginated)
			pr.Get("/proofs/me/last3", d.ProofHandler.GetLastThreeByUserID)
			pr.Get("/proofs/me/{id}", d.ProofHandler.GetByID)
			pr.Post("/proof", d.ProofHandler.Create)
			pr.Post("/proof/others", d.ProofHandler.CreateFromOthers)

			// Voucher
			pr.Get("/voucher/me", d.VoucherHandler.GetAllByUserID)
			pr.Get("/voucher/available", d.VoucherHandler.GetAvailableCount)
			pr.Delete("/voucher/{id}", d.VoucherHandler.DeleteVoucher)

			// PRODE
			if d.Config.IsProdeEnabled() {
				pr.Get("/prode/matches", d.ProdeHandler.ListMatches)
				pr.Get("/prode/matches/{matchID}", d.ProdeHandler.GetMatch)
				pr.Put("/prode/matches/{matchID}/prediction", d.ProdeHandler.CreateOrUpdatePrediction)
				pr.Get("/prode/predictions/me", d.ProdeHandler.GetMyPredictions)
			}
		})

		// PRODE Admin — protegido por maintenance key
		if d.Config.IsProdeEnabled() {
			r.Group(func(ar chi.Router) {
				ar.Use(middlewares.MaintenanceKey(d.Config))

				ar.Post("/prode/admin/matches", d.ProdeHandler.AdminCreateMatch)
				ar.Patch("/prode/admin/matches/{matchID}", d.ProdeHandler.AdminUpdateMatch)
				ar.Put("/prode/admin/matches/{matchID}/result", d.ProdeHandler.AdminRecordResult)
				ar.Post("/prode/admin/matches/{matchID}/settle", d.ProdeHandler.AdminSettleMatch)
				ar.Post("/prode/admin/rewards/retry", d.ProdeHandler.AdminRetryPendingRewards)
			})
		}
	})

	return r
}
