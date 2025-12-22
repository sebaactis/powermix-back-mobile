package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
)

type VoucherCron struct {
	vouchers  *voucher.Service
	spec      string
	batchSize int
	timeout   time.Duration

	mu      sync.Mutex
	running bool

	cron *cron.Cron
}

func NewVoucherCron(vouchers *voucher.Service, spec string, batchSize int, timeout time.Duration) *VoucherCron {
	if spec == "" {
		spec = "@every 5m"
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &VoucherCron{
		vouchers:  vouchers,
		spec:      spec,
		batchSize: batchSize,
		timeout:   timeout,
	}
}

func (vc *VoucherCron) Start() error {
	vc.cron = cron.New()

	_, err := vc.cron.AddFunc(vc.spec, func() {
		vc.runOnce()
	})

	if err != nil {
		return err
	}

	vc.cron.Start()
	log.Printf("[cron] voucher job started spec=%s batch=%d timeout=%s", vc.spec, vc.batchSize, vc.timeout)
	return nil
}

func (vc *VoucherCron) Stop() {
	if vc.cron != nil {
		ctx := vc.cron.Stop()
		<-ctx.Done()
		log.Printf("[cron] voucher job stopped")
	}
}

func (vc *VoucherCron) runOnce() {
	vc.mu.Lock()

	if vc.running {
		vc.mu.Unlock()
		log.Printf("[cron] voucher job skipped (previous run still running)")
		return
	}

	vc.running = true
	vc.mu.Unlock()

	defer func() {
		vc.mu.Lock()
		vc.running = false
		vc.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), vc.timeout)
	defer cancel()

	if err := vc.vouchers.CheckUsedVouchers(ctx, vc.batchSize); err != nil {
		log.Printf("[cron] voucher job failed: %v", err)
		return
	}
}
