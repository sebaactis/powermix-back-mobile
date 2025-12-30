package proof

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type MockProofRepository struct {
	Proofs       map[string]*Proof
	CreateCalled bool
	ShouldFail   bool
	FailError    error
}

func NewMockProofRepository() *MockProofRepository {
	return &MockProofRepository{
		Proofs: make(map[string]*Proof),
	}
}

func (m *MockProofRepository) Create(ctx context.Context, proof *Proof) (*Proof, error) {
	m.CreateCalled = true
	if m.ShouldFail {
		return nil, m.FailError
	}
	proof.ID = uuid.New()
	m.Proofs[proof.ID_MP] = proof
	return proof, nil
}

func (m *MockProofRepository) GetById(ctx context.Context, id string) (*Proof, error) {
	if proof, ok := m.Proofs[id]; ok {
		return proof, nil
	}
	return nil, nil
}

func (m *MockProofRepository) Rollback() {
	// Simula rollback limpiando los proofs creados en esta "transaccion"
	m.Proofs = make(map[string]*Proof)
}

// MockUserRepository simula el repositorio de usuarios
type MockUserRepository struct {
	Users                 map[uuid.UUID]*MockUser
	IncrementCalled       bool
	ResetCalled           bool
	ShouldFailIncrement   bool
	ShouldFailReset       bool
	FailError             error
	lastIncrementedStamps int
	pendingIncrements     map[uuid.UUID]int // Para simular transacciones
	pendingResets         map[uuid.UUID]bool
}

type MockUser struct {
	ID            uuid.UUID
	StampsCounter int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:             make(map[uuid.UUID]*MockUser),
		pendingIncrements: make(map[uuid.UUID]int),
		pendingResets:     make(map[uuid.UUID]bool),
	}
}

func (m *MockUserRepository) IncrementStampsCounter(ctx context.Context, userID uuid.UUID) (int, error) {
	m.IncrementCalled = true
	if m.ShouldFailIncrement {
		return 0, m.FailError
	}

	user, ok := m.Users[userID]
	if !ok {
		return 0, errors.New("user not found")
	}

	// Guardar como pendiente (no commitear aun)
	newCount := user.StampsCounter + 1
	m.pendingIncrements[userID] = newCount
	m.lastIncrementedStamps = newCount
	return newCount, nil
}

func (m *MockUserRepository) ResetStampsCounter(ctx context.Context, userID uuid.UUID) (int, error) {
	m.ResetCalled = true
	if m.ShouldFailReset {
		return 0, m.FailError
	}

	m.pendingResets[userID] = true
	return 0, nil
}

func (m *MockUserRepository) Commit() {
	// Aplica los cambios pendientes
	for userID, stamps := range m.pendingIncrements {
		if m.Users[userID] != nil {
			m.Users[userID].StampsCounter = stamps
		}
	}
	for userID := range m.pendingResets {
		if m.Users[userID] != nil {
			m.Users[userID].StampsCounter = 0
		}
	}
	m.pendingIncrements = make(map[uuid.UUID]int)
	m.pendingResets = make(map[uuid.UUID]bool)
}

func (m *MockUserRepository) Rollback() {
	// Descarta los cambios pendientes
	m.pendingIncrements = make(map[uuid.UUID]int)
	m.pendingResets = make(map[uuid.UUID]bool)
}

// MockVoucherRepository simula el repositorio de vouchers
type MockVoucherRepository struct {
	Vouchers           map[uuid.UUID]*MockVoucher
	AssignCalled       bool
	ShouldFailAssign   bool
	FailError          error
	pendingAssignments map[uuid.UUID]uuid.UUID // voucherID -> userID
}

type MockVoucher struct {
	ID         uuid.UUID
	IsAssigned bool
	UserID     uuid.UUID
}

