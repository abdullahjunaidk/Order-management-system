package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
	"math/rand/v2"
)

// ConsulRegistry struct.
// This struct is used to represent a Consul registry.
//
// Attributes:
//   - client (*api.Client): The Consul client.
type ConsulRegistry struct {
	client *api.Client
}

// NewConsulRegistry function.
// This function is used to create a new Consul registry.
//
// Parameters:
//   - address (string): The Consul address.
//
// Returns:
//   - *ConsulRegistry: The Consul registry.
//   - error: The error.
func NewConsulRegistry(address string) (*ConsulRegistry, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulRegistry{client: client}, nil
}

// RegisterService function.
// This function is used to register a service in Consul.
//
// Parameters:
//   - serviceName (string): The service name.
//   - instanceID (string): The instance ID.
//   - host (string): The host.
//   - port (int): The port.
//   - isGRPC (bool): The flag to indicate if the service is a gRPC service.
//   - healthCheckURL (string): The health check URL.
//
// Returns:
//   - error: The error.
func (r *ConsulRegistry) RegisterService(serviceName, instanceID, host string, port int, isGRPC bool, healthCheckURL, healthCheckInterval string) error {
	address := host
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP: %w", err)
	}
	address = localIP

	registration := &api.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
	}

	if !isGRPC {
		registration.Check = &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", address, port, healthCheckURL),
			Interval:                       healthCheckInterval,
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		}
	} else {
		registration.Check = &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", address, port),
			Interval:                       healthCheckInterval,
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		}
	}

	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// DeregisterService function.
// This function is used to deregister a service from Consul.
//
// Parameters:
//   - instanceID (string): The instance ID.
//
// Returns:
//   - error: The error.
func (r *ConsulRegistry) DeregisterService(instanceID string) error {
	if err := r.client.Agent().ServiceDeregister(instanceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	return nil
}

// GetServiceAddress function.
// This function is used to get the address of a service from Consul.
//
// Parameters:
//   - serviceName (string): The service name.
//
// Returns:
//   - string: The service address.
//   - error: The error.
func (r *ConsulRegistry) GetServiceAddress(serviceName string) (string, error) {
	service, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve service '%s' from Consul: %w", serviceName, err)
	}

	if len(service) == 0 {
		return "", fmt.Errorf("no healthy instances of service '%s' found in Consul", serviceName)
	}

	numInstances := len(service)
	instance := service[rand.IntN(numInstances)] //instance here picking randomly from the list of instances
	address := instance.Service.Address
	port := instance.Service.Port

	return net.JoinHostPort(address, strconv.Itoa(port)), nil
}

// getLocalIP function.
// This function is used to get the local IP address.
//
// Returns:
//   - string: The local IP address.
//   - error: The error.
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no non-loopback IPv4 address found")
}
