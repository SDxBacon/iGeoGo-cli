package main

import (
	"bufio"
	"context"
	"fmt"
	"iGeoGo-cli/simulatelocation"
	"iGeoGo-cli/tunnel"
	"iGeoGo-cli/utils"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func logInputHelp() {
	fmt.Println("Enter GPS location or command:")
	fmt.Println("  - GPS: 'lat,lng' or '[lat,lng]' or '[[lat,lng], ...]'")
	fmt.Println("  - File: 'read filename.txt'")
	fmt.Println("  - Exit: 'quit' or Ctrl+C'")
}

func startSimulateLocation(service *simulatelocation.SimlulateLocationService, coordinates []utils.Coordinate) {
	go func() {
		if len(coordinates) < 2 {
			_, err := service.Set(coordinates[0])
			if err != nil {
				log.Fatalf("Failed to start location simulation: %v", err)
			}

		} else {
			// 模擬騎行（單位: km/h）
			speed := 15.0

			// 定義位置更新的回調函數
			onPositionUpdate := func(coord utils.Coordinate, index int, totalDistance float64) {
				fmt.Printf("[位置更新] 點 %d: (%.6f, %.6f), 已騎行 %.2f 公尺\n",
					index, coord.Lng, coord.Lat, totalDistance)

				// 呼叫 simulateLocation 函數來模擬位置更新
				log.Printf("Starting location simulation...")
				_, err := service.Set(coord)
				if err != nil {
					log.Fatalf("Failed to start location simulation: %v", err)
				}
			}

			// 開始模擬騎行！
			utils.SimulateBikeRide(coordinates, speed, onPositionUpdate)
		}

		// 完成模擬，再 log 一個 helper 訊息
		logInputHelp()
	}()
}

func main() {
	// 建立 IGeoGoIosTunnel instance 並啟動
	tun := tunnel.NewIGeoGoIosTunnel()
	err := tun.Start()
	if err != nil {
		panic(err)
	}

	// 取得第一個連接的設備
	device, err := utils.GetFirstDevice()
	if err != nil {
		log.Fatalf("Failed to get first device: %v", err)
	}

	// 檢查設備是否受支援
	if !utils.IsDeviceSupported(device) {
		log.Fatalf("Device %s is not supported", device.Properties.SerialNumber)
	}

	// 建立 context 用於優雅關閉
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 建立 scanner 和 input channel
	scanner := bufio.NewScanner(os.Stdin)
	inputChan := make(chan string)

	// 在背景 goroutine 中讀取 stdin
	go func() {
		defer close(inputChan)

		// 顯示輸入說明
		logInputHelp()

		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Scanner error: %v", err)
		}
	}()

	// 主循環：無止盡收 scanner 或 ctx.Done
	var service *simulatelocation.SimlulateLocationService
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nExiting...")
			return
		case input, ok := <-inputChan:
			if !ok {
				return
			}
			if input == "quit" || input == "exit" {
				fmt.Println("Exiting...")
				return
			}

			// 如果 Simulate Location Service 還沒建立，就先建立一次
			if service == nil {
				service = simulatelocation.NewSimulateLocationService(device)
				defer service.Destory() // 確保在 main 函數結束時停止模擬
			}

			var coords []utils.Coordinate
			var err error

			// 檢查是否為 read 指令
			if strings.HasPrefix(input, "read ") {
				filename := strings.TrimSpace(strings.TrimPrefix(input, "read"))
				fmt.Printf("Reading GPS coordinates from file: %s\n", filename)
				coords, err = utils.GetCoordinatesFromGeoJSON(filename)
				if err != nil {
					log.Printf("Failed to read file: %v", err)
					continue
				}
			} else {
				// 直接解析 GPS 座標
				coords, err = utils.ParseGPSInput(input)
				if err != nil {
					log.Printf("Failed to parse GPS input: %v", err)
					continue
				}
			}

			fmt.Printf("Parsed %d coordinate(s):\n", len(coords))
			// for i, coord := range coords {
			// 	fmt.Printf("  %d. Lat: %.6f, Lng: %.6f\n", i+1, coord.Lat, coord.Lng)
			// }
			// TODO: 將座標設定到 iOS 設備

			startSimulateLocation(service, coords)
		}
	}
}
