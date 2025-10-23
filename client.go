// Client CLI utility for Chrono-DB
// Build: go build -o chrono-client client.go
// Usage: ./chrono-client <command> [options]

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	baseURL = flag.String("url", "http://localhost:8080", "Chrono-DB API URL")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Args()[0]

	switch command {
	case "insert":
		if len(flag.Args()) < 3 {
			fmt.Println("Usage: client insert <key> <value>")
			os.Exit(1)
		}
		key := flag.Args()[1]
		value := flag.Args()[2]
		insertData(key, value)

	case "query":
		if len(flag.Args()) < 2 {
			fmt.Println("Usage: client query <key>")
			os.Exit(1)
		}
		key := flag.Args()[1]
		queryData(key)

	case "history":
		if len(flag.Args()) < 2 {
			fmt.Println("Usage: client history <key>")
			os.Exit(1)
		}
		key := flag.Args()[1]
		getHistory(key)

	case "status":
		getStatus()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Chrono-DB CLI Client")
	fmt.Println("\nUsage: client [options] <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  insert <key> <value>  - Insert a key-value pair")
	fmt.Println("  query <key>          - Query current value for a key")
	fmt.Println("  history <key>        - Get full history for a key")
	fmt.Println("  status               - Get cluster status")
	fmt.Println("\nOptions:")
	fmt.Println("  -url string          - API URL (default: http://localhost:8080)")
}

func insertData(key, value string) {
	data := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*baseURL+"/api/v1/insert", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
}

func queryData(key string) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?key=%s", *baseURL, key))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		if found, ok := result["found"].(bool); ok && found {
			fmt.Printf("Key: %s\n", result["key"])
			fmt.Printf("Value: %v\n", result["value"])
		} else {
			fmt.Printf("Key not found: %s\n", key)
		}
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
}

func getHistory(key string) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/history?key=%s", *baseURL, key))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		jsonStr, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonStr))
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
}

func getStatus() {
	resp, err := http.Get(*baseURL + "/api/v1/status")
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		jsonStr, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonStr))
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
}
