package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// IOTDataChaincode provides IoT data storage and retrieval
type IOTDataChaincode struct {
	contractapi.Contract
}

// TemperatureReading represents a single temperature measurement
type TemperatureReading struct {
	ReadingID   string  `json:"readingID"`
	DeviceID    string  `json:"deviceID"`
	Temperature float64 `json:"temperature"`
	Timestamp   int64   `json:"timestamp"`
	SessionID   string  `json:"sessionID"` // Session ID from ISV
	Unit        string  `json:"unit"`      // "C" or "F"
	Status      string  `json:"status"`    // "normal", "anomaly"
}

// DeviceStatistics represents aggregated stats for a device
type DeviceStatistics struct {
	DeviceID       string  `json:"deviceID"`
	ReadingCount   int     `json:"readingCount"`
	MinTemperature float64 `json:"minTemperature"`
	MaxTemperature float64 `json:"maxTemperature"`
	AvgTemperature float64 `json:"avgTemperature"`
	LastReading    int64   `json:"lastReading"`
	FirstReading   int64   `json:"firstReading"`
}

// InitLedger initializes the chaincode
func (s *IOTDataChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Initializing IOT-DATA Chaincode")
	return nil
}

// StoreTemperature stores a temperature reading
func (s *IOTDataChaincode) StoreTemperature(ctx contractapi.TransactionContextInterface, deviceID string, temperature float64, timestamp int64, sessionID string) error {
	// Validate inputs
	if len(deviceID) < 3 || len(deviceID) > 64 {
		return fmt.Errorf("invalid deviceID length")
	}

	if temperature < -50 || temperature > 100 {
		return fmt.Errorf("temperature out of valid range (-50 to 100°C)")
	}

	// Validate timestamp (must be within 5 minutes)
	currentTime := getCurrentTimestamp()
	if timestamp < currentTime-300 || timestamp > currentTime+300 {
		return fmt.Errorf("timestamp is invalid or too old/future")
	}

	// Verify device exists in USER-ACL chaincode (cross-chaincode call)
	deviceExists, err := s.verifyDeviceExists(ctx, deviceID)
	if err != nil || !deviceExists {
		return fmt.Errorf("device %s not registered in USER-ACL: %v", deviceID, err)
	}

	// Verify session is valid via ISV chaincode (cross-chaincode call)
	// In production, this should call ISV to validate session
	if len(sessionID) < 5 {
		return fmt.Errorf("invalid session ID")
	}

	// Generate unique reading ID
	readingID := fmt.Sprintf("READING_%s_%d", deviceID, timestamp)

	// Detect anomaly (simple rule: > 28°C or < 18°C)
	status := "normal"
	if temperature > 28.0 || temperature < 18.0 {
		status = "anomaly"
	}

	// Create reading
	reading := TemperatureReading{
		ReadingID:   readingID,
		DeviceID:    deviceID,
		Temperature: temperature,
		Timestamp:   timestamp,
		SessionID:   sessionID,
		Unit:        "C",
		Status:      status,
	}

	readingJSON, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %v", err)
	}

	// Store reading
	err = ctx.GetStub().PutState(readingID, readingJSON)
	if err != nil {
		return fmt.Errorf("failed to store reading: %v", err)
	}

	// Update device statistics
	err = s.updateDeviceStatistics(ctx, deviceID, temperature, timestamp)
	if err != nil {
		log.Printf("Warning: failed to update statistics: %v", err)
		// Don't fail the transaction if stats update fails
	}

	// Emit event
	eventData := map[string]interface{}{
		"deviceID":    deviceID,
		"temperature": temperature,
		"timestamp":   timestamp,
		"status":      status,
	}
	eventJSON, _ := json.Marshal(eventData)
	err = ctx.GetStub().SetEvent("TemperatureStored", eventJSON)
	if err != nil {
		return fmt.Errorf("failed to emit event: %v", err)
	}

	if status == "anomaly" {
		log.Printf("⚠️  ANOMALY DETECTED: Device %s reported %s°C at %d", deviceID, fmt.Sprintf("%.1f", temperature), timestamp)
	} else {
		log.Printf("Temperature stored: Device %s, %.1f°C, Session %s", deviceID, temperature, sessionID)
	}

	return nil
}

