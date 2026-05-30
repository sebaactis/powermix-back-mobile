package prode

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
)

// Clock permite inyectar la hora actual para tests (sin tests por ahora, pero
// la interfaz queda para mantener el diseño limpio).
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type Service struct {
	repo        *Repository
	voucherRepo *voucher.Repository
	userRepo    *user.Repository
	mailer      mailer.Mailer
	adminEmails []string
	clock       Clock
}

func NewService(repo *Repository, voucherRepo *voucher.Repository, userRepo *user.Repository, mailer mailer.Mailer, adminEmails []string) *Service {
	return &Service{
		repo:        repo,
		voucherRepo: voucherRepo,
		userRepo:    userRepo,
		mailer:      mailer,
		adminEmails: adminEmails,
		clock:       realClock{},
	}
}

// WithTx devuelve un Service que opera sobre la transacción recibida.
func (s *Service) WithTx(txRepo *Repository) *Service {
	return &Service{
		repo:        txRepo,
		voucherRepo: s.voucherRepo,
		userRepo:    s.userRepo,
		mailer:      s.mailer,
		adminEmails: s.adminEmails,
		clock:       s.clock,
	}
}

// ListMatches devuelve los partidos visibles con la predicción del usuario si existe.
func (s *Service) ListMatches(ctx context.Context) ([]MatchResponse, error) {
	userID, _ := middlewares.UserIDFromContext(ctx)

	matches, err := s.repo.GetVisibleMatches(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error al obtener partidos visibles", "error", err)
		return nil, err
	}

	responses := make([]MatchResponse, 0, len(matches))
	for _, m := range matches {
		resp := matchToResponse(m)

		if userID != uuid.Nil {
			pred, err := s.repo.GetUserPrediction(ctx, userID, m.ID)
			if err != nil {
				slog.ErrorContext(ctx, "error al obtener predicción del usuario", "match_id", m.ID, "user_id", userID, "error", err)
			} else if pred != nil {
				resp.MyPrediction = predictionToResponse(pred)
			}
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// GetMatch devuelve un partido con la predicción del usuario.
func (s *Service) GetMatch(ctx context.Context, matchID uuid.UUID) (*MatchResponse, error) {
	userID, _ := middlewares.UserIDFromContext(ctx)

	match, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	resp := matchToResponse(*match)

	if userID != uuid.Nil {
		pred, err := s.repo.GetUserPrediction(ctx, userID, matchID)
		if err != nil {
			slog.ErrorContext(ctx, "error al obtener predicción del usuario", "match_id", matchID, "user_id", userID, "error", err)
		} else if pred != nil {
			resp.MyPrediction = predictionToResponse(pred)
		}
	}

	return &resp, nil
}

// CreateOrUpdatePrediction crea o actualiza la predicción de un usuario para un partido.
func (s *Service) CreateOrUpdatePrediction(ctx context.Context, userID uuid.UUID, matchID uuid.UUID, req PredictionRequest) (*PredictionResponse, error) {
	match, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "partido no encontrado para predicción", "match_id", matchID, "user_id", userID, "error", err)
		return nil, err
	}

	if !match.IsOpenForPrediction(s.clock.Now()) {
		slog.WarnContext(ctx, "intento de predicción fuera de plazo", "user_id", userID,
			"match_id", matchID,
			"kickoff", match.KickoffAt,
			"now", s.clock.Now(),)
		return nil, ErrCutoffPassed
	}

	pred := ProdePrediction{
		UserID:         userID,
		MatchID:        matchID,
		ArgentinaGoals: req.ArgentinaGoals,
		OpponentGoals:  req.OpponentGoals,
		Status:         PredStatusPending,
	}

	if err := pred.ValidateScore(); err != nil {
		return nil, err
	}

	created, err := s.repo.UpsertPrediction(ctx, &pred)
	if err != nil {
		slog.ErrorContext(ctx, "error al guardar predicción", "user_id", userID,
			"match_id", matchID,
			"error", err,)
		return nil, err
	}

	slog.InfoContext(ctx, "predicción guardada", "user_id", userID,
		"match_id", matchID,
		"argentina_goals", req.ArgentinaGoals,
		"opponent_goals", req.OpponentGoals,)

	return predictionToResponse(created), nil
}

// GetMyPredictions devuelve todas las predicciones del usuario autenticado.
func (s *Service) GetMyPredictions(ctx context.Context) ([]PredictionResponse, error) {
	userID, _ := middlewares.UserIDFromContext(ctx)

	predictions, err := s.repo.GetPredictionsByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "error al obtener predicciones del usuario", "user_id", userID, "error", err)
		return nil, err
	}

	responses := make([]PredictionResponse, 0, len(predictions))
	for _, p := range predictions {
		responses = append(responses, *predictionToResponse(&p))
	}

	return responses, nil
}

