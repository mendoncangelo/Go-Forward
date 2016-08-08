package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const listenAddress = "localhost:5514"

// AWS CloudWatch specific constants.
const (
	// Maximum number of log events in a batch.
	maxBatchEvents = 10000
	// Maximum batch size in bytes.
	maxBatchSize = 1048576
	// A batch of log events in a single PutLogEvents request cannot span more than 24 hours.
	maxBatchSpanTime = 24 * time.Hour
	// How many bytes to append to each log event.
	eventSizeOverhead = 26
	// None of the log events in the batch can be more than 2 hours in the future.
	eventFutureTimedelta = 2 * time.Hour
	// None of the log events in the batch can be older than 14 days.
	eventPastTimedelta = 14 * 24 * time.Hour
	// DescribeLogStreams transactions/second.
	describelogstreamsTps = 5
)

var params = &cloudwatchlogs.DescribeLogGroupsInput{Limit: aws.Int64(50)}

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	var buf [2048]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			panic(err)
		}
		parsed, err := decodeMessage(string(buf[0:n]))
		if err == nil {
			fmt.Println(parsed)
		} else {
			fmt.Println(err)
		}
	}
}

func getLogGroups(region string) (groups []string) {
	svc := cloudwatchlogs.New(session.New(), aws.NewConfig().WithRegion(region))

	err := svc.DescribeLogGroupsPages(params,
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			groups = append(groups, getGroupNames(page)...)
			return !lastPage
		})

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		// awsErr  := err.(awserr.Error)
		// Generic AWS Error with Code, Message, and original error (if any)
		// fmt.Println(awsErr.Code(), awsErr.Message())
		panic(err)
	}
	return
}

func getGroupNames(page *cloudwatchlogs.DescribeLogGroupsOutput) (names []string) {
	for _, val := range page.LogGroups {
		names = append(names, *val.LogGroupName)
	}
	return
}