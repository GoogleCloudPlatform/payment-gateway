package output

import (
        "encoding/json"
        "fmt"
        "regexp"
        "time"

        "cloud.google.com/go/logging/logadmin"
        "google.golang.org/api/iterator"
        "google.golang.org/protobuf/types/known/structpb"
)

var (
        istLoc, _ = time.LoadLocation("Asia/Kolkata")
)

// Log represents the fields to be printed for each log extracted from Cloud Logging.
type Log struct {
        DescMsg     string
        ErrMsg      string
        Timestamp   time.Time
        PayloadSent string
        PayloadRecv string
}

type TxnDetails struct {
        TxnID   string
        MsgID   string
        TxnType string
}

// PrintBALogs prints a logs obtained from bank adapter.
func PrintBALogs(it *logadmin.EntryIterator, reqID string) error {
        fmt.Println("\n\n______________________BANK_ADAPTER_LOGS________________________\n")
        fmt.Printf("Request ID  ::  %s\n\n", reqID)
        for {
                log, err := it.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return err
                }
                fmt.Printf("[ %v ]\n", log.Timestamp.In(istLoc))
                payload := log.Payload.(*structpb.Struct)
                jsonString, _ := json.MarshalIndent(payload, "", "  ")
                fmt.Print(string(jsonString))
        }
        return nil
}

// PrintTxnDetails prints the transaction details that were obtained from Cloud Logging.
func PrintTxnDetails(td *TxnDetails) {
        fmt.Print("\n\n_____________________TRANSACTION_DETAILS_______________________\n\n")
        fmt.Println("Transaction ID :: ", td.TxnID)
        fmt.Println("Message ID :: ", td.MsgID)
        fmt.Println("Transaction Type :: ", td.TxnType)
        fmt.Println("\n_______________________________________________________________")
}

// PrintLogsAndExtractBALog classifies prints the logs of a particular transactionID, messageID and
// also returns the log which contains the bank adapter request ID.
func PrintLogsAndExtractReqIDLog(txnLogs []*Log, verbose bool) string {
        var baLog string
        fmt.Printf("\nPrinting Logs for the above Transaction :: \n\n")
        for _, log := range txnLogs {
                printLog(log, verbose)
                re := regexp.MustCompile(".*bank adapter.*request id:.*")
                if re.MatchString(log.DescMsg) {
                        baLog = log.DescMsg
                }
        }
        return baLog
}

// printLog prints a single logObj based on verbose flag.
func printLog(l *Log, verbose bool) {
        fmt.Printf("[ %v ] ::  ", l.Timestamp.In(istLoc))
        fmt.Print(l.DescMsg)
        if l.ErrMsg != "" {
                fmt.Printf(":: ERROR : %v\n", l.ErrMsg)
        } else {
                fmt.Println()
        }

        // incase the verbose flag is set to true, we will print payload received and sent as well
        // along with the logs.
        if verbose {
                if l.PayloadSent != "" {
                        fmt.Println(l.PayloadSent)
                }
                if l.PayloadRecv != "" {
                        fmt.Println(l.PayloadRecv)
                }
                fmt.Println()
        }
}
