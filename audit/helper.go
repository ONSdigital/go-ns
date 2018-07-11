package audit

import (
	"context"
	"strings"

	"github.com/ONSdigital/go-ns/common"
)

// pathIDs relate the possible string matches in path with the respective
// parameter name as key values pairs
var pathIDs = map[string]string{
	"jobs":                  "job_id",
	"datasets":              "dataset_id",
	"editions":              "edition",
	"versions":              "version",
	"dimensions":            "dimension",
	"options":               "option",
	"instances":             "instance_id",
	"inserted_observations": "inserted_observations",
	"node_id":               "node_id",
	"filters":               "filter_blueprint_id",
	"filter-outputs":        "filter_output_id",
	"hierarchies":           "instance_id",
	// TODO add params for codelist API endpoints
}

// GetParameters populates audit parameters with path variable values
func GetParameters(ctx context.Context, path string) common.Params {
	auditParams := common.Params{}

	callerIdentity := common.Caller(ctx)
	if callerIdentity != "" {
		auditParams["caller_identity"] = callerIdentity
	}

	splitPath := strings.Split(path, "/")

	for i, pathValue := range splitPath {
		if i+1 >= len(splitPath) {
			break
		}

		parameter, ok := pathIDs[pathValue]
		if ok {
			auditParams[parameter] = splitPath[i+1]
		}
	}

	return auditParams
}
