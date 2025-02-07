package main

import (
	"common/pkg/http_client"
	"common/pkg/http_client/interceptors"
	"common/pkg/logger"
	"context"
	"fmt"
	"net/http"
)

func main() {

	client := http_client.New(http_client.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		GlobalHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		// Interceptor: &interceptors.LoggingInterceptor{
		// 	Next: http.DefaultTransport,
		// },
		Interceptor: interceptors.NewLoggerInterceptor(
			http.DefaultTransport,
			nil,
			// interceptors.LoggingOptions{
			// 	LogRequestBody: true,
			// 	LogHeaders:     true,
			// },
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

	// example-3
	builder := client.Post(context.TODO(), "https://jsonplaceholder.typicode.com/posts").
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"name": "test"})

	// Get response when needed
	resp, err = builder.Result()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	// example-4
	client.Get(context.TODO(), "/posts/1").
		SetHeader("Authorization", "Bearer token").
		SetQueryParam("filter", "active").
		OnSuccess(func(res *http_client.Response) {
			fmt.Println("Data received:", string(res.Body))
		}).
		OnError(func(err *http_client.RequestError) {
			fmt.Println("Request failed:", err)
		})

	// Upload a single file
	resp, err = client.Post(context.Background(), "/upload").
		AddFile("file", "/path/to/file.txt").
		Result()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("response========>", string(resp.Body))

	// Upload multiple files with form data
	resp, err = client.Post(context.Background(), "/upload").
		SetBody(map[string]interface{}{
			"description": "My files",
		}).
		AddFile("file1", "/path/to/file1.txt").
		AddFile("file2", "/path/to/file2.txt").
		Result()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	client.SetBearerToken("Bearer token")
	client.WithBasicAuth("username", "password")

	// Example-5
	var user User
	err = client.Get(context.Background(), "http://jsonplaceholder.typicode.com/posts/1").Into(&user)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("response========>", user)

	// Example-6 custom logger interceptor
	myLoggerSrv := logger.New()
	client2 := http_client.New(http_client.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Interceptor: interceptors.NewLoggerInterceptor(
			nil,
			// interceptors.NewAuthInterceptor(nil),
			myLoggerSrv,
			// nil,
			&interceptors.LoggingOptions{
				LogRequestBody: true,
				LogHeaders:     true,
			}),
	})

	var user2 User
	err = client2.Get(context.Background(), "/posts/1").Into(&user2)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("response========>", user)
}

type User struct {
	UserId int    `json:"userId"`
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}