func NewMockVoucherRepository() *MockVoucherRepository {
	return &MockVoucherRepository{
		Vouchers:           make(map[uuid.UUID]*MockVoucher),
		pendingAssignments: make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *MockVoucherRepository) AssignNextVoucher(ctx context.Context, userID uuid.UUID) (*MockVoucher, error) {
	m.AssignCalled = true
	if m.ShouldFailAssign {
		return nil, m.FailError
	}

	// Buscar un voucher disponible
	for id, v := range m.Vouchers {
		if !v.IsAssigned {
			m.pendingAssignments[id] = userID
			return v, nil
		}
	}

	return nil, voucher.ErrNoAvailableVouchers
}

func (m *MockVoucherRepository) Commit() {
	for voucherID, userID := range m.pendingAssignments {
		if m.Vouchers[voucherID] != nil {
			m.Vouchers[voucherID].IsAssigned = true
			m.Vouchers[voucherID].UserID = userID
		}
	}
	m.pendingAssignments = make(map[uuid.UUID]uuid.UUID)
}

func (m *MockVoucherRepository) Rollback() {
	m.pendingAssignments = make(map[uuid.UUID]uuid.UUID)
}

// ============================================================================
// TRANSACTION SIMULATOR - Simula el comportamiento de una transaccion de DB
// ============================================================================

type TransactionSimulator struct {
	proofRepo   *MockProofRepository
	userRepo    *MockUserRepository
	voucherRepo *MockVoucherRepository
}

func NewTransactionSimulator(p *MockProofRepository, u *MockUserRepository, v *MockVoucherRepository) *TransactionSimulator {
	return &TransactionSimulator{
		proofRepo:   p,
		userRepo:    u,
		voucherRepo: v,
	}
}

func (t *TransactionSimulator) Execute(fn func() error) error {
	err := fn()
	if err != nil {
		// ROLLBACK - descartar todos los cambios
		t.proofRepo.Rollback()
		t.userRepo.Rollback()
		t.voucherRepo.Rollback()
		return err
	}
	// COMMIT - aplicar todos los cambios
	t.userRepo.Commit()
	t.voucherRepo.Commit()
	return nil
}

// ============================================================================
// TESTS
// ============================================================================

// TestAtomicity_RollbackWhenNoVouchersAvailable verifica que cuando no hay
// vouchers disponibles y el usuario llega a 5 stamps, todo hace rollback
func TestAtomicity_RollbackWhenNoVouchersAvailable(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	proofRepo := NewMockProofRepository()
	userRepo := NewMockUserRepository()
	voucherRepo := NewMockVoucherRepository()

	// Crear usuario con 4 stamps
	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 4,
	}

	// NO crear vouchers disponibles - esto causara el error

	txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

	// Simular la logica de Create
	newProof := &Proof{
		UserID:            userID,
		ID_MP:             "TEST-123",
		Date_Approved_MP:  utils.FormattedTime{Time: time.Now()},
		Operation_Type_MP: "regular_payment",
		Status_MP:         "approved",
		Amount_MP:         1000.00,
		ProofDate:         utils.NowFormatted(),
	}

	err := txSimulator.Execute(func() error {
		// 1. Crear proof
		_, createErr := proofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		// 2. Incrementar stamps
		stamps, createErr := userRepo.IncrementStampsCounter(ctx, userID)
		if createErr != nil {
			return createErr
		}

		t.Logf("Stamps despues de incrementar: %d", stamps)

		// 3. Si llega a 5, asignar voucher
		if stamps == 5 {
			_, createErr = voucherRepo.AssignNextVoucher(ctx, userID)
			if createErr != nil {
				t.Logf("Error al asignar voucher (esperado): %v", createErr)
				return createErr // Esto dispara el rollback
			}

			_, createErr = userRepo.ResetStampsCounter(ctx, userID)
			if createErr != nil {
				return createErr
			}
		}

		return nil
	})

	// VERIFICACIONES

	// 1. Debe haber error
	if err == nil {
		t.Fatal("Se esperaba error cuando no hay vouchers disponibles")
	}

	if err != voucher.ErrNoAvailableVouchers {
		t.Errorf("Se esperaba ErrNoAvailableVouchers, se obtuvo: %v", err)
	}

	// 2. El proof NO debe existir (rollback)
	if len(proofRepo.Proofs) != 0 {
		t.Errorf("Se esperaban 0 proofs despues del rollback, hay %d", len(proofRepo.Proofs))
	}

	// 3. El contador de stamps debe seguir en 4 (rollback)
	if userRepo.Users[userID].StampsCounter != 4 {
		t.Errorf("Se esperaba stamps en 4 despues del rollback, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	// 4. El voucher NO debe estar asignado
	for _, v := range voucherRepo.Vouchers {
		if v.IsAssigned {
			t.Error("No deberia haber vouchers asignados")
		}
	}

	t.Log("TEST PASSED: Atomicidad verificada - rollback funciona correctamente")
}

// TestAtomicity_SuccessWhenVouchersAvailable verifica el happy path
func TestAtomicity_SuccessWhenVouchersAvailable(t *testing.T) {
	ctx := context.Background()

	proofRepo := NewMockProofRepository()
	userRepo := NewMockUserRepository()
	voucherRepo := NewMockVoucherRepository()

	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 4,
	}

	// Crear un voucher disponible
	voucherID := uuid.New()
	voucherRepo.Vouchers[voucherID] = &MockVoucher{
		ID:         voucherID,
		IsAssigned: false,
	}

	txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

	newProof := &Proof{
		UserID:            userID,
		ID_MP:             "TEST-456",
		Date_Approved_MP:  utils.FormattedTime{Time: time.Now()},
		Operation_Type_MP: "regular_payment",
		Status_MP:         "approved",
		Amount_MP:         1500.00,
		ProofDate:         utils.NowFormatted(),
	}

	err := txSimulator.Execute(func() error {
		_, createErr := proofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		stamps, createErr := userRepo.IncrementStampsCounter(ctx, userID)
		if createErr != nil {
			return createErr
		}

		t.Logf("Stamps despues de incrementar: %d", stamps)

		if stamps == 5 {
			_, createErr = voucherRepo.AssignNextVoucher(ctx, userID)
			if createErr != nil {
				return createErr
			}

			_, createErr = userRepo.ResetStampsCounter(ctx, userID)
			if createErr != nil {
				return createErr
			}
		}

		return nil
	})

	// VERIFICACIONES

	// 1. No debe haber error
	if err != nil {
		t.Fatalf("No se esperaba error, se obtuvo: %v", err)
	}

	// 2. El proof debe existir
	if len(proofRepo.Proofs) != 1 {
		t.Errorf("Se esperaba 1 proof, hay %d", len(proofRepo.Proofs))
	}

	// 3. El contador debe estar en 0 (reseteado)
	if userRepo.Users[userID].StampsCounter != 0 {
		t.Errorf("Se esperaba stamps en 0, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	// 4. El voucher debe estar asignado
	if !voucherRepo.Vouchers[voucherID].IsAssigned {
		t.Error("El voucher deberia estar asignado")
	}
	if voucherRepo.Vouchers[voucherID].UserID != userID {
		t.Errorf("Voucher asignado al usuario incorrecto")
	}

	t.Log("TEST PASSED: Happy path funciona correctamente")
}

// TestAtomicity_RollbackWhenProofCreationFails verifica rollback cuando falla crear proof
func TestAtomicity_RollbackWhenProofCreationFails(t *testing.T) {
	ctx := context.Background()

	proofRepo := NewMockProofRepository()
	proofRepo.ShouldFail = true
	proofRepo.FailError = errors.New("database connection error")

	userRepo := NewMockUserRepository()
	voucherRepo := NewMockVoucherRepository()

	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 4,
	}

	txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

	newProof := &Proof{
		UserID: userID,
		ID_MP:  "TEST-FAIL",
	}

	err := txSimulator.Execute(func() error {
		_, createErr := proofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		// Esto no deberia ejecutarse
		_, _ = userRepo.IncrementStampsCounter(ctx, userID)
		return nil
	})

	// VERIFICACIONES

	if err == nil {
		t.Fatal("Se esperaba error al crear proof")
	}

	// El contador no debe haberse modificado
	if userRepo.IncrementCalled {
		t.Error("IncrementStampsCounter no deberia haberse llamado")
	}

	if userRepo.Users[userID].StampsCounter != 4 {
		t.Errorf("Stamps deberia seguir en 4, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	t.Log("TEST PASSED: Rollback cuando falla crear proof")
}

// TestAtomicity_RollbackWhenIncrementFails verifica rollback cuando falla incrementar
func TestAtomicity_RollbackWhenIncrementFails(t *testing.T) {
	ctx := context.Background()

	proofRepo := NewMockProofRepository()
	userRepo := NewMockUserRepository()
	userRepo.ShouldFailIncrement = true
	userRepo.FailError = errors.New("database error on increment")

	voucherRepo := NewMockVoucherRepository()

	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 4,
	}

	txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

	newProof := &Proof{
		UserID: userID,
		ID_MP:  "TEST-INCREMENT-FAIL",
	}

	err := txSimulator.Execute(func() error {
		_, createErr := proofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		_, createErr = userRepo.IncrementStampsCounter(ctx, userID)
		if createErr != nil {
			return createErr
		}

		return nil
	})

	// VERIFICACIONES

	if err == nil {
		t.Fatal("Se esperaba error al incrementar stamps")
	}

	// El proof fue creado pero debe haber hecho rollback
	if len(proofRepo.Proofs) != 0 {
		t.Errorf("Proof deberia haberse eliminado en rollback, hay %d", len(proofRepo.Proofs))
	}

	// El contador no debe haberse modificado
	if userRepo.Users[userID].StampsCounter != 4 {
		t.Errorf("Stamps deberia seguir en 4, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	t.Log("TEST PASSED: Rollback cuando falla incrementar stamps")
}

// TestAtomicity_NoVoucherAssignmentBelow5Stamps verifica que no se asigna voucher con menos de 5 stamps
func TestAtomicity_NoVoucherAssignmentBelow5Stamps(t *testing.T) {
	ctx := context.Background()

	proofRepo := NewMockProofRepository()
	userRepo := NewMockUserRepository()
	voucherRepo := NewMockVoucherRepository()

	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 2, // Solo 2 stamps, llegara a 3
	}

	// NO importa si hay vouchers o no, no deberia intentar asignar
	txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

	newProof := &Proof{
		UserID: userID,
		ID_MP:  "TEST-BELOW-5",
	}

	err := txSimulator.Execute(func() error {
		_, createErr := proofRepo.Create(ctx, newProof)
		if createErr != nil {
			return createErr
		}

		stamps, createErr := userRepo.IncrementStampsCounter(ctx, userID)
		if createErr != nil {
			return createErr
		}

		t.Logf("Stamps: %d (deberia ser 3)", stamps)

		if stamps == 5 {
			t.Error("No deberia llegar a 5 stamps en este test")
		}

		return nil
	})

	// VERIFICACIONES

	if err != nil {
		t.Fatalf("No se esperaba error: %v", err)
	}

	// El proof debe existir
	if len(proofRepo.Proofs) != 1 {
		t.Errorf("Se esperaba 1 proof, hay %d", len(proofRepo.Proofs))
	}

	// El contador debe estar en 3
	if userRepo.Users[userID].StampsCounter != 3 {
		t.Errorf("Se esperaba stamps en 3, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	// No se debe haber intentado asignar voucher
	if voucherRepo.AssignCalled {
		t.Error("No deberia haberse llamado AssignNextVoucher")
	}

	t.Log("TEST PASSED: No se asigna voucher cuando stamps < 5")
}

// TestAtomicity_MultipleProofsToReach5 simula cargar varios proofs hasta llegar a 5
func TestAtomicity_MultipleProofsToReach5(t *testing.T) {
	ctx := context.Background()

	proofRepo := NewMockProofRepository()
	userRepo := NewMockUserRepository()
	voucherRepo := NewMockVoucherRepository()

	userID := uuid.New()
	userRepo.Users[userID] = &MockUser{
		ID:            userID,
		StampsCounter: 0,
	}

	// Crear un voucher disponible
	voucherID := uuid.New()
	voucherRepo.Vouchers[voucherID] = &MockVoucher{
		ID:         voucherID,
		IsAssigned: false,
	}

	// Simular 5 cargas de proof
	for i := 1; i <= 5; i++ {
		txSimulator := NewTransactionSimulator(proofRepo, userRepo, voucherRepo)

		newProof := &Proof{
			UserID: userID,
			ID_MP:  "TEST-MULTI-" + string(rune('0'+i)),
		}

		err := txSimulator.Execute(func() error {
			_, createErr := proofRepo.Create(ctx, newProof)
			if createErr != nil {
				return createErr
			}

			stamps, createErr := userRepo.IncrementStampsCounter(ctx, userID)
			if createErr != nil {
				return createErr
			}

			t.Logf("Proof %d: stamps = %d", i, stamps)

			if stamps == 5 {
				_, createErr = voucherRepo.AssignNextVoucher(ctx, userID)
				if createErr != nil {
					return createErr
				}

				_, createErr = userRepo.ResetStampsCounter(ctx, userID)
				if createErr != nil {
					return createErr
				}
				t.Log("Voucher asignado y contador reseteado!")
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Error en proof %d: %v", i, err)
		}
	}

	// VERIFICACIONES FINALES

	// 5 proofs creados
	if len(proofRepo.Proofs) != 5 {
		t.Errorf("Se esperaban 5 proofs, hay %d", len(proofRepo.Proofs))
	}

	// Contador reseteado a 0
	if userRepo.Users[userID].StampsCounter != 0 {
		t.Errorf("Se esperaba stamps en 0 despues del ciclo, esta en %d", userRepo.Users[userID].StampsCounter)
	}

	// Voucher asignado
	if !voucherRepo.Vouchers[voucherID].IsAssigned {
		t.Error("El voucher deberia estar asignado")
	}

	t.Log("TEST PASSED: Ciclo completo de 5 proofs funciona correctamente")
}