// GetDeviceReadings retrieves temperature readings for a device within time range
func (s *IOTDataChaincode) GetDeviceReadings(ctx contractapi.TransactionContextInterface, deviceID string, startTime int64, endTime int64) (string, error) {
	// Validate inputs
	if endTime == 0 {
		endTime = getCurrentTimestamp()
	}
	if startTime == 0 {
		startTime = endTime - 86400 // Default to last 24 hours
	}

	// Query readings by range
	startKey := fmt.Sprintf("READING_%s_%d", deviceID, startTime)
	endKey := fmt.Sprintf("READING_%s_%d", deviceID, endTime)

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return "", fmt.Errorf("failed to query readings: %v", err)
	}
	defer resultsIterator.Close()

	var readings []TemperatureReading
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var reading TemperatureReading
		err = json.Unmarshal(queryResponse.Value, &reading)
		if err != nil {
			continue
		}

		// Filter by deviceID (in case range query includes other devices)
		if reading.DeviceID == deviceID {
			readings = append(readings, reading)
		}
	}

	readingsJSON, err := json.Marshal(readings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal readings: %v", err)
	}

	return string(readingsJSON), nil
}

// GetLatestReading retrieves the most recent temperature reading for a device
func (s *IOTDataChaincode) GetLatestReading(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	// Get all readings for device (last 24 hours)
	endTime := getCurrentTimestamp()
	startTime := endTime - 86400

	readingsJSON, err := s.GetDeviceReadings(ctx, deviceID, startTime, endTime)
	if err != nil {
		return "", err
	}

	var readings []TemperatureReading
	json.Unmarshal([]byte(readingsJSON), &readings)

	if len(readings) == 0 {
		return "", fmt.Errorf("no readings found for device %s", deviceID)
	}

	// Find latest reading
	latestReading := readings[0]
	for _, reading := range readings {
		if reading.Timestamp > latestReading.Timestamp {
			latestReading = reading
		}
	}

	latestJSON, err := json.Marshal(latestReading)
	if err != nil {
		return "", fmt.Errorf("failed to marshal latest reading: %v", err)
	}

	return string(latestJSON), nil
}

// GetLatestReadings retrieves the most recent N readings from all devices
func (s *IOTDataChaincode) GetLatestReadings(ctx contractapi.TransactionContextInterface, limit int) (string, error) {
	if limit <= 0 || limit > 100 {
		limit = 10 // Default to 10
	}

	// Query all readings
	resultsIterator, err := ctx.GetStub().GetStateByRange("READING_", "READING_~")
	if err != nil {
		return "", fmt.Errorf("failed to query readings: %v", err)
	}
	defer resultsIterator.Close()

	var readings []TemperatureReading
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var reading TemperatureReading
		err = json.Unmarshal(queryResponse.Value, &reading)
		if err != nil {
			continue
		}

		readings = append(readings, reading)
	}

	// Sort by timestamp (descending) and take top N
	// Simple bubble sort (for production, use more efficient sorting)
	for i := 0; i < len(readings)-1; i++ {
		for j := i + 1; j < len(readings); j++ {
			if readings[j].Timestamp > readings[i].Timestamp {
				readings[i], readings[j] = readings[j], readings[i]
			}
		}
	}

	// Take top N
	if len(readings) > limit {
		readings = readings[:limit]
	}

	readingsJSON, err := json.Marshal(readings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal readings: %v", err)
	}

	return string(readingsJSON), nil
}

// GetDeviceStatistics retrieves aggregated statistics for a device
func (s *IOTDataChaincode) GetDeviceStatistics(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	statsKey := fmt.Sprintf("STATS_%s", deviceID)
	statsJSON, err := ctx.GetStub().GetState(statsKey)
	if err != nil {
		return "", fmt.Errorf("failed to read statistics: %v", err)
	}

	if statsJSON == nil {
		// No stats yet, return empty stats
		stats := DeviceStatistics{
			DeviceID:       deviceID,
			ReadingCount:   0,
			MinTemperature: 0,
			MaxTemperature: 0,
			AvgTemperature: 0,
			LastReading:    0,
			FirstReading:   0,
		}
		statsJSON, _ = json.Marshal(stats)
	}

	return string(statsJSON), nil
}

