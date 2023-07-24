package main

import (
        "context"
        "encoding/json"
        "fmt"
        "log"

        "github.com/golang/protobuf/jsonpb"
        "github.com/golang/protobuf/proto"

        pb "google.golang.org/genproto/googleapis/cloud/paymentgateway/issuerswitch/v1"
)

func protoToJSON(message proto.Message) (string, error) {
        marshaler := &jsonpb.Marshaler{}
        jsonString, err := marshaler.MarshalToString(message)
        if err != nil {
                return "", err
        }
        return jsonString, nil
}

func indentedJSON(jsonString string) (string, error) {
        var rawJSON map[string]interface{}
        err := json.Unmarshal([]byte(jsonString), &rawJSON)
        if err != nil {
                return "", err
        }
        indentedJSONBytes, err := json.MarshalIndent(rawJSON, "", "  ")
        if err != nil {
                return "", err
        }
        return string(indentedJSONBytes), nil
}

func printJSONResponse(message proto.Message) {
        // Convert proto message to JSON
        jsonString, err := protoToJSON(message)
        if err != nil {
                log.Fatal("Error converting proto message to JSON:", err)
        }

        indentedJSONString, err := indentedJSON(jsonString)
        if err != nil {
                log.Fatal("Error printing JSON:", err)
        }
        fmt.Println(indentedJSONString)
}

func PrintListTransactionsResponse(ctx context.Context, istc pb.IssuerSwitchTransactionsClient, reqFilter []string) error {
        // Invoke ListFinancialTransactions
        for _, filter := range reqFilter {
                reqFinancial := pb.ListFinancialTransactionsRequest{
                        Parent:   "projects/common-dev-6",
                        Filter:   filter,
                        PageSize: 10,
                }
                respFinancial, err := istc.ListFinancialTransactions(ctx, &reqFinancial)
                if err != nil {
                        return fmt.Errorf("ListFinancialTransactions returned unexpected error: %v", err)
                }
                allRespFinancial := respFinancial.GetFinancialTransactions()
                if allRespFinancial != nil {
                        for _, r := range allRespFinancial {
                                fmt.Println("\n\n_______________________Financial_Response_______________________\n")
                                printJSONResponse(r)
                        }
                }

                // Invoke ListMetadataTransactions
                reqMedatada := pb.ListMetadataTransactionsRequest{
                        Parent:   "projects/common-dev-6",
                        Filter:   filter,
                        PageSize: 10,
                }
                respMetadata, err := istc.ListMetadataTransactions(ctx, &reqMedatada)
                if err != nil {
                        return fmt.Errorf("ListMetadataTransactions returned unexpected error: %v", err)
                }
                allRespMetadata := respMetadata.GetMetadataTransactions()
                if allRespMetadata != nil {
                        for _, r := range allRespMetadata {
                                fmt.Println("\n\n_______________________Metadata_Response_______________________\n")
                                printJSONResponse(r)
                        }
                }

                // Invoke ListMandateTransactions
                reqMandate := pb.ListMandateTransactionsRequest{
                        Parent:   "projects/common-dev-6",
                        Filter:   filter,
                        PageSize: 10,
                }
                respMandate, err := istc.ListMandateTransactions(ctx, &reqMandate)
                if err != nil {
                        return fmt.Errorf("ListMandateTransactions returned unexpected error: %v", err)
                }
                allRespMandate := respMandate.GetMandateTransactions()
                if allRespMandate != nil {
                        for _, r := range allRespMandate {
                                fmt.Println("\n\n_______________________Mandate_Response_______________________\n")
                                printJSONResponse(r)
                        }
                }
        }
        return nil
}
