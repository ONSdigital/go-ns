package audit

import (
	"context"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetParameters(t *testing.T) {

	Convey("Given a context without a caller identity and any path parameters", t, func() {
		ctx := context.Background()
		path := "/"

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path)

			Convey("Then the audit parameters are empty", func() {
				So(auditParams, ShouldResemble, common.Params{})
			})
		})
	})

	Convey("Given a context with a caller identity and no path parameters", t, func() {
		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "gerald")
		path := "/"

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path)

			Convey("Then the audit parameters should contain 'caller_identity'", func() {
				So(auditParams, ShouldResemble, common.Params{"caller_identity": "gerald"})
			})
		})
	})

	Convey("Given a context without a caller identity but with path parameters", t, func() {
		ctx := context.Background()

		Convey("of job id", func() {
			path := "/jobs/123"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'job_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"job_id": "123"})
				})
			})
		})

		Convey("of dataset id", func() {
			path := "/datasets/234"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'dataset_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"dataset_id": "234"})
				})
			})
		})

		Convey("of edition", func() {
			path := "/editions/2017"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'edition'", func() {
					So(auditParams, ShouldResemble, common.Params{"edition": "2017"})
				})
			})
		})

		Convey("of version", func() {
			path := "/versions/1"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'version'", func() {
					So(auditParams, ShouldResemble, common.Params{"version": "1"})
				})
			})
		})

		Convey("of dimension", func() {
			path := "/dimensions/age"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'dimension'", func() {
					So(auditParams, ShouldResemble, common.Params{"dimension": "age"})
				})
			})
		})

		Convey("of option", func() {
			path := "/options/23"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'option'", func() {
					So(auditParams, ShouldResemble, common.Params{"option": "23"})
				})
			})
		})

		Convey("of instance", func() {
			path := "/instances/246"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'instance'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "246"})
				})
			})
		})

		Convey("of inserted observations", func() {
			path := "/inserted_observations/500"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'inserted_observations'", func() {
					So(auditParams, ShouldResemble, common.Params{"inserted_observations": "500"})
				})
			})
		})

		Convey("of node id", func() {
			path := "/node_id/543"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'node_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"node_id": "543"})
				})
			})
		})

		Convey("of filter blueprint id", func() {
			path := "/filters/789"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'filter_blueprint_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"filter_blueprint_id": "789"})
				})
			})
		})

		Convey("of filter output id", func() {
			path := "/filter-outputs/111"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'filter_output_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"filter_output_id": "111"})
				})
			})
		})

		Convey("of a hierarchy instance id", func() {
			path := "/hierarchies/965"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'instance_id'", func() {
					So(auditParams, ShouldResemble, common.Params{"instance_id": "965"})
				})
			})
		})

		Convey("of dataset id and edition", func() {
			path := "/datasets/234/editions/2017"

			Convey("When GetParameters is called with the context and path", func() {
				auditParams := GetParameters(ctx, path)

				Convey("Then the audit parameters should contain 'dataset_id' and 'edition'", func() {
					So(auditParams, ShouldResemble, common.Params{"dataset_id": "234", "edition": "2017"})
				})
			})
		})
	})

	Convey("Given a context with a caller identity and path parameters for a unique option against a dataset", t, func() {
		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "harold")
		path := "/datasets/999/editions/2018/versions/3/dimensions/gender/options/male"

		Convey("When GetParameters is called with the context and path", func() {
			auditParams := GetParameters(ctx, path)

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
