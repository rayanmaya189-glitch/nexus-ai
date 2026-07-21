package grpc

import (
	"fmt"
	"sync"
	"time"
)

type ServiceInstance struct {
	ID        string
	Name      string
	Address   string
	Port      int
	GRPCPort  int
	Status    string
	Metadata  map[string]string
	LastSeen  time.Time
}

func (s *ServiceInstance) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", s.Address, s.GRPCPort)
}

type ServiceRegistry struct {
	services map[string][]*ServiceInstance
	mu       sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
	}
}

func (r *ServiceRegistry) Register(instance *ServiceInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance.LastSeen = time.Now()
	instance.Status = "healthy"

	r.services[instance.Name] = append(r.services[instance.Name], instance)
}

func (r *ServiceRegistry) Deregister(name, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[name]
	for i, inst := range instances {
		if inst.ID == id {
			r.services[name] = append(instances[:i], instances[i+1:]...)
			return
		}
	}
}

func (r *ServiceRegistry) GetService(name string) (*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := r.services[name]
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", name)
	}

	// Simple round-robin
	for _, inst := range instances {
		if inst.Status == "healthy" {
			return inst, nil
		}
	}

	return nil, fmt.Errorf("no healthy instances for service: %s", name)
}

func (r *ServiceRegistry) GetServiceAddress(name string) (string, error) {
	inst, err := r.GetService(name)
	if err != nil {
		return "", err
	}
	return inst.GRPCAddress(), nil
}

func (r *ServiceRegistry) ListServices() map[string][]*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*ServiceInstance)
	for name, instances := range r.services {
		result[name] = make([]*ServiceInstance, len(instances))
		copy(result[name], instances)
	}
	return result
}

func (r *ServiceRegistry) UpdateHealth(name, id string, healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[name]
	for _, inst := range instances {
		if inst.ID == id {
			if healthy {
				inst.Status = "healthy"
			} else {
				inst.Status = "unhealthy"
			}
			inst.LastSeen = time.Now()
			return
		}
	}
}

// CleanupStale removes instances that haven't been seen in a while
func (r *ServiceRegistry) CleanupStale(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for name, instances := range r.services {
		var healthy []*ServiceInstance
		for _, inst := range instances {
			if now.Sub(inst.LastSeen) < maxAge {
				healthy = append(healthy, inst)
			}
		}
		r.services[name] = healthy
	}
}
