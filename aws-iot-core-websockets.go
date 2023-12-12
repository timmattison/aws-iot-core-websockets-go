package aws_iot_core_websockets_go

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"net/url"
	"strings"
	"time"
)

type IotWsConfig struct {
	AwsConfig aws.Config
	Endpoint  string
}

const (
	emptyStringHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// AwsIotWsMqttOptionsCustom AwsIotWsMqttOptions creates a new MQTT client options object for connecting to AWS IoT Core via WebSockets.
// The options object will have the broker URL and TLS config set but not the client ID. The TLS config will
// utilize the certificate pool passed into it but will not use the AWS root CA certificate automatically.
func AwsIotWsMqttOptionsCustom(ctx context.Context, iotWsConfig IotWsConfig, certificatePool *x509.CertPool) (*mqtt.ClientOptions, error) {
	if iotWsConfig.Endpoint == "" {
		endpoint, err := getEndpoint(ctx, iotWsConfig.AwsConfig)

		if err != nil {
			return nil, fmt.Errorf("could not get endpoint. [%s]", err.Error())
		}

		iotWsConfig.Endpoint = endpoint
	}

	wsUrl, err := AwsIotWsUrl(ctx, iotWsConfig)

	if err != nil {
		return nil, fmt.Errorf("could not get WebSocket URL. [%s]", err.Error())
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(wsUrl)
	opts.SetTLSConfig(&tls.Config{RootCAs: certificatePool})

	return opts, nil
}

// AwsIotWsMqttOptions creates a new MQTT client options object for connecting to AWS IoT Core via WebSockets.
// The options object will have the broker URL and TLS config set but not the client ID. The TLS config will
// utilize the AWS root CA certificate automatically.
func AwsIotWsMqttOptions(ctx context.Context, iotWsConfig IotWsConfig) (*mqtt.ClientOptions, error) {
	certificatePool, err := createCertificatePool()

	if err != nil {
		return nil, fmt.Errorf("could not create certificate pool. [%s]", err.Error())
	}

	return AwsIotWsMqttOptionsCustom(ctx, iotWsConfig, certificatePool)
}

func AwsIotWsUrl(ctx context.Context, iotWsConfig IotWsConfig) (string, error) {
	credentials, err := iotWsConfig.AwsConfig.Credentials.Retrieve(ctx)

	if err != nil {
		return "", err
	}

	// according to docs, time must be within 5min of actual time (or at least according to AWS servers)
	now := time.Now().UTC()

	dateLong := now.Format("20060102T150405Z")
	dateShort := dateLong[:8]
	serviceName := "iotdevicegateway"
	scope := fmt.Sprintf("%s/%s/%s/aws4_request", dateShort, iotWsConfig.AwsConfig.Region, serviceName)
	alg := "AWS4-HMAC-SHA256"
	q := [][2]string{
		{"X-Amz-Algorithm", alg},
		{"X-Amz-Credential", credentials.AccessKeyID + "/" + scope},
		{"X-Amz-Date", dateLong},
		{"X-Amz-SignedHeaders", "host"},
	}

	query := awsQueryParams(q)

	signKey := awsSignKey(credentials.SecretAccessKey, dateShort, iotWsConfig.AwsConfig.Region, serviceName)
	stringToSign := awsSignString(query, iotWsConfig.Endpoint, dateLong, alg, scope)
	signature := fmt.Sprintf("%x", awsHmac(signKey, []byte(stringToSign)))

	wsUrl := fmt.Sprintf("wss://%s/mqtt?%s&X-Amz-Signature=%s", iotWsConfig.Endpoint, query, signature)

	if credentials.SessionToken != "" {
		wsUrl = fmt.Sprintf("%s&X-Amz-Security-Token=%s", wsUrl, url.QueryEscape(credentials.SessionToken))
	}

	return wsUrl, nil
}

func awsQueryParams(q [][2]string) string {
	var buff bytes.Buffer
	var i int
	for _, param := range q {
		if i != 0 {
			buff.WriteRune('&')
		}
		i++
		buff.WriteString(param[0])
		buff.WriteRune('=')
		buff.WriteString(url.QueryEscape(param[1]))
	}
	return buff.String()
}

func awsSignString(query string, host string, dateLongStr string, alg string, scopeStr string) string {
	req := strings.Join([]string{
		"GET",
		"/mqtt",
		query,
		"host:" + host,
		"", // separator
		"host",
		emptyStringHash,
	}, "\n")
	return strings.Join([]string{
		alg,
		dateLongStr,
		scopeStr,
		awsSha(req),
	}, "\n")
}

func awsHmac(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func awsSignKey(secretKey string, dateShort string, region string, serviceName string) []byte {
	h := awsHmac([]byte("AWS4"+secretKey), []byte(dateShort))
	h = awsHmac(h, []byte(region))
	h = awsHmac(h, []byte(serviceName))
	h = awsHmac(h, []byte("aws4_request"))
	return h
}

func awsSha(in string) string {
	h := sha256.New()
	h.Write([]byte(in))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// rootCa This is the root CA certificate from https://www.amazontrust.com/repository/AmazonRootCA1.pem
const rootCa = `-----BEGIN CERTIFICATE-----
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

	if ok := certificatePool.AppendCertsFromPEM([]byte(rootCa)); !ok {
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
