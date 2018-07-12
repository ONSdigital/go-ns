package audit

import (
	"context"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetParameters(t *testing.T) {

	Convey("Given a context without a caller identity and no path variables", t, func() {
		ctx := context.Background()
		path := "/jobs"

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path, nil)

			Convey("Then the audit parameters are empty", func() {
				So(auditParams, ShouldResemble, common.Params{})
			})
		})
	})

	Convey("Given a context with a caller identity and no path variables", t, func() {
		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "gerald")
		path := "/jobs"

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path, nil)

			Convey("Then the audit parameters should contain 'caller_identity'", func() {
				So(auditParams, ShouldResemble, common.Params{"caller_identity": "gerald"})
			})
		})
	})

	Convey("Given a context without a caller identity but with path variables", t, func() {
		ctx := context.Background()

		Convey("of job 'id'", func() {
			path := "/jobs/123"
			vars := map[string]string{
				"id": "123",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'job_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"job_id": "123"})
				})
			})
		})

		Convey("of dataset 'id'", func() {
			path := "/datasets/234"
			vars := map[string]string{
				"id": "234",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'dataset_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"dataset_id": "234"})
				})
			})
		})

		Convey("of instance 'id'", func() {
			path := "/instances/345"
			vars := map[string]string{
				"id": "345",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'instance_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "345"})
				})
			})
		})

		Convey("of multiple path variables", func() {
			path := "/datasets/123/editions/2017"
			vars := map[string]string{
				"id":      "123",
				"edition": "2017",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'dataset_id' and 'edition'", func() {
					So(auditParams, ShouldResemble, common.Params{"dataset_id": "123", "edition": "2017"})
				})
			})
		})

		Convey("of filter blueprint id", func() {
			path := "/filters/456"
			vars := map[string]string{
				"filter_blueprint_id": "456",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'filter_blueprint_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"filter_blueprint_id": "456"})
				})
			})
		})

		Convey("of a hierarchy instance id", func() {
			path := "/hierarchies/345"
			vars := map[string]string{
				"instance": "345",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'instance_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "345"})
				})
			})
		})

		Convey("of a hierarchy instance id and dimension", func() {
			path := "/hierarchies/345/age"
			vars := map[string]string{
				"instance":  "345",
				"dimension": "age",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'instance_id' and 'dimension'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "345", "dimension": "age"})
				})
			})
		})

		Convey("of a hierarchy instance id, dimension and code", func() {
			path := "/hierarchies/345/geography/K0100000"
			vars := map[string]string{
				"instance":  "345",
				"dimension": "geography",
				"code":      "K0100000",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'instance_id', 'dimension' and 'code'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "345", "dimension": "geography", "code": "K0100000"})
				})
			})
		})

		Convey("of a search dataset id, edition and version", func() {
			path := "/search/datasets/234/editions/2017/version/2"
			vars := map[string]string{
				"id":      "234",
				"edition": "2017",
				"version": "2",
			}

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path, vars)

				Convey("Then the audit parameters should contain 'dataset_id', 'edition' and 'version'", func() {
					So(auditParams, ShouldResemble, common.Params{"dataset_id": "234", "edition": "2017", "version": "2"})
				})
			})
		})
	})

	Convey("Given a context with a caller identity and path parameters for a unique option against a dataset", t, func() {
		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "harold")
		path := "/datasets/999/editions/2018/versions/3/dimensions/gender/options/male"
		vars := map[string]string{
			"id":        "999",
			"edition":   "2018",
			"version":   "3",
			"dimension": "gender",
			"option":    "male",
		}

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path, vars)

			Convey("Then the audit parameters should contain a list of key value pairs", func() {
				So(auditParams, ShouldResemble, common.Params{
					"caller_identity": "harold",
					"dataset_id":      "999",
					"edition":         "2018",
					"version":         "3",
					"dimension":       "gender",
					"option":          "male",
				})
			})
		})
	})
}
