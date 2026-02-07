package simulatelocation

import (
	"iGeoGo-cli/tunnel"
	"iGeoGo-cli/utils"
	"log"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

type SimlulateLocationService struct {
	service *instruments.LocationSimulationService
}

func NewSimulateLocationService(device ios.DeviceEntry) *SimlulateLocationService {
	// create rsd device connection
	tun, err := tunnel.GetTunnelForDevice(device.Properties.SerialNumber)
	if err != nil {
		log.Fatalf("Failed to get tunnel for device %s: %v", device.Properties.SerialNumber, err)
	}
	rsdDevice, err := utils.NewDeviceWithRsd(device, tun)
	if err != nil {
		log.Fatalf("Failed to create RSD device for %s: %v", device.Properties.SerialNumber, err)
	}

	// create location simulation service
	service, err := instruments.NewLocationSimulationService(rsdDevice)
	if err != nil {
		log.Fatalf("Failed to create location service: %v\nMake sure Developer Mode is enabled on device", err)
	}
	return &SimlulateLocationService{service: service}
}

func (s *SimlulateLocationService) Set(coordinates utils.Coordinate) (utils.Coordinate, error) {
	err := s.service.StartSimulateLocation(coordinates.Lat, coordinates.Lng)
	if err != nil {
		return utils.Coordinate{}, err
	}
	return coordinates, nil
}

func (s *SimlulateLocationService) Destory() {
	s.service.StopSimulateLocation()
}
