package utils

import (
	"fmt"
	"log"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

func NewDeviceWithRsd(device ios.DeviceEntry, tun tunnel.Tunnel) (ios.DeviceEntry, error) {
	if tun.UserspaceTUNPort > 0 {
		device.UserspaceTUN = true
		device.UserspaceTUNHost = "localhost"
		device.UserspaceTUNPort = tun.UserspaceTUNPort
	}

	// 連接到 RSD 服務
	var rsdService ios.RsdService
	var err error

	udid := device.Properties.SerialNumber

	// 直接通過 IPv6 連接
	log.Printf("Connecting via RSD...")
	rsdService, err = ios.NewWithAddrPortDevice(tun.Address, tun.RsdPort, device)

	if err != nil {
		return device, fmt.Errorf("failed to connect to RSD: %w", err)
	}
	defer rsdService.Close()

	// 執行 RSD握手
	log.Printf("Performing RSD handshake...")
	rsdProvider, err := rsdService.Handshake()
	if err != nil {
		return device, fmt.Errorf("RSD handshake failed: %w", err)
	}

	// 使用 RSD provider 創建設備
	rsdDevice, err := ios.GetDeviceWithAddress(udid, tun.Address, rsdProvider)
	if err != nil {
		return device, fmt.Errorf("failed to create RSD device: %w", err)
	}

	// 保留 userspace TUN 配置
	if tun.UserspaceTUNPort > 0 {
		rsdDevice.UserspaceTUN = true
		rsdDevice.UserspaceTUNHost = "localhost"
		rsdDevice.UserspaceTUNPort = tun.UserspaceTUNPort
	}

	log.Printf("✓ RSD connection established")
	return rsdDevice, nil
}
