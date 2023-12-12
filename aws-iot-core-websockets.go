package aws_iot_core_websockets_go

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type IotWsConfig struct {
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	Endpoint     string
}

const (
	emptyStringHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

func AwsIotWsUrl(iotWsConfig IotWsConfig) string {
	// according to docs, time must be within 5min of actual time (or at least according to AWS servers)
	now := time.Now().UTC()

	dateLong := now.Format("20060102T150405Z")
	dateShort := dateLong[:8]
	serviceName := "iotdevicegateway"
	scope := fmt.Sprintf("%s/%s/%s/aws4_request", dateShort, iotWsConfig.Region, serviceName)
	alg := "AWS4-HMAC-SHA256"
	q := [][2]string{
		{"X-Amz-Algorithm", alg},
		{"X-Amz-Credential", iotWsConfig.AccessKey + "/" + scope},
		{"X-Amz-Date", dateLong},
		{"X-Amz-SignedHeaders", "host"},
	}

	query := awsQueryParams(q)

	signKey := awsSignKey(iotWsConfig.SecretKey, dateShort, iotWsConfig.Region, serviceName)
	stringToSign := awsSignString(query, iotWsConfig.Endpoint, dateLong, alg, scope)
	signature := fmt.Sprintf("%x", awsHmac(signKey, []byte(stringToSign)))

	wsurl := fmt.Sprintf("wss://%s/mqtt?%s&X-Amz-Signature=%s", iotWsConfig.Endpoint, query, signature)

	if iotWsConfig.SessionToken != "" {
		wsurl = fmt.Sprintf("%s&X-Amz-Security-Token=%s", wsurl, url.QueryEscape(iotWsConfig.SessionToken))
	}

	return wsurl
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
	fmt.Fprintf(h, "%s", in)
	return fmt.Sprintf("%x", h.Sum(nil))
}