// ---- Admin methods ----

// CreateMatch crea un nuevo partido.
func (s *Service) CreateMatch(ctx context.Context, req CreateMatchRequest) (*AdminMatchResponse, error) {
	// Si el partido es visible y tiene horario futuro, arranca como SCHEDULED.
	// Si no es visible, arranca como DRAFT hasta que el admin lo programe.
	status := MatchStatusDraft
	if req.IsVisible && !req.KickoffAt.IsZero() {
		status = MatchStatusScheduled
	}

	match := ProdeMatch{
		Stage:     req.Stage,
		Opponent:  req.Opponent,
		KickoffAt: req.KickoffAt,
		IsVisible: req.IsVisible,
		Status:    status,
	}

	if err := s.repo.CreateMatch(ctx, &match); err != nil {
		slog.ErrorContext(ctx, "error al crear partido", "stage", req.Stage,
			"opponent", req.Opponent,
			"error", err,)
		return nil, err
	}

	slog.InfoContext(ctx, "partido creado", "match_id", match.ID,
		"stage", match.Stage,
		"opponent", match.Opponent,)

	return adminMatchToResponse(&match), nil
}

// UpdateMatch actualiza los campos de un partido existente.
func (s *Service) UpdateMatch(ctx context.Context, matchID uuid.UUID, req UpdateMatchRequest) (*AdminMatchResponse, error) {
	match, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	if req.Stage != nil {
		match.Stage = *req.Stage
	}
	if req.Opponent != nil {
		match.Opponent = *req.Opponent
	}
	if req.KickoffAt != nil {
		match.KickoffAt = *req.KickoffAt
	}
	if req.IsVisible != nil {
		match.IsVisible = *req.IsVisible
	}
	if req.Status != nil {
		match.Status = *req.Status
	}

	if err := s.repo.UpdateMatch(ctx, match); err != nil {
		slog.ErrorContext(ctx, "error al actualizar partido", "match_id", matchID,
			"error", err,)
		return nil, err
	}

	slog.InfoContext(ctx, "partido actualizado", "match_id", matchID,
		"status", match.Status,
		"is_visible", match.IsVisible,)

	return adminMatchToResponse(match), nil
}

// RecordResult registra el resultado de 90 minutos de un partido.
func (s *Service) RecordResult(ctx context.Context, matchID uuid.UUID, req RecordResultRequest) (*AdminMatchResponse, error) {
	match, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	if req.ArgentinaGoals < 0 || req.OpponentGoals < 0 {
		return nil, ErrInvalidScore
	}

	argentinaGoals := req.ArgentinaGoals
	opponentGoals := req.OpponentGoals
	match.ArgentinaGoals = &argentinaGoals
	match.OpponentGoals = &opponentGoals
	match.Status = MatchStatusResultRecorded

	if err := s.repo.UpdateMatch(ctx, match); err != nil {
		slog.ErrorContext(ctx, "error al guardar resultado", "match_id", matchID,
			"argentina_goals", req.ArgentinaGoals,
			"opponent_goals", req.OpponentGoals,
			"error", err,)
		return nil, err
	}

	slog.InfoContext(ctx, "resultado registrado", "match_id", matchID,
		"argentina_goals", req.ArgentinaGoals,
		"opponent_goals", req.OpponentGoals,)

	return adminMatchToResponse(match), nil
}

