package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

func main() {
    // Path to the connection profile
    configPath := "config/connection-profile.json"
    
    // Read the file
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        os.Exit(1)
    }
    
    // Try to parse it as JSON
    var config map[string]interface{}
    err = json.Unmarshal(data, &config)
    if err != nil {
        fmt.Printf("Error parsing JSON: %v\n", err)
        // Print first 100 characters to see where the error might be
        fmt.Printf("Start of file: %s\n", string(data[:100]))
        os.Exit(1)
    }
    
    fmt.Println("JSON is valid!")
}
