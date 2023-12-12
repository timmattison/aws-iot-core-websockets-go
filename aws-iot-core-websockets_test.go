package aws_iot_core_websockets_go

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"testing"
)

func TestCredentialsAreValid(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sts.NewFromConfig(cfg)

	input := &sts.GetCallerIdentityInput{}

	_, err = client.GetCallerIdentity(ctx, input)

	if err != nil {
		t.Errorf("Could not authenticate to AWS APIs. This is likely due to credentials missing or having expired. [%s]", err.Error())
	}
}

func TestGetEndpoint(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	endpoint, err := getEndpoint(ctx, cfg)

	if err != nil {
		t.Errorf("Could not get endpoint. [%s]", err.Error())
	}

	if endpoint == "" {
		t.Errorf("Endpoint was empty.")
	}
}

func TestGetWebSocketUrl(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	endpoint, err := getEndpoint(ctx, cfg)

	if err != nil {
		t.Errorf("Could not get endpoint. [%s]", err.Error())
	}

	iotWsConfig := IotWsConfig{
		AwsConfig: cfg,
		Endpoint:  endpoint,
	}

	wsUrl, _ := AwsIotWsUrl(ctx, iotWsConfig)

	if wsUrl == "" {
		t.Errorf("Could not get WebSocket URL")
	}
}

func TestWebSocketConnect(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	endpoint, err := getEndpoint(ctx, cfg)

	if err != nil {
		t.Errorf("Could not get endpoint. [%s]", err.Error())
	}

	iotWsConfig := IotWsConfig{
		AwsConfig: cfg,
		Endpoint:  endpoint,
	}

	wsUrl, _ := AwsIotWsUrl(ctx, iotWsConfig)

	certificatePool, err := createCertificatePool()

	if err != nil {
		t.Errorf("Could not create certificate pool. [%s]", err.Error())
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(wsUrl)
	opts.SetTLSConfig(&tls.Config{RootCAs: certificatePool})
	opts.SetClientID("test")

	client := mqtt.NewClient(opts)

	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		t.Errorf("Could not connect to WebSocket URL. [%s]", token.Error())
	}
}

// This is the root CA certificate from https://www.amazontrust.com/repository/AmazonRootCA1.pem
const RootCa = `-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
ADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6
b24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv
b3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj
ca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM
9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw
IFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6
VOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L
93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm
jgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC
AYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA
A4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI
U5PMCCjjmCXPI6T53iHTfIUJrU6adTrCC2qJeHZERxhlbI1Bjjt/msv0tadQ1wUs
N+gDS63pYaACbvXy8MWy7Vu33PqUXHeeE6V/Uq2V8viTO96LXFvKWlJbYK8U90vv
o/ufQJVtMVT8QtPHRh8jrdkPSHCa2XV4cdFyQzR1bldZwgJcJmApzyMZFo6IQ6XU
5MsI+yMRQ+hDKXJioaldXgjUkK642M4UwtBV8ob2xJNDd2ZhwLnoQdeXeGADbkpy
rqXRfboQnoZsG4q5WTP468SQvvG5
-----END CERTIFICATE-----`

func createCertificatePool() (*x509.CertPool, error) {
	certificatePool := x509.NewCertPool()

	if ok := certificatePool.AppendCertsFromPEM([]byte(RootCa)); !ok {
		return nil, fmt.Errorf("failed to add root CA certificate to certificate pool")
	}

	return certificatePool, nil
}

func getEndpoint(ctx context.Context, cfg aws.Config) (string, error) {
	client := iot.NewFromConfig(cfg)

	endpoint, err := client.DescribeEndpoint(ctx, &iot.DescribeEndpointInput{
		EndpointType: aws.String("iot:Data-ATS"),
	})

	if err != nil {
		return "", err
	}

	return *endpoint.EndpointAddress, nil
}
