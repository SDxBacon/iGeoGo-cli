package utils

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// SimulateBikeRide 模擬騎腳踏車移動
// speedKmH: 速度（公里/小時）
// onPositionUpdate: 回調函數，每次位置更新時調用
// 替換原本的 SimulateBikeRide 函數
func SimulateBikeRide(coordinates []Coordinate, speedKmH float64, onPositionUpdate func(Coordinate, int, float64)) {
	if len(coordinates) < 2 {
		fmt.Println("座標點數量不足")
		return
	}

	totalDistance := 0.0

	fmt.Printf("開始模擬騎行，基準速度: %.1f km/h\n", speedKmH)

	for i := 0; i < len(coordinates)-1; i++ {
		currentCoord := coordinates[i]
		nextCoord := coordinates[i+1]

		// 計算兩點間距離
		distance := CalculateDistance(currentCoord, nextCoord)
		segmentStartDistance := totalDistance // 這段開始前的累積距離
		totalDistance += distance

		// 加入隨機變化 (-20% ~ +20%)
		randomFactor := 0.8 + rand.Float64()*0.4 // 0.8 到 1.2
		actualSpeed := speedKmH * randomFactor
		speedMeterPerSecond := actualSpeed * 1000 / 3600

		// 計算需要的時間（秒）
		travelTime := distance / speedMeterPerSecond

		fmt.Printf("點 %d -> %d: 距離 %.2f 公尺, 實際速度 %.2f km/h, 預計 %.2f 秒\n",
			i, i+1, distance, actualSpeed, travelTime)

		// 分段移動，每 1~2 秒更新一次位置
		elapsedTime := 0.0
		for elapsedTime < travelTime {
			// 隨機 tick 間隔 (1~2 秒)
			tickDuration := 1.0 + rand.Float64() // 1.0 到 2.0 秒

			// 如果剩餘時間不足一個完整 tick，就使用剩餘時間
			if elapsedTime+tickDuration > travelTime {
				tickDuration = travelTime - elapsedTime
			}

			elapsedTime += tickDuration

			// 計算當前進度（0.0 到 1.0）
			progress := elapsedTime / travelTime

			// 線性插值計算當前位置
			currentLat := currentCoord.Lat + (nextCoord.Lat-currentCoord.Lat)*progress
			currentLng := currentCoord.Lng + (nextCoord.Lng-currentCoord.Lng)*progress

			interpolatedCoord := Coordinate{
				Lat: currentLat,
				Lng: currentLng,
			}

			// 當前位置回調
			if onPositionUpdate != nil {
				currentTotalDistance := segmentStartDistance + distance*progress
				onPositionUpdate(interpolatedCoord, i, currentTotalDistance)
			}

			// 休眠這個 tick 的時間
			time.Sleep(time.Duration(tickDuration*1000) * time.Millisecond)
		}
	}

	// 最後一個點
	if onPositionUpdate != nil {
		onPositionUpdate(coordinates[len(coordinates)-1], len(coordinates)-1, totalDistance)
	}

	fmt.Printf("騎行完成！總距離: %.2f 公尺 (%.2f 公里)\n",
		totalDistance, totalDistance/1000)
}
