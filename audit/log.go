package audit

import (
	"context"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// LogError creates a structured error message when auditing fails
func LogError(ctx context.Context, err error, data log.Data) {
	if data == nil {
		data = log.Data{}
	}

	if user := common.User(ctx); user != "" {
		data["user"] = user
	}

	if caller := common.Caller(ctx); caller != "" {
		data["caller"] = caller
	}

	log.ErrorCtx(ctx, err, data)
}

// LogInfo creates a structured info message when auditing succeeds
func LogInfo(ctx context.Context, message string, data log.Data) {
	if data == nil {
		data = log.Data{}
	}

	if user := common.User(ctx); user != "" {
		data["user"] = user
	}

	if caller := common.Caller(ctx); caller != "" {
		data["caller"] = caller
	}

	log.InfoCtx(ctx, message, data)
}