// GetAllDeviceStats retrieves statistics for all devices
func (s *IOTDataChaincode) GetAllDeviceStats(ctx contractapi.TransactionContextInterface) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("STATS_", "STATS_~")
	if err != nil {
		return "", fmt.Errorf("failed to query statistics: %v", err)
	}
	defer resultsIterator.Close()

	var allStats []DeviceStatistics
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var stats DeviceStatistics
		err = json.Unmarshal(queryResponse.Value, &stats)
		if err != nil {
			continue
		}

		allStats = append(allStats, stats)
	}

	statsJSON, err := json.Marshal(allStats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal statistics: %v", err)
	}

	return string(statsJSON), nil
}

// GetAnomalies retrieves all anomalous readings
func (s *IOTDataChaincode) GetAnomalies(ctx contractapi.TransactionContextInterface, limit int) (string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Query all readings
	resultsIterator, err := ctx.GetStub().GetStateByRange("READING_", "READING_~")
	if err != nil {
		return "", fmt.Errorf("failed to query readings: %v", err)
	}
	defer resultsIterator.Close()

	var anomalies []TemperatureReading
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var reading TemperatureReading
		err = json.Unmarshal(queryResponse.Value, &reading)
		if err != nil {
			continue
		}

		if reading.Status == "anomaly" {
			anomalies = append(anomalies, reading)
		}

		if len(anomalies) >= limit {
			break
		}
	}

	anomaliesJSON, err := json.Marshal(anomalies)
	if err != nil {
		return "", fmt.Errorf("failed to marshal anomalies: %v", err)
	}

	return string(anomaliesJSON), nil
}

// Helper functions

// verifyDeviceExists checks if device exists in USER-ACL chaincode
func (s *IOTDataChaincode) verifyDeviceExists(ctx contractapi.TransactionContextInterface, deviceID string) (bool, error) {
	// Cross-chaincode call to USER-ACL
	// In production implementation:
	/*
		response := ctx.GetStub().InvokeChaincode(
			"user-acl",
			[][]byte{[]byte("GetDevice"), []byte(deviceID)},
			"authchannel",
		)

		if response.Status != shim.OK {
			return false, fmt.Errorf("device not found")
		}
	*/

	// Simplified validation for now
	if len(deviceID) >= 3 {
		return true, nil
	}

	return false, fmt.Errorf("invalid device ID")
}

// updateDeviceStatistics updates aggregated statistics for a device
func (s *IOTDataChaincode) updateDeviceStatistics(ctx contractapi.TransactionContextInterface, deviceID string, temperature float64, timestamp int64) error {
	statsKey := fmt.Sprintf("STATS_%s", deviceID)

	// Get existing stats
	statsJSON, err := ctx.GetStub().GetState(statsKey)

	var stats DeviceStatistics

	if statsJSON == nil {
		// First reading for this device
		stats = DeviceStatistics{
			DeviceID:       deviceID,
			ReadingCount:   1,
			MinTemperature: temperature,
			MaxTemperature: temperature,
			AvgTemperature: temperature,
			LastReading:    timestamp,
			FirstReading:   timestamp,
		}
	} else {
		// Update existing stats
		err = json.Unmarshal(statsJSON, &stats)
		if err != nil {
			return err
		}

		// Update count
		stats.ReadingCount++

		// Update min/max
		if temperature < stats.MinTemperature {
			stats.MinTemperature = temperature
		}
		if temperature > stats.MaxTemperature {
			stats.MaxTemperature = temperature
		}

		// Update average (running average)
		stats.AvgTemperature = ((stats.AvgTemperature * float64(stats.ReadingCount-1)) + temperature) / float64(stats.ReadingCount)
		stats.AvgTemperature = math.Round(stats.AvgTemperature*10) / 10 // Round to 1 decimal

		// Update last reading
		if timestamp > stats.LastReading {
			stats.LastReading = timestamp
		}
	}

	// Store updated stats
	statsJSON, err = json.Marshal(stats)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(statsKey, statsJSON)
	if err != nil {
		return err
	}

	return nil
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func main() {
	chaincode, err := contractapi.NewChaincode(&IOTDataChaincode{})
	if err != nil {
		log.Panicf("Error creating IOT-DATA chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting IOT-DATA chaincode: %v", err)
	}
}
