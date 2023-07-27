package listtransactions

import (
        "context"
        "encoding/json"
        "fmt"

        "github.com/GoogleCloudPlatform/payment-gateway/debugging_helper_tool/issuerswitch/output"
        "google.golang.org/grpc"

        pb "google.golang.org/genproto/googleapis/cloud/paymentgateway/issuerswitch/v1"
)

func listTransactionsFilter(rrn string, td []*output.TxnDetails) []string {
        var filterList []string
        for _, d := range td {
                id := d.TxnID
                filterList = append(filterList, fmt.Sprintf("transactionID = %s", id))
        }
        if filterList == nil {
                filterList = append(filterList, fmt.Sprintf("rrn = %s", rrn))
        }
        return filterList
}

type Client struct {
        listTransactionsClient pb.IssuerSwitchTransactionsClient
}

func NewClient(conn *grpc.ClientConn) *Client {
        return &Client{
                listTransactionsClient: pb.NewIssuerSwitchTransactionsClient(conn),
        }
}

func (c *Client) ExtractResponseAndPrint(ctx context.Context, projectID, rrn string, txnDetails []*output.TxnDetails) error {
        parent := fmt.Sprintf("projects/%s", projectID)
        reqFilterList := listTransactionsFilter(rrn, txnDetails)
        for _, filter := range reqFilterList {
                // Invoke ListFinancialTransactions
                reqFinancial := pb.ListFinancialTransactionsRequest{
                        Parent:   parent,
                        Filter:   filter,
                        PageSize: 10,
                }
                respFinancial, err := c.listTransactionsClient.ListFinancialTransactions(ctx, &reqFinancial)
                if err != nil {
                        return fmt.Errorf("ListFinancialTransactions returned unexpected error: %v", err)
                }
                financialRespList := respFinancial.GetFinancialTransactions()
                if len(financialRespList) > 0 {
                        fmt.Println("\n\n_______________________FINANCIAL_RESPONSE_______________________\n")
                        for _, resp := range financialRespList {
                                indentedResp, _ := json.MarshalIndent(resp, "", "  ")
                                fmt.Println(string(indentedResp))
                        }
                }

                // Invoke ListMetadataTransactions
                reqMedatada := pb.ListMetadataTransactionsRequest{
                        Parent:   parent,
                        Filter:   filter,
                        PageSize: 10,
                }
                respMetadata, err := c.listTransactionsClient.ListMetadataTransactions(ctx, &reqMedatada)
                if err != nil {
                        return fmt.Errorf("ListMetadataTransactions returned unexpected error: %v", err)
                }
                metadataRespList := respMetadata.GetMetadataTransactions()
                if len(metadataRespList) > 0 {
                        fmt.Println("\n\n_______________________METADATA_RESPONSE_______________________\n")
                        for _, resp := range metadataRespList {
                                indentedResp, _ := json.MarshalIndent(resp, "", "  ")
                                fmt.Println(string(indentedResp))
                        }
                }

                // Invoke ListMandateTransactions
                reqMandate := pb.ListMandateTransactionsRequest{
                        Parent:   parent,
                        Filter:   filter,
                        PageSize: 10,
                }
                respMandate, err := c.listTransactionsClient.ListMandateTransactions(ctx, &reqMandate)
                if err != nil {
                        return fmt.Errorf("ListMandateTransactions returned unexpected error: %v", err)
                }
                mandateRespList := respMandate.GetMandateTransactions()
                if len(mandateRespList) > 0 {
                        fmt.Println("\n\n_______________________MANDATE_RESPONSE_______________________\n")
                        for _, resp := range mandateRespList {
                                indentedResp, _ := json.MarshalIndent(resp, "", "  ")
                                fmt.Println(string(indentedResp))
                        }
                }
        }
        return nil
}