// SettleMatch evalúa todas las predicciones de un partido y asigna premios.
func (s *Service) SettleMatch(ctx context.Context, matchID uuid.UUID) (*SettlementResponse, error) {
	match, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	if match.Status != MatchStatusResultRecorded {
		return nil, ErrResultMissing
	}

	if match.ArgentinaGoals == nil || match.OpponentGoals == nil {
		return nil, ErrResultMissing
	}

	predictions, err := s.repo.GetPredictionsByMatchID(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "error al obtener predicciones para settlement", "match_id", matchID, "error", err)
		return nil, err
	}

	correctCount := 0
	incorrectCount := 0
	pendingInventory := 0
	totalPreds := len(predictions)
	needsAdminNotify := false

	for _, pred := range predictions {
		existingReward, err := s.repo.GetRewardByPredictionID(ctx, pred.ID)
		if err != nil {
			slog.ErrorContext(ctx, "error al verificar premio existente", "prediction_id", pred.ID, "error", err)
			continue
		}

		if existingReward != nil {
			if existingReward.Status == RewardStatusFulfilled {
				correctCount++
				continue
			}
			if existingReward.Status == RewardStatusPendingInventory {
				if s.tryAssignVoucher(ctx, existingReward) {
					correctCount++
				} else {
					pendingInventory++
					needsAdminNotify = true
				}
				continue
			}
			continue
		}

		exactMatch := pred.ArgentinaGoals == *match.ArgentinaGoals &&
			pred.OpponentGoals == *match.OpponentGoals

		if exactMatch {
			reward := &ProdeReward{
				PredictionID: pred.ID,
				UserID:       pred.UserID,
				Status:       RewardStatusPending,
			}

			if err := s.repo.CreateReward(ctx, reward); err != nil {
				slog.ErrorContext(ctx, "error al crear premio", "prediction_id", pred.ID, "error", err)
				continue
			}

			if s.tryAssignVoucher(ctx, reward) {
				correctCount++
			} else {
				pendingInventory++
				needsAdminNotify = true
			}

			pred.Status = PredStatusCorrect
			if err := s.repo.UpdatePrediction(ctx, &pred); err != nil {
				slog.ErrorContext(ctx, "error al actualizar predicción", "prediction_id", pred.ID, "error", err)
			}
		} else {
			incorrectCount++
			pred.Status = PredStatusIncorrect
			if err := s.repo.UpdatePrediction(ctx, &pred); err != nil {
				slog.ErrorContext(ctx, "error al actualizar predicción", "prediction_id", pred.ID, "error", err)
			}
		}
	}

	match.Status = MatchStatusEvaluated
	if err := s.repo.UpdateMatch(ctx, match); err != nil {
		slog.ErrorContext(ctx, "error al actualizar estado del partido", "match_id", matchID, "error", err)
	}

	if needsAdminNotify {
		s.notifyAdmins(ctx, match, pendingInventory)
	}

	slog.InfoContext(ctx, "settlement completado", "match_id", matchID,
		"total", totalPreds,
		"correct", correctCount,
		"incorrect", incorrectCount,
		"pending_inventory", pendingInventory,)

	return &SettlementResponse{
		MatchID:    matchID.String(),
		Status:     MatchStatusEvaluated,
		TotalPreds: totalPreds,
		Correct:    correctCount,
	}, nil
}

// RetryPendingRewards reintenta asignar vouchers a premios pendientes por inventario.
func (s *Service) RetryPendingRewards(ctx context.Context) (*RewardRetryResponse, error) {
	pending, err := s.repo.GetPendingInventoryRewards(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error al obtener premios pendientes", "error", err)
		return nil, err
	}

	processed := 0
	assigned := 0
	failed := 0

	for i := range pending {
		reward := &pending[i]
		processed++

		pred, err := s.repo.GetPredictionByID(ctx, reward.PredictionID)
		if err != nil {
			slog.ErrorContext(ctx, "error al obtener predicción", "prediction_id", reward.PredictionID, "error", err)
			reward.Status = RewardStatusFailed
			reward.FailureReason = err.Error()
			if err := s.repo.UpdateReward(ctx, reward); err != nil {
				slog.ErrorContext(ctx, "error al actualizar premio", "reward_id", reward.ID, "error", err)
			}
			failed++
			continue
		}

		if pred.Status != PredStatusCorrect {
			reward.Status = RewardStatusSkipped
			reward.FailureReason = "predicción ya no es correcta"
			if err := s.repo.UpdateReward(ctx, reward); err != nil {
				slog.ErrorContext(ctx, "error al actualizar premio", "reward_id", reward.ID, "error", err)
			}
			failed++
			continue
		}

		if s.tryAssignVoucher(ctx, reward) {
			assigned++
		} else {
			failed++
		}
	}

	remaining, _ := s.repo.CountPendingInventoryRewards(ctx)

	slog.InfoContext(ctx, "retry de premios completado", "processed", processed,
		"assigned", assigned,
		"failed", failed,
		"remaining", remaining,)

	return &RewardRetryResponse{
		Processed: processed,
		Assigned:  assigned,
		Failed:    failed,
		Remaining: remaining,
	}, nil
}

