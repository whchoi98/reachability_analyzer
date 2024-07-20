package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatalf("Usage: %s <region> <source> <destination> <protocol> <port>", os.Args[0])
	}

	region, source, destination, protocol, portStr := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	sourceIP, err := resolveToIP(source)
	if err != nil || !isPrivateIP(sourceIP) {
		log.Fatalf("Failed to resolve or invalid source IP: %v", err)
	}

	destinationIP, err := resolveToIP(destination)
	if err != nil {
		log.Fatalf("Failed to resolve destination: %v", err)
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	ec2Svc := ec2.New(sess)
	sourceNetworkInterfaceID, err := findNetworkInterfaceByIP(ec2Svc, sourceIP)
	if err != nil {
		log.Fatalf("Failed to find network interface for source IP: %v", err)
	}

	var destinationNetworkInterfaceID string
	if isPrivateIP(destinationIP) {
		destinationNetworkInterfaceID, err = findNetworkInterfaceByIP(ec2Svc, destinationIP)
		if err != nil {
			log.Fatalf("Failed to find network interface for destination IP: %v", err)
		}
	} else {
		destinationNetworkInterfaceID, err = findInternetGateway(ec2Svc)
		if err != nil {
			log.Fatalf("Failed to find internet gateway for destination IP: %v", err)
		}
	}

	result, err := ec2Svc.CreateNetworkInsightsPath(&ec2.CreateNetworkInsightsPathInput{
		Source:          aws.String(sourceNetworkInterfaceID),
		Destination:     aws.String(destinationNetworkInterfaceID),
		Protocol:        aws.String(protocol),
		DestinationPort: aws.Int64(int64(port)),
	})
	if err != nil {
		log.Fatalf("Failed to create network insights path: %v", err)
	}

	networkInsightsPathID := *result.NetworkInsightsPath.NetworkInsightsPathId
	fmt.Printf("Network Insights Path ID: %s\n", networkInsightsPathID)

	analysisResult, err := ec2Svc.StartNetworkInsightsAnalysis(&ec2.StartNetworkInsightsAnalysisInput{
		NetworkInsightsPathId: result.NetworkInsightsPath.NetworkInsightsPathId,
	})
	if err != nil {
		log.Fatalf("Failed to start network insights analysis: %v", err)
	}

	analysisID := analysisResult.NetworkInsightsAnalysis.NetworkInsightsAnalysisId
	describeInput := &ec2.DescribeNetworkInsightsAnalysesInput{
		NetworkInsightsAnalysisIds: []*string{analysisID},
	}

	for {
		describeResult, err := ec2Svc.DescribeNetworkInsightsAnalyses(describeInput)
		if err != nil {
			log.Fatalf("Failed to describe network insights analysis: %v", err)
		}

		analysis := describeResult.NetworkInsightsAnalyses[0]
		if *analysis.Status == "succeeded" || *analysis.Status == "failed" {
			visualizeAnalysis(analysis, sourceIP, destinationIP, networkInsightsPathID)
			break
		}

		fmt.Println("Analysis in progress...")
		time.Sleep(10 * time.Second)
	}
}

func resolveToIP(address string) (string, error) {
	if net.ParseIP(address) != nil {
		return address, nil
	}
	ips, err := net.LookupIP(address)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("failed to resolve address %s: %v", address, err)
	}
	return ips[0].String(), nil
}

func findNetworkInterfaceByIP(svc *ec2.EC2, ip string) (string, error) {
	result, err := svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("addresses.private-ip-address"), Values: []*string{aws.String(ip)}},
		},
	})
	if err != nil || len(result.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("no network interface found for IP %s", ip)
	}
	return *result.NetworkInterfaces[0].NetworkInterfaceId, nil
}

func findInternetGateway(svc *ec2.EC2) (string, error) {
	result, err := svc.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err != nil || len(result.InternetGateways) == 0 {
		return "", fmt.Errorf("no internet gateway found")
	}
	return *result.InternetGateways[0].InternetGatewayId, nil
}

func isPrivateIP(ip string) bool {
	privateIPBlocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
	}
	parsedIP := net.ParseIP(ip)
	for _, block := range privateIPBlocks {
		if block.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func visualizeAnalysis(analysis *ec2.NetworkInsightsAnalysis, sourceIP, destinationIP, networkInsightsPathID string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Println("Path Visualization:")
	fmt.Fprintln(w, "Hop\tComponent ID\tACL Rule\t")
	fmt.Fprintf(w, "Source\t%s\t\t\n", sourceIP)

	for i, component := range analysis.ForwardPathComponents {
		componentID := *component.Component.Id
		aclRule := "N/A"
		if component.AclRule != nil {
			aclRule = *component.AclRule.RuleAction
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t\n", i+1, componentID, aclRule)
	}

	fmt.Fprintf(w, "Destination\t%s\t\t\n", destinationIP)
	w.Flush()

	fmt.Println("\nReachability status:")
	if *analysis.NetworkPathFound {
		fmt.Println("Reachable")
		fmt.Println("모든 정책이 정상적으로 허용되어 있습니다.")
	} else {
		fmt.Println("Not reachable")
		fmt.Printf("Network Insights Path ID: %s 값을 확인해서, Cloud 운영팀으로 연락 주세요.\n", networkInsightsPathID)
	}
}
