package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type sgProvider struct{}

func init() {
	Register(&sgProvider{})
}

func (p *sgProvider) ResourceType() string {
	return "ec2:security-group"
}

func (p *sgProvider) List(ctx context.Context, cfg aws.Config) ([]Resource, error) {
	accountID, err := getAccountID(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get account ID: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	now := time.Now()

	var resources []Resource
	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe security groups: %w", err)
		}

		for _, sg := range page.SecurityGroups {
			r := Resource{
				Type:      "ec2:security-group",
				ID:        aws.ToString(sg.GroupId),
				Region:    cfg.Region,
				AccountID: accountID,
				Tags:      tagsToMap(sg.Tags),
				Properties: map[string]string{
					"group_name":          aws.ToString(sg.GroupName),
					"vpc_id":              aws.ToString(sg.VpcId),
					"description":         aws.ToString(sg.Description),
					"ingress_rules_count": strconv.Itoa(len(sg.IpPermissions)),
					"egress_rules_count":  strconv.Itoa(len(sg.IpPermissionsEgress)),
				},
				FetchedAt: now,
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}
