package api

import (
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

func (lm *loggingMiddleware) IssueCert(token string) (res certs.Cert, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_cert for token: %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.IssueCert(token)
}
