package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type ec2Provider struct{}

func init() {
	Register(&ec2Provider{})
}

func (p *ec2Provider) ResourceType() string {
	return "ec2:instance"
}

func (p *ec2Provider) List(ctx context.Context, cfg aws.Config) ([]Resource, error) {
	accountID, err := getAccountID(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get account ID: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	now := time.Now()

	// Filter out terminated instances
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "shutting-down", "stopping", "stopped"},
			},
		},
	}

	var resources []Resource
	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				r := Resource{
					Type:      "ec2:instance",
					ID:        aws.ToString(inst.InstanceId),
					Region:    cfg.Region,
					AccountID: accountID,
					Tags:      tagsToMap(inst.Tags),
					Properties: map[string]string{
						"instance_type": string(inst.InstanceType),
						"state":         string(inst.State.Name),
						"vpc_id":        aws.ToString(inst.VpcId),
						"subnet_id":     aws.ToString(inst.SubnetId),
						"public_ip":     aws.ToString(inst.PublicIpAddress),
						"private_ip":    aws.ToString(inst.PrivateIpAddress),
						"ami_id":        aws.ToString(inst.ImageId),
						"architecture":  string(inst.Architecture),
					},
					FetchedAt: now,
				}

				if inst.LaunchTime != nil {
					r.Properties["launch_time"] = inst.LaunchTime.Format(time.RFC3339)
				}

				if inst.PlatformDetails != nil {
					r.Properties["platform"] = aws.ToString(inst.PlatformDetails)
				}

				resources = append(resources, r)
			}
		}
	}

	return resources, nil
}

func tagsToMap(tags []types.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return m
}

func getAccountID(ctx context.Context, cfg aws.Config) (string, error) {
	client := sts.NewFromConfig(cfg)
	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.ToString(output.Account), nil
}
