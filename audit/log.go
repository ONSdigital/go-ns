package audit

import (
	"context"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

const (
	reqUser   = "req_user"
	reqCaller = "req_caller"
)

// LogActionFailure adds auditting data to log.Data before calling LogError
func LogActionFailure(ctx context.Context, auditedAction string, auditedResult string, err error, logData log.Data) {
	if logData == nil {
		logData = log.Data{}
	}

	logData["auditAction"] = auditedAction
	logData["auditResult"] = auditedResult

	LogError(ctx, errors.WithMessage(err, AuditActionErr), logData)
}

// LogError creates a structured error message when auditing fails
func LogError(ctx context.Context, err error, data log.Data) {
	data = addLogData(ctx, data)

	log.ErrorCtx(ctx, err, data)
}

// LogInfo creates a structured info message when auditing succeeds
func LogInfo(ctx context.Context, message string, data log.Data) {
	data = addLogData(ctx, data)

	log.InfoCtx(ctx, message, data)
}

func addLogData(ctx context.Context, data log.Data) log.Data {
	if data == nil {
		data = log.Data{}
	}

	if user := common.User(ctx); user != "" {
		data[reqUser] = user
	}

	if caller := common.Caller(ctx); caller != "" {
		data[reqCaller] = caller
	}

	return data
}
