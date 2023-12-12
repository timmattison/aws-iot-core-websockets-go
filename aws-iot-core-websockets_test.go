package aws_iot_core_websockets_go

import (
	"context"
	"crypto/x509"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strings"
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

	wsUrl, err := AwsIotWsUrl(ctx, iotWsConfig)

	if err != nil {
		t.Errorf("Could not get WebSocket URL. [%s]", err.Error())
	}

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

	opts, err := AwsIotWsMqttOptions(ctx, iotWsConfig)

	if err != nil {
		t.Errorf("Could not get MQTT config. [%s]", err.Error())
	}

	opts.SetClientID("test")

	client := mqtt.NewClient(opts)

	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		t.Errorf("Could not connect to WebSocket URL. [%s]", token.Error())
	}
}

func TestWebSocketConnectWithEndpointDiscovery(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	iotWsConfig := IotWsConfig{
		AwsConfig: cfg,
	}

	opts, err := AwsIotWsMqttOptions(ctx, iotWsConfig)

	if err != nil {
		t.Errorf("Could not get MQTT config. [%s]", err.Error())
	}

	opts.SetClientID("test")

	client := mqtt.NewClient(opts)

	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		t.Errorf("Could not connect to WebSocket URL. [%s]", token.Error())
	}
}

func TestWebSocketConnectWithCustomCertificatePool(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	iotWsConfig := IotWsConfig{
		AwsConfig: cfg,
	}

	certificatePool := x509.NewCertPool()
	opts, err := AwsIotWsMqttOptionsCustom(ctx, iotWsConfig, certificatePool)

	if err != nil {
		t.Errorf("Could not get MQTT config. [%s]", err.Error())
	}

	opts.SetClientID("test")

	client := mqtt.NewClient(opts)

	token := client.Connect()

	token.Wait()

	if token.Error() == nil {
		t.Errorf("Connected to WebSocket URL without a valid root CA. This is a bug.")
	}

	if !strings.Contains(token.Error().Error(), "x509: certificate signed by unknown authority") {
		t.Errorf("Didn't get the expected error for this test. Expected TLS error, got [%s]", token.Error())
	}
}
