package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
			props := map[string]string{
				"group_name":          aws.ToString(sg.GroupName),
				"vpc_id":              aws.ToString(sg.VpcId),
				"description":         aws.ToString(sg.Description),
				"ingress_rules_count": strconv.Itoa(len(sg.IpPermissions)),
				"egress_rules_count":  strconv.Itoa(len(sg.IpPermissionsEgress)),
			}

			for i, perm := range sg.IpPermissions {
				flattenPermission(props, "ingress", i, perm)
			}
			for i, perm := range sg.IpPermissionsEgress {
				flattenPermission(props, "egress", i, perm)
			}

			r := Resource{
				Type:       "ec2:security-group",
				ID:         aws.ToString(sg.GroupId),
				Region:     cfg.Region,
				AccountID:  accountID,
				Tags:       tagsToMap(sg.Tags),
				Properties: props,
				FetchedAt:  now,
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

// flattenPermission encodes a single IpPermission into flat properties.
// Example keys:
//
//	ingress.0.protocol: tcp
//	ingress.0.from_port: 443
//	ingress.0.to_port: 443
//	ingress.0.cidr: 0.0.0.0/0,10.0.0.0/8
//	ingress.0.ipv6_cidr: ::/0
//	ingress.0.source_sg: sg-abc123
//	ingress.0.prefix_list: pl-abc123
func flattenPermission(props map[string]string, direction string, index int, perm types.IpPermission) {
	prefix := fmt.Sprintf("%s.%d", direction, index)

	protocol := aws.ToString(perm.IpProtocol)
	if protocol == "-1" {
		protocol = "all"
	}
	props[prefix+".protocol"] = protocol

	if perm.FromPort != nil {
		props[prefix+".from_port"] = strconv.Itoa(int(*perm.FromPort))
	}
	if perm.ToPort != nil {
		props[prefix+".to_port"] = strconv.Itoa(int(*perm.ToPort))
	}

	if len(perm.IpRanges) > 0 {
		var cidrs []string
		for _, r := range perm.IpRanges {
			cidrs = append(cidrs, aws.ToString(r.CidrIp))
		}
		props[prefix+".cidr"] = strings.Join(cidrs, ",")
	}

	if len(perm.Ipv6Ranges) > 0 {
		var cidrs []string
		for _, r := range perm.Ipv6Ranges {
			cidrs = append(cidrs, aws.ToString(r.CidrIpv6))
		}
		props[prefix+".ipv6_cidr"] = strings.Join(cidrs, ",")
	}

	if len(perm.UserIdGroupPairs) > 0 {
		var sgs []string
		for _, pair := range perm.UserIdGroupPairs {
			sgs = append(sgs, aws.ToString(pair.GroupId))
		}
		props[prefix+".source_sg"] = strings.Join(sgs, ",")
	}

	if len(perm.PrefixListIds) > 0 {
		var pls []string
		for _, pl := range perm.PrefixListIds {
			pls = append(pls, aws.ToString(pl.PrefixListId))
		}
		props[prefix+".prefix_list"] = strings.Join(pls, ",")
	}
}
