package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/certs"
	log "github.com/mainflux/mainflux/logger"
)

var _ certs.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    certs.Service
}

// NewLoggingMiddleware adds logging facilities to the core service.
func NewLoggingMiddleware(svc certs.Service, logger log.Logger) certs.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) IssueCert(ctx context.Context, token, thingID, daysValid string, keyBits int, keyType string) (c certs.Cert, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_cert for token: %s and thing: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.IssueCert(ctx, token, thingID, daysValid, keyBits, keyType)
}

func (lm *loggingMiddleware) ListCertificates(ctx context.Context, token, thingID string, offset, limit uint64) (cp certs.CertsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_certificates for token: %s and thing: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListCertificates(ctx, token, thingID, offset, limit)
}

func (lm *loggingMiddleware) RevokeCert(ctx context.Context, token, thingID, certSerial string) (c certs.Revoke, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_certificates for token: %s and thing: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RevokeCert(ctx, token, thingID, certSerial)
}
