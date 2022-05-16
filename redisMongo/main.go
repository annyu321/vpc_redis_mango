package main

import (
	"log"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/elasticache"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a vpc security group.
		group, err := ec2.NewSecurityGroup(ctx, "vpc-secgrp", &ec2.SecurityGroupArgs{
			VpcId: pulumi.String("vpc-022497bab7cbfcb72"),
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
				},
			},

			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("TLS from VPC"),

					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(443),
					ToPort:     pulumi.Int(443),
					CidrBlocks: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
				},
				ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("Aloow inbound TCP 22"),

					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
				},
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(80),
					ToPort:     pulumi.Int(80),
					CidrBlocks: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
				},
			},
		})
		if err != nil {
			return err
		}

		// Create a private VPC for the Redis and MangoDB
		vpcArgs := &ec2.VpcArgs{
			Tags:            pulumi.Map{"Name": pulumi.String("redis-mangodb-test")},
			CidrBlock:       pulumi.String("10.0.0.0/16"),
			InstanceTenancy: pulumi.String("default"),
		}

		vpc, err := ec2.NewVpc(ctx, "redisMangoDB", vpcArgs)
		if err != nil {
			log.Printf("error creating VPC: %s", err.Error())
			return err
		}

		//Create a subnet
		subnetArgs := &ec2.SubnetArgs{
			Tags:             pulumi.Map{"Name": pulumi.String("redis-mangodb")},
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("10.0.3.0/24"),
			AvailabilityZone: pulumi.String("ca-central-1a"),
		}

		subnet, err := ec2.NewSubnet(ctx, "redisMangoDB", subnetArgs)
		if err != nil {
			log.Printf("error creating VPC Subnet: %s", err.Error())
		}

		// ElastiCache Subnet Groups are only for use when working with an ElastiCache cluster inside of a VPC.
		_, err = elasticache.NewSubnetGroup(ctx, "vpcs", &elasticache.SubnetGroupArgs{
			SubnetIds: pulumi.StringArray{
				subnet.ID(),
			},
		})
		if err != nil {
			return err
		}

		// Create a ec2 server for Mangodb installation.
		srv, err := ec2.NewInstance(ctx, "DB-Server", &ec2.InstanceArgs{
			Tags:                pulumi.Map{"Name": pulumi.String("DB-Server")},
			InstanceType:        pulumi.String("t2.micro"),
			VpcSecurityGroupIds: pulumi.StringArray{group.ID()},
			Ami:                 pulumi.String("ami-0843f7c45354d48b5"),
			SubnetId:            subnet.ID(),
		})

		if err != nil {
			return err
		}

		//Provides an ElastiCache Cluster resource, which manages Redis instance.
		//Putting them in a virtual private cloud (VPC).
		//Reference for migrating an EC2-Classic Redis cluster into a VPC:
		//https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Migrating-ec2-classic_to_VPC.html
		/*Execution Failure
				Diagnostics:
		  		aws:elasticache:Cluster (redis-mangodb):
		    	error: 1 error occurred:
		    	* error creating ElastiCache Cache Cluster: InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
		    	status code: 400, request id: 8fd6f7d3-1cb6-43cc-9c8b-3bb90ec6a353
		*/

		//Redis instance is created in the default VPC currently
		_, err1 := elasticache.NewCluster(ctx, "redis-mangodb", &elasticache.ClusterArgs{
			Engine:             pulumi.String("redis"),
			EngineVersion:      pulumi.String("3.2.10"),
			NodeType:           pulumi.String("cache.m4.large"),
			NumCacheNodes:      pulumi.Int(1),
			ParameterGroupName: pulumi.String("default.redis3.2"),
			Port:               pulumi.Int(6379),
			//SecurityGroupNames: pulumi.StringArray{
			//	group.Name,
			//},

		})
		if err1 != nil {
			return err1
		}

		// Export IDs of the created resources to the Pulumi stack
		ctx.Export("VPC-ID", vpc.ID())
		ctx.Export("Subnet-ID", subnet.ID())
		ctx.Export("Server IP", srv.PrivateIp)
		ctx.Export("Security-Group-ID", group.ID())

		return nil
	})
}
