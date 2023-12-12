# AWS IoT Core WebSockets support for Go (Golang)

Connecting to AWS IoT Core with certificates is great. But sometimes you just want to use WebSockets with your AWS
credentials. This library makes it easy to do that.

[If you need something similar in Java I wrote that a few years back while at AWS and have my own personal fork of it](https://github.com/timmattison/aws-iot-core-websockets-java).

## This is not my code originally!

[This code was posted in a GitHub issue years ago](https://github.com/aws/aws-sdk-go/issues/2485#issuecomment-469366846).
I created a similar library for Java when I was working at AWS and I wanted to do the same for Go.

# Version history

## 0.4.0

**TL;DR - new pattern to get the MQTT config directly**

Switched to functional options pattern so the setup is easier to understand and use. Everything other
than the context can be set up automatically for the user.

## 0.3.1

**TL;DR - new functions to get the MQTT config directly**

Added `AwsIotWsMqttOptionsCustom` and `AwsIotWsMqttOptions` to get the MQTT options for AWS IoT Core.

`AwsIotWsMqttOptionsCustom` allows the user to specify their own certificate pool.

`AwsIotWsMqttOptions` uses the AWS root CA certificate automatically.

Both new functions will retrieve the endpoint automatically if it is not specified in the `IotWsConfig` struct.

## 0.3.0

**TL;DR - pass in the AWS.Config and endpoint**

Passing in the AWS.Config instead of having the user copy the data out makes more sense.

## 0.2.0

**TL;DR - now just pass in the full endpoint DNS name, not just the host**

Using the true endpoint value returned from AWS IoT's describe endpoint call instead of using the host name (the
customer specific hostname) and then building up the URL from that.

## 0.1.0

**TL;DR - now pass in the IotWsConfig struct**

Created a struct to pass in the parameters to avoid getting them out of order

## 0.0.1

Original version from https://github.com/aws/aws-sdk-go/issues/2485#issuecomment-469366846 with some tests but no modifications.
