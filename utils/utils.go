package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
)

// Coordinate 代表一個 GPS 座標 (緯度, 經度)
type Coordinate struct {
	Lat float64
	Lng float64
}

// ParseGPSInput 解析使用者輸入的 GPS 座標
// 支援三種格式:
// 1. "lat,lng" - 例如: "25.033,121.564"
// 2. "[lat,lng]" - 例如: "[25.033,121.564]"
// 3. "[[lat,lng], ...]" - 例如: "[[25.033,121.564], [24.137,120.686]]"
func ParseGPSInput(input string) ([]Coordinate, error) {
	input = strings.TrimSpace(input)

	// 檢查是否為 JSON 陣列格式
	if strings.HasPrefix(input, "[") {
		return parseJSONFormat(input)
	}

	// 簡單的 "lat,lng" 格式
	return parseSimpleFormat(input)
}

// parseJSONFormat 解析 JSON 格式的座標
func parseJSONFormat(input string) ([]Coordinate, error) {
	// 嘗試解析為二維陣列 [[lat,lng], ...]
	var coords [][]float64
	if err := json.Unmarshal([]byte(input), &coords); err == nil {
		result := make([]Coordinate, 0, len(coords))
		for i, coord := range coords {
			if len(coord) != 2 {
				return nil, fmt.Errorf("invalid coordinate format at index %d: expected [lat,lng]", i)
			}
			result = append(result, Coordinate{Lat: coord[0], Lng: coord[1]})
		}
		return result, nil
	}

	// 嘗試解析為一維陣列 [lat,lng]
	var coord []float64
	if err := json.Unmarshal([]byte(input), &coord); err == nil {
		if len(coord) != 2 {
			return nil, fmt.Errorf("invalid coordinate format: expected [lat,lng]")
		}
		return []Coordinate{{Lat: coord[0], Lng: coord[1]}}, nil
	}

	return nil, fmt.Errorf("invalid JSON format")
}

// parseSimpleFormat 解析簡單的 "lat,lng" 格式
func parseSimpleFormat(input string) ([]Coordinate, error) {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format: expected 'lat,lng'")
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	// 驗證座標範圍
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90, got %v", lat)
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180, got %v", lng)
	}

	return []Coordinate{{Lat: lat, Lng: lng}}, nil
}

func GetFirstDevice() (ios.DeviceEntry, error) {
	devices, err := ios.ListDevices()
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
		return ios.DeviceEntry{}, err
	}

	if len(devices.DeviceList) > 0 {
		fmt.Printf("Found %d device(s):\n\n", len(devices.DeviceList))
		for i, device := range devices.DeviceList {
			fmt.Printf("%d. %s\n", i+1, device.Properties.SerialNumber)
			fmt.Printf("   UDID: %s\n", device.Properties.SerialNumber)
		}
	} else {
		fmt.Println("No devices connected")
	}

	return devices.DeviceList[0], err
}

func IsDeviceSupported(device ios.DeviceEntry) bool {
	version, err := ios.GetProductVersion(device)
	if err != nil {
		log.Printf("Failed to get product version for device %s: %v", device.Properties.SerialNumber, err)
		return false
	}

	return version.Major() >= 17 // 假設我們需要 iOS 17 或更高版本
}
