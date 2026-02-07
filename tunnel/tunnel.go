package tunnel

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	iostunnel "github.com/danielpaulus/go-ios/ios/tunnel"
)

type IGeoGoIosTunnel struct {
	PairRecordPath         string // pair record 儲存路徑
	TunnelInfoPort         int    // tunnel info API 的 port (預設 28100)
	UseUserspaceNetworking bool   // 是否使用 userspace networking (相當於 --userspace flag)
}

// constructor
func NewIGeoGoIosTunnel() *IGeoGoIosTunnel {
	return &IGeoGoIosTunnel{
		PairRecordPath:         ".",
		TunnelInfoPort:         28100,
		UseUserspaceNetworking: true,
	}
}

func (t *IGeoGoIosTunnel) Start() error {
	// 建立 PairRecordManager
	pm, err := iostunnel.NewPairRecordManager(t.PairRecordPath)
	if err != nil {
		return fmt.Errorf("could not create pair record manager: %w", err)
	}

	// 建立 TunnelManager
	tm := iostunnel.NewTunnelManager(pm, t.UseUserspaceNetworking)

	// 建立一個會在收到 SIGINT 或 SIGTERM 時自動取消的 context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // 確保在函數結束時停止監聽信號

	// 啟動定期更新 tunnel 的 goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := tm.UpdateTunnels(ctx)
				if err != nil {
					log.Printf("Warning: failed to update tunnels: %v", err)
				} else {
					// 印出所有活躍的 tunnels
					// tunnels, err := tm.ListTunnels()
					_, err := tm.ListTunnels()
					if err != nil {
						log.Printf("Warning: failed to list tunnels: %v", err)
						continue
					}
					// if len(tunnels) > 0 {
					// 	log.Printf("Active tunnels: %d", len(tunnels))
					// 	for _, t := range tunnels {
					// 		log.Printf("  UDID: %s, Address: %s, RSD Port: %d, Userspace Port: %d",
					// 			t.Udid, t.Address, t.RsdPort, t.UserspaceTUNPort)
					// 	}
					// }
				}
			}
		}
	}()

	// 使用正確的 tunnel.ServeTunnelInfo 函數啟動 HTTP API
	go func() {
		log.Printf("Starting tunnel info server on port %d", t.TunnelInfoPort)
		err := iostunnel.ServeTunnelInfo(tm, t.TunnelInfoPort)
		if err != nil {
			log.Fatalf("Failed to start tunnel server: %v", err)
		}
	}()

	return nil
}
