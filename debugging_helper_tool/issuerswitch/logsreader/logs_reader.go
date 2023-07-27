package logsreader

import (
        "context"
        "fmt"
        "strings"

        "cloud.google.com/go/logging/logadmin"
        "github.com/GoogleCloudPlatform/payment-gateway/debugging_helper_tool/issuerswitch/output"
        "google.golang.org/api/iterator"
        "google.golang.org/api/option"
        "google.golang.org/protobuf/types/known/structpb"
)

var (
        resourceType = "issuerswitch.googleapis.com/UPIInstance"
)

// Client is a wrapper around the underlying cloud logging client.
type Client struct {
        logClient *logadmin.Client
}

// txnIDsFilterString returns the filter string in cloud logging query language format required
// to get txnIDs which are associaciated with the given input parameters.
func txnIDsFilterString(txnID, rrn, instanceID string) string {
        var filter string
        if txnID != "" {
                filter = txnID
        } else {
                filter = fmt.Sprintf("\"custRef=\\\"%s\\\"\"", rrn)
        }
        return fmt.Sprintf("resource.type = %s\n"+
                "resource.labels.instance_id = %s\n"+
                "%s\n", resourceType, instanceID, filter)
}

// txnLogsFilterString returns the filter string in cloud logging query language format required
// to get logs for a given txnID and msgID.
func txnLogsFilterString(txnID, msgID, instanceID string) string {
        return fmt.Sprintf("resource.type = %s\n"+
                "resource.labels.instance_id = %s\n"+
                "jsonPayload.transactionId=%s\n"+
                "jsonPayload.messageId=%s\n", resourceType, instanceID, txnID, msgID)
}

// baLogsFilterString returns the filter string in cloud logging query language format required
// for getting bank adapter logs.
func baLogsFilterString(reqID, instanceID string) string {
        return fmt.Sprintf("resource.type = %s\n"+
                "NOT resource.labels.instance_id = %s\n"+
                "%s\n", resourceType, instanceID, reqID)
}

// New returns a new Cloud Logging client given a project id.
func New(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
        logClient, err := logadmin.NewClient(ctx, projectID, opts...)
        if err != nil {
                return nil, fmt.Errorf("error while getting cloud logging client: %v", err)
        }
        return &Client{
                logClient: logClient,
        }, nil
}

// ExtractTxnDetails returns the transaction details of all APIs whose logs can be extracted using
// the given filter.
func (c *Client) ExtractTxnDetails(ctx context.Context, txnID, rrn, instanceID string) ([]*output.TxnDetails, error) {
        filter := logadmin.Filter(txnIDsFilterString(txnID, rrn, instanceID))
        var txnList []*output.TxnDetails
        idMap := make(map[string]bool)
        it := c.logClient.Entries(ctx, filter)
        for {
                nextLog, err := it.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return nil, fmt.Errorf("error occurred while getting next Log: %v", err)
                }

                txnID := nextLog.Payload.(*structpb.Struct).Fields["transactionId"].GetStringValue()
                msgID := nextLog.Payload.(*structpb.Struct).Fields["messageId"].GetStringValue()
                txnType := nextLog.Payload.(*structpb.Struct).Fields["transactionType"].GetStringValue()
                if msgID == "" {
                        continue
                }
                if !idMap[txnID] {
                        idMap[txnID] = true
                        txnList = append(txnList, &output.TxnDetails{
                                TxnID:   txnID,
                                MsgID:   msgID,
                                TxnType: txnType,
                        })
                }
        }
        return txnList, nil
}

// ExtractTxnLogs extracts all logs matching the filter of transactionID and messageID and
// returns as a list of Log.
func (c *Client) ExtractTxnLogs(ctx context.Context, txnID, msgID, instanceID string) ([]*output.Log, error) {
        filter := logadmin.Filter(txnLogsFilterString(txnID, msgID, instanceID))
        var txnLogs []*output.Log
        it := c.logClient.Entries(ctx, filter)
        for {
                nextLog, err := it.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return nil, fmt.Errorf("error occured while getting next Log: %v", err)
                }

                logPayload := &output.Log{
                        DescMsg:     nextLog.Payload.(*structpb.Struct).Fields["message"].GetStringValue(),
                        ErrMsg:      nextLog.Payload.(*structpb.Struct).Fields["errorMessage"].GetStringValue(),
                        Timestamp:   nextLog.Timestamp,
                        PayloadSent: nextLog.Payload.(*structpb.Struct).Fields["sent"].GetStringValue(),
                        PayloadRecv: nextLog.Payload.(*structpb.Struct).Fields["received"].GetStringValue(),
                }
                txnLogs = append(txnLogs, logPayload)
        }
        return txnLogs, nil
}

// ExtractBAlogsAndPrint prints logs from bank adapter.
func (c *Client) ExtractBALogsAndPrint(ctx context.Context, baLog, instanceID string) error {
        reqID := requestID(baLog)
        filter := logadmin.Filter(baLogsFilterString(reqID, instanceID))
        it := c.logClient.Entries(ctx, filter)
        return output.PrintBALogs(it, reqID)
}

// requestID extracts the bank adapter request ID from the log
func requestID(baLog string) string {
        s := strings.Split(baLog, " ")
        for i, word := range s {
                if word == "id:" {
                        return s[i+1]
                }
        }
        return ""
}
