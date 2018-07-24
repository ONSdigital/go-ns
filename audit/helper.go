package audit

import (
	"context"
	"strings"

	"github.com/ONSdigital/go-ns/common"
)

// pathIDs relate the possible string matches in path with the respective
// parameter name as key values pairs
var pathIDs = map[string]string{
	"jobs":      "job_id",
	"datasets":  "dataset_id",
	"instances": "instance_id",
}

// GetParameters populates audit parameters with path variable values
func GetParameters(ctx context.Context, path string, vars map[string]string) common.Params {
	auditParams := common.Params{}

	callerIdentity := common.Caller(ctx)
	if callerIdentity != "" {
		auditParams["caller_identity"] = callerIdentity
	}

	pathSegments := strings.Split(path, "/")
	// Remove initial segment if empty
	if pathSegments[0] == "" {
		pathSegments = pathSegments[1:]
	}
	numberOfSegments := len(pathSegments)

	if pathSegments[0] == "hierarchies" {
		if numberOfSegments > 1 {
			auditParams["instance_id"] = pathSegments[1]

			if numberOfSegments > 2 {
				auditParams["dimension"] = pathSegments[2]

				if numberOfSegments > 3 {
					auditParams["code"] = pathSegments[3]
				}
			}
		}

		return auditParams
	}

	if pathSegments[0] == "search" {
		pathSegments = pathSegments[1:]
	}

	for key, value := range vars {
		if key == "id" {
			auditParams[pathIDs[pathSegments[0]]] = value
		} else {
			auditParams[key] = value
		}
	}

	return auditParams
}
