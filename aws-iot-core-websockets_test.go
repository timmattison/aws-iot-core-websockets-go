package aws_iot_core_websockets_go

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
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
