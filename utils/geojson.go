package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// GeoJSON 結構定義
type GeoJSON struct {
	Type     string   `json:"type"`
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	Type        string        `json:"type"`
	Coordinates []interface{} `json:"coordinates"`
}

// ReadGeoJSON 讀取 GeoJSON 檔案
func readGeoJSON(filename string) (*GeoJSON, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("無法讀取檔案: %v", err)
	}

	var geojson GeoJSON
	err = json.Unmarshal(data, &geojson)
	if err != nil {
		return nil, fmt.Errorf("無法解析 JSON: %v", err)
	}

	return &geojson, nil
}

// ParseCoordinates 解析座標陣列
func ParseCoordinates(coords []interface{}) ([]Coordinate, error) {
	var coordinates []Coordinate

	for _, coord := range coords {
		switch v := coord.(type) {
		case []interface{}:
			// 處理 [longitude, latitude] 格式
			if len(v) >= 2 {
				lng, ok1 := v[0].(float64)
				lat, ok2 := v[1].(float64)
				if ok1 && ok2 {
					coordinates = append(coordinates, Coordinate{
						Lng: lng,
						Lat: lat,
					})
				}
			}
		}
	}

	return coordinates, nil
}

// CalculateDistance 計算兩點之間的距離（使用 Haversine 公式，單位：公尺）
func CalculateDistance(coord1, coord2 Coordinate) float64 {
	const earthRadius = 6371000 // 地球半徑（公尺）

	lat1 := coord1.Lat * math.Pi / 180
	lat2 := coord2.Lat * math.Pi / 180
	deltaLat := (coord2.Lat - coord1.Lat) * math.Pi / 180
	deltaLng := (coord2.Lng - coord1.Lng) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func GetCoordinatesFromGeoJSON(filename string) ([]Coordinate, error) {
	geojson, err := readGeoJSON(filename)
	if err != nil {
		return nil, fmt.Errorf("無法讀取 GeoJSON: %v", err)
	}

	if len(geojson.Geometry.Coordinates) == 0 {
		return nil, fmt.Errorf("沒有找到任何 coordinates")
	}

	coordinates, err := ParseCoordinates(geojson.Geometry.Coordinates)
	if err != nil {
		return nil, fmt.Errorf("解析座標錯誤: %v", err)
	}

	return coordinates, nil
}
