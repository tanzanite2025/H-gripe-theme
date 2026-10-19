package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYanwenProductIdentifiersRequireBothCodesAcrossAllEuropeanUnionCountriesBeforeNetwork(t *testing.T) {
	client := NewYanwenGatewayClient()
	for _, country := range strings.Fields("AT BE BG CY HR CZ DK EE FI FR DE GR HU IE IT LV LT LU MT NL PL PT RO SK SI ES SE") {
		t.Run(country, func(t *testing.T) {
			_, err := client.CreateYanwenOrder(context.Background(), yanwenGatewayCredentials{}, YanwenCreateOrderRequest{
				ChannelID: "481", OrderSource: "test", OrderNumber: "ORDER-PID",
				ReceiverInfo: YanwenReceiverInfo{Country: " " + strings.ToLower(country) + " "},
				ParcelInfo: YanwenParcelInfo{ProductList: []YanwenProductDeclare{
					{GoodsNameChinese: "Spoke", SKU: "SKU-IS-NOT-PID", MerchantProductIdentifier: " \t ", ManufacturerProductIdentifier: "\n"},
				}},
			})
			require.ErrorContains(t, err, "MPID")
			require.ErrorContains(t, err, "NSPID")
			require.ErrorContains(t, err, "SKU-IS-NOT-PID")
			require.ErrorContains(t, err, "本次未提交燕文")
		})
	}
}

func TestYanwenProductIdentifiersValidateOptionalLengthAndKeepNonEuropeanDestinationsOptional(t *testing.T) {
	for _, country := range []string{"US", "GB", "NO", "CH", "KR", "AU"} {
		request := YanwenCreateOrderRequest{ReceiverInfo: YanwenReceiverInfo{Country: country}, ParcelInfo: YanwenParcelInfo{ProductList: []YanwenProductDeclare{{}}}}
		require.NoError(t, normalizeAndValidateYanwenProductDeclarationIdentifiers(&request))
	}
	for _, field := range []string{"mpid", "nspid", "spid"} {
		t.Run(field, func(t *testing.T) {
			item := YanwenProductDeclare{MerchantProductIdentifier: "SHOP", ManufacturerProductIdentifier: "FACTORY"}
			long := strings.Repeat("字", 71)
			switch field {
			case "mpid":
				item.MerchantProductIdentifier = long
			case "nspid":
				item.ManufacturerProductIdentifier = long
			case "spid":
				item.StandardProductIdentifier = long
			}
			request := YanwenCreateOrderRequest{ReceiverInfo: YanwenReceiverInfo{Country: "DE"}, ParcelInfo: YanwenParcelInfo{ProductList: []YanwenProductDeclare{item}}}
			err := normalizeAndValidateYanwenProductDeclarationIdentifiers(&request)
			require.ErrorContains(t, err, strings.ToUpper(field))
			require.ErrorContains(t, err, "70")
		})
	}
}

func TestYanwenProductIdentifiersSerializeExactOfficialKeysAndOmitAbsentStandardCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		products := payload["parcelInfo"].(map[string]any)["productList"].([]any)
		first := products[0].(map[string]any)
		require.Equal(t, strings.Repeat("字", 70), first["mpid"])
		require.Equal(t, "FACTORY-001", first["nspid"])
		require.Equal(t, "00012345678905", first["spid"])
		require.NotContains(t, first, "merchant_product_identifier")
		require.NotContains(t, products[1].(map[string]any), "spid")
		raw := "user" + string(body) + "json" + yanwenCreateOrderMethod + r.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest("token", raw), r.URL.Query().Get("sign"))
		fmt.Fprint(w, `{"success":true,"code":"0","data":{"waybillNumber":"YE-PID","orderNumber":"ORDER-PID"}}`)
	}))
	defer server.Close()
	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	request := YanwenCreateOrderRequest{
		ChannelID: "481", OrderSource: "test", OrderNumber: "ORDER-PID",
		ReceiverInfo: YanwenReceiverInfo{Country: "DE"},
		ParcelInfo: YanwenParcelInfo{ProductList: []YanwenProductDeclare{
			{MerchantProductIdentifier: " " + strings.Repeat("字", 70) + " ", ManufacturerProductIdentifier: " FACTORY-001 ", StandardProductIdentifier: " 00012345678905 "},
			{MerchantProductIdentifier: "SHOP-002", ManufacturerProductIdentifier: "FACTORY-002", StandardProductIdentifier: " "},
		}},
	}
	_, err := client.CreateYanwenOrder(context.Background(), yanwenGatewayCredentials{environment: "fat", endpoint: server.URL, userID: "user", apiToken: "token"}, request)
	require.NoError(t, err)
	require.Equal(t, " FACTORY-001 ", request.ParcelInfo.ProductList[0].ManufacturerProductIdentifier, "gateway normalization must not mutate the caller's snapshot")
}
