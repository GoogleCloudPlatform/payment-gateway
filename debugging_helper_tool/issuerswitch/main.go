package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/logging"
	"github.com/GoogleCloudPlatform/payment-gateway/debugging_helper_tool/issuerswitch/listtransactions"
	"github.com/GoogleCloudPlatform/payment-gateway/debugging_helper_tool/issuerswitch/logsreader"
	"github.com/GoogleCloudPlatform/payment-gateway/debugging_helper_tool/issuerswitch/output"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
)

var (
	txnID      = flag.String("txnid", "", "Transaction ID of the request for which logs are to be displayed")
	rrn        = flag.String("rrn", "", "RRN of the payment for which logs are to be displayed")
	addr       = flag.String("addr", "", "API Service external endpoint")
	verbose    = flag.Bool("verbose", false, "print verbose output")
	projectID  = flag.String("project_id", "", "project ID required for Cloud Logging API and for List Transactions API")
	instanceID = flag.String("instance_id", "", "issuerswitch instance ID for Cloud Logging API")
)

func validateParams() error {
	if *rrn == "" && *txnID == "" {
		return fmt.Errorf("User should provide atleast one of rrn or txnid")
	}
	if *addr == "" {
		return fmt.Errorf("User should provide api service endpoint")
	}
	if *projectID == "" {
		return fmt.Errorf("User should provide project ID for Cloud Logging and List Transactions API")
	}
	if *instanceID == "" {
		return fmt.Errorf("User should provide instance ID for Cloud Logging API")
	}
	return nil
}

func fetchLogsAndPrintTxnDetails(ctx context.Context, client *logsreader.Client, detailsList []*output.TxnDetails) {
	for _, d := range detailsList {
		logs, err := client.ExtractTxnLogs(ctx, d.TxnID, d.MsgID, *instanceID)
		if err != nil {
			log.Printf("Error occurred while getting transaction logs from Cloud Logging: %v", err)
		}
		output.PrintTxnDetails(d)
		reqIDLog := output.PrintLogsAndExtractReqIDLog(logs, *verbose)
		if reqIDLog != "" {
			err = client.ExtractBALogsAndPrint(ctx, reqIDLog, *instanceID)
			if err != nil {
				log.Printf("Error occurred while printing Bank Adapter logs")
			}
		}
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	err := validateParams()
	if err != nil {
		log.Fatalf("Missing Argument: %v", err)
	}

	opts := []option.ClientOption{
		option.WithScopes(logging.ReadScope),
		internaloption.WithDefaultEndpoint("staging-logging.sandbox.googleapis.com:443"),
	}
	logClient, err := logsreader.New(ctx, *projectID, opts...)
	if err != nil {
		log.Printf("Error occurred while getting cloud logging API client: %v", err)
	}

	// getting logs from cloud Logging.
	txnDetails, err := logClient.ExtractTxnDetails(ctx, *txnID, *rrn, *instanceID)
	if err != nil {
		log.Printf("Error occurred while getting transaction details from Cloud Logging: %v", err)
	}
	fetchLogsAndPrintTxnDetails(ctx, logClient, txnDetails)

	// Flow for calling ListTransactions APIs.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	listTxnClient := listtransactions.NewClient(conn)
	if err := listTxnClient.ExtractResponseAndPrint(ctx, *projectID, *rrn, txnDetails); err != nil {
		fmt.Printf("Error occurred while invoking ListTransactions API: %v", err)
	}
}

