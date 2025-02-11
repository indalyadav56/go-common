package main

import (
	"common/pkg/http_client"
	"fmt"

	"golang.org/x/net/context"
)

func main() {
	http_client := http_client.New(http_client.Config{
		BaseURL: "http://localhost:4100",
	})

	http_client.SetGlobalHeaders(map[string]string{
		"Content-Type": "application/json",
		"X-Device-Id":  "123",
		"X-Device-Ip":  "test",
		"X-Lat-Long":   "test,test",
		"X-OS":         "android",
		"X-OS-Version": "c91adcbcef6b8ee12",
	})

	http_client.SetBearerToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDVzNzkyNGUtY2VkOC00fDc3ZXpoNSs9P0owczpjaGlaLS8kQTloI0BEVFBnUUtsIiwiZGV2aWNlX2luZm8iOiJQb3N0bWFuUnVudGltZS83LjQzLjAiLCJ1c2VyX2lwIjoidGVzdCIsImlzcyI6InBheWRvaC1iYW5rIiwiZXhwIjoxNzUxMjU1MjIxfQ.ID5pS54Dok0pGvx9rA-ko8sUOgkxyQW7DYCe0trfuNE")

	res, err := http_client.Post(context.Background(), "/api/upi/simbinding/sms-verification").
		SetBody(map[string]interface{}{
			"data": "77f57e53de75bfb0676f2877f64903b5916a0645e8f094a30d3eea40f0f07364db6ea740db2db7b9",
		}).SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"X-Device-Id":  "123",
		"X-Device-Ip":  "test",
		"X-Lat-Long":   "test,test",
		"X-OS":         "android",
		"X-OS-Version": "c91adcbcef6b8ee12",
	}).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res.Body))

	res, err = http_client.Post(context.Background(), "/api/upi/remapping-upi-id").SetBody(map[string]interface{}{"data": "838b12e6f7793f8781a50e204187f6b586f3cdff9e3fef650a8b3e4223b9027a97435f33123f4101e0f10e4da52ab7c7"}).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res.Body))
}
