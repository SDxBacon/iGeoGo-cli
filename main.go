package main

import (
	"bufio"
	"context"
	"fmt"
	"iGeoGo-cli/tunnel"
	"iGeoGo-cli/utils"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

func logInputHelp() {
	fmt.Println("Enter GPS location or command:")
	fmt.Println("  - GPS: 'lat,lng' or '[lat,lng]' or '[[lat,lng], ...]'")
	fmt.Println("  - File: 'read filename.txt'")
	fmt.Println("  - Exit: 'quit' or Ctrl+C'")
}

func startSimulateLocation(device ios.DeviceEntry, coordinates []utils.Coordinate) {
	go func() {
		// create rsd device connection
		tun, err := tunnel.GetTunnelForDevice(device.Properties.SerialNumber)
		if err != nil {
			log.Printf("Failed to get tunnel for device %s: %v", device.Properties.SerialNumber, err)
			return
		}
		rsdDevice, err := utils.NewDeviceWithRsd(device, tun)
		if err != nil {
			log.Printf("Failed to create RSD device for %s: %v", device.Properties.SerialNumber, err)
			return
		}

		// create location simulation service
		service, err := instruments.NewLocationSimulationService(rsdDevice)
		if err != nil {
			log.Fatalf("Failed to create location service: %v\nMake sure Developer Mode is enabled on device", err)
		}
		defer service.StopSimulateLocation()

		if len(coordinates) < 2 {
			err = service.StartSimulateLocation(coordinates[0].Lat, coordinates[0].Lng)
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
				err = service.StartSimulateLocation(coord.Lat, coord.Lng)
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

			startSimulateLocation(device, coords)
		}
	}
}
