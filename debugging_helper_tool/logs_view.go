package main

import (
  "context"
  "fmt"
  "regexp"
  "strings"
  "time"

  "cloud.google.com/go/logging/logadmin"
  "google.golang.org/api/iterator"
  "google.golang.org/api/option"
  "google.golang.org/protobuf/types/known/structpb"
)

const (
  resourceType     = "issuerswitch.googleapis.com/UPIInstance"
  projectIDLogging = "common-dev-6"
)

// logReader is a wrapper around the underlying cloud logging client.
type logReader struct {
  client *logadmin.Client
}

// getFilterStringingIDs returns the filter string in cloud logging query language format.
func getFilterStringIDs(filter string) string {
  return fmt.Sprintf("resource.type = %s\n%s\n", resourceType, filter)
}

// getFilterStringing returns the filter string in cloud logging query language format.
func getFilterString(txnID, msgID string) string {
  return fmt.Sprintf("resource.type = %s\n"+
   "jsonPayload.transactionId=%s\n"+
   "jsonPayload.messageId=%s\n", resourceType, txnID, msgID)
}

// logObj represents the fields to be printed for each log extracted from Cloud Logging.
type logObj struct {
  logMsg    string
  errMsg    string
  timestamp time.Time
}

type txnLogDetails struct {
  txnid   string
  msgid   string
  txnType string
}

func (tld *txnLogDetails) getMsgId() string {
  return tld.msgid
}

func (tld *txnLogDetails) getTxnId() string {
  return tld.txnid
}

// newLogClient returns a new Cloud Logging client given a project id.
func newLogClient(ctx context.Context, opts ...option.ClientOption) (*logReader, error) {
  logClient, err := logadmin.NewClient(ctx, projectIDLogging, opts...)
  if err != nil {
   return nil, fmt.Errorf("error while getting cloud logging client: %v", err)
  }
  return &logReader{
   client: logClient,
  }, nil
}

func (lr *logReader) getTxnIDs(ctx context.Context, filter logadmin.EntriesOption) ([]*txnLogDetails, error) {
  var txnList []*txnLogDetails
  idMap := make(map[string]bool)
  it := lr.client.Entries(ctx, filter)
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
    txnList = append(txnList, &txnLogDetails{
     txnid:   txnID,
     msgid:   msgID,
     txnType: txnType,
    })
   }
  }
  return txnList, nil
}

// getLogs returns all logs matching the filter, classifies them into different services and prints the result.
func (lr *logReader) getLogs(ctx context.Context, filter logadmin.EntriesOption) ([]*logObj, error) {
  it := lr.client.Entries(ctx, filter)
  logs, err := processLogs(it)
  if err != nil {
   return nil, err
  }
  return logs, nil
}

// processLogs process logs corresponding to a transaction id and analyse it.
func processLogs(it *logadmin.EntryIterator) ([]*logObj, error) {
  var txnLogs []*logObj
  for {
   nextLog, err := it.Next()
   if err == iterator.Done {
    break
   }
   if err != nil {
    return nil, fmt.Errorf("error occured while getting next Log: %v", err)
   }

   logPayload := &logObj{
    logMsg:    nextLog.Payload.(*structpb.Struct).Fields["message"].GetStringValue(),
    errMsg:    nextLog.Payload.(*structpb.Struct).Fields["errorMessage"].GetStringValue(),
    timestamp: nextLog.Timestamp,
   }
   txnLogs = append(txnLogs, logPayload)
  }
  return txnLogs, nil
}

// processAndPrintLogs classifies each logObj into the service that they arrive from and print it.
func processAndPrintLogs(txnLogs []*logObj) (string, error) {
  var baLog string
  istLoc, _ := time.LoadLocation("Asia/Kolkata")
  fmt.Printf("\nPrinting Logs for the above Transaction :: \n")
  for _, logObj := range txnLogs {
   printLog(logObj, istLoc)
   re := regexp.MustCompile(".*request id:.*")
   if re.MatchString(logObj.logMsg) {
    baLog = logObj.logMsg
   }
  }
  fmt.Printf("\n\n")
  return baLog, nil
}

func printTxnDetails(tld *txnLogDetails) {
  fmt.Print("\n\n_____________________TRANSACTION_DETAILS_______________________\n\n")
  fmt.Println("Transaction ID :: ", tld.txnid)
  fmt.Println("Message ID :: ", tld.msgid)
  fmt.Println("Transaction Type :: ", tld.txnType)
  fmt.Println("\n_____________________________________________________________")
}

func (lr *logReader) requestID(logObj string) string {
  s := strings.Split(logObj, " ")
  for i, word := range s {
   if word == "id:" {
    return s[i+1]
   }
  }
  return ""
}

func printLog(l *logObj, istLoc *time.Location) {
  fmt.Printf("[ %v ] ::  ", l.timestamp.In(istLoc))
  fmt.Print(l.logMsg)
  if l.errMsg != "" {
   fmt.Printf(":: ERROR : %v\n", l.errMsg)
  } else {
   fmt.Println()
  }
}
