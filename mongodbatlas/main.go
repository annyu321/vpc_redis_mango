package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/redis"
	"github.com/pulumi/pulumi-mongodbatlas/sdk/v2/go/mongodbatlas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		//Deploy redis
		_, err := redis.NewInstance(ctx, "cache", &redis.InstanceArgs{
			MemorySizeGb: pulumi.Int(1),
		})
		if err != nil {
			return err
		}

		//Deploy mongodb
		_, err1 := mongodbatlas.NewProject(ctx, "demo", &mongodbatlas.ProjectArgs{
			OrgId: pulumi.String("12345"),
		})
		if err1 != nil {
			return err1
		}

		return nil
	})
}