// tryAssignVoucher intenta asignar un voucher a un premio.
// Retorna true si se asignó correctamente, false si no hay inventario.
func (s *Service) tryAssignVoucher(ctx context.Context, reward *ProdeReward) bool {
	voucherEntity, err := s.voucherRepo.AssignNextVoucher(ctx, &voucher.VoucherRequest{UserID: reward.UserID})
	if err != nil {
		if err == voucher.ErrNoAvailableVouchers {
			reward.Status = RewardStatusPendingInventory
			if err := s.repo.UpdateReward(ctx, reward); err != nil {
				slog.ErrorContext(ctx, "error al actualizar premio pendiente", "reward_id", reward.ID, "error", err)
			}
			return false
		}
		slog.ErrorContext(ctx, "error al asignar voucher", "reward_id", reward.ID, "error", err)
		reward.Status = RewardStatusFailed
		reward.FailureReason = err.Error()
		if err := s.repo.UpdateReward(ctx, reward); err != nil {
			slog.ErrorContext(ctx, "error al actualizar premio", "reward_id", reward.ID, "error", err)
		}
		return false
	}

	voucherID := voucherEntity.ID
	reward.VoucherID = &voucherID
	reward.Status = RewardStatusFulfilled
	if err := s.repo.UpdateReward(ctx, reward); err != nil {
		slog.ErrorContext(ctx, "error al actualizar premio", "reward_id", reward.ID, "error", err)
		return false
	}

	if err := s.sendVoucherEmail(ctx, reward.UserID, voucherEntity); err != nil {
		slog.ErrorContext(ctx, "error al enviar email del voucher", "reward_id", reward.ID,
			"user_id", reward.UserID,
			"error", err,)
	}

	return true
}

// sendVoucherEmail envía el email del voucher al usuario.
func (s *Service) sendVoucherEmail(ctx context.Context, userID uuid.UUID, voucherEntity *voucher.Voucher) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	imageURL := fmt.Sprintf("%s/%s", os.Getenv("VOUCHER_BUCKET_URL"), voucherEntity.StoragePath)
	return s.mailer.SendVoucherEmail(ctx, user.Email, imageURL)
}

// notifyAdmins envía notificación a los administradores sobre premios pendientes.
func (s *Service) notifyAdmins(ctx context.Context, match *ProdeMatch, pendingCount int) {
	for _, email := range s.adminEmails {
		if err := s.mailer.SendProdeAdminNotification(ctx, email, match.Opponent, match.Stage, pendingCount); err != nil {
			slog.ErrorContext(ctx, "error al notificar a admin", "email", email, "error", err)
		}
	}

	// Marcar rewards pendientes como notificados
	pending, err := s.repo.GetPendingInventoryRewards(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error al obtener premios pendientes", "error", err)
		return
	}
	for i := range pending {
		pending[i].AdminNotified = true
		if err := s.repo.UpdateReward(ctx, &pending[i]); err != nil {
			slog.ErrorContext(ctx, "error al marcar premio como notificado", "reward_id", pending[i].ID, "error", err)
		}
	}
}

// helpers de conversión

func matchToResponse(m ProdeMatch) MatchResponse {
	now := time.Now()
	resp := MatchResponse{
		ID:            m.ID.String(),
		Stage:         m.Stage,
		Opponent:      m.Opponent,
		KickoffAt:     m.KickoffAt,
		CutoffAt:      m.CutoffAt(),
		Status:        m.Status,
		IsOpen:        m.IsOpenForPrediction(now),
		ArgentinaGoals: m.ArgentinaGoals,
		OpponentGoals:  m.OpponentGoals,
	}
	return resp
}

func adminMatchToResponse(m *ProdeMatch) *AdminMatchResponse {
	return &AdminMatchResponse{
		ID:             m.ID.String(),
		Stage:          m.Stage,
		Opponent:       m.Opponent,
		KickoffAt:      m.KickoffAt,
		CutoffAt:       m.CutoffAt(),
		Status:         m.Status,
		IsVisible:      m.IsVisible,
		ArgentinaGoals: m.ArgentinaGoals,
		OpponentGoals:  m.OpponentGoals,
		ExternalID:     m.ExternalID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func predictionToResponse(p *ProdePrediction) *PredictionResponse {
	return &PredictionResponse{
		ID:             p.ID.String(),
		MatchID:        p.MatchID.String(),
		ArgentinaGoals: p.ArgentinaGoals,
		OpponentGoals:  p.OpponentGoals,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
