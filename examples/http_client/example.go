package main

import (
	"common/examples/http_client/interceptors"
	"common/pkg/http_client"
	"context"
	"fmt"
	"net/http"
)

func main() {
	logger := interceptors.NewStandardLogger()

	client := http_client.New(http_client.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		GlobalHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		// Interceptor: &interceptors.LoggingInterceptor{
		// 	Next: http.DefaultTransport,
		// },
		Interceptor: interceptors.NewLoggingInterceptor(
			http.DefaultTransport,
			logger,
			interceptors.LoggingOptions{
				LogRequestBody: true,
				LogHeaders:     true,
			},
		),
	})

	// example 1
	resp, err := client.Get(context.TODO(), "/todos/1").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("response========>", string(resp.Body))

	// example 2
	var responseData map[string]interface{}
	err = client.Get(context.TODO(), "/todos/1").Into(&responseData)
	if err != nil {
		panic(err)
	}

	fmt.Println("response========>", responseData)

	// example 3
	res, err := client.Post(context.TODO(), "/todos").
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"title": "foo"}).
		Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("response========>", string(res.Body))

}
