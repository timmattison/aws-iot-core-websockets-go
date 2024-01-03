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
		return
	}
}

func TestWithManualConfigAndManualEndpointAndManualCertPool(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	endpoint, err := getEndpoint(ctx, cfg)

	if err != nil {
		t.Errorf("Could not get endpoint. [%s]", err.Error())
		return
	}

	certPool, err := createCertificatePool()

	if err != nil {
		t.Errorf("Could not create certificate pool. [%s]", err.Error())
	}

	mqttOptions, err := NewMqttOptions(ctx,
		NewOptions().
			WithAwsConfig(cfg).
			WithEndpoint(endpoint).
			WithCertificatePool(certPool))

	if err != nil {
		t.Errorf("Could not get MQTT options. [%s]", err.Error())
		return
	}

	connect(t, mqttOptions)
}

func TestWithManualConfigAndManualEndpoint(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	endpoint, err := getEndpoint(ctx, cfg)

	if err != nil {
		t.Errorf("Could not get endpoint. [%s]", err.Error())
		return
	}

	mqttOptions, err := NewMqttOptions(ctx,
		NewOptions().
			WithAwsConfig(cfg).
			WithEndpoint(endpoint))

	if err != nil {
		t.Errorf("Could not get MQTT options. [%s]", err.Error())
		return
	}

	connect(t, mqttOptions)
}

func TestWithManualConfigAndAutoEndpoint(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	mqttOptions, err := NewMqttOptions(ctx,
		NewOptions().
			WithAwsConfig(cfg))

	if err != nil {
		t.Errorf("Could not get MQTT options. [%s]", err.Error())
		return
	}

	connect(t, mqttOptions)
}

func TestWithAutoConfigAndAutoEndpoint(t *testing.T) {
	ctx := context.Background()

	mqttOptions, err := NewMqttOptions(ctx, NewOptions())

	if err != nil {
		t.Errorf("Could not get MQTT options. [%s]", err.Error())
		return
	}

	connect(t, mqttOptions)
}

func TestWithClientCertificateFile(t *testing.T) {
	// NOTE: This test connects with "normal" MQTT, not MQTT over WebSockets.
	ctx := context.Background()

	mqttOptions, err := NewMqttOptions(ctx,
		NewOptions().
			WithClientCertificateFile("certificate.pem", "private.key"))

	if err != nil {
		t.Errorf("Could not get MQTT options. [%s]", err.Error())
		return
	}

	connect(t, mqttOptions)
}

func connect(t *testing.T, mqttOptions *mqtt.ClientOptions) {
	mqttOptions.SetClientID("test")

	client := mqtt.NewClient(mqttOptions)

	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		t.Errorf("Could not connect to broker. [%s]", token.Error())
		return
	}
}
