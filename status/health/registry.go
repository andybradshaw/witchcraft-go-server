package health

import (
	"context"
	"sync"

	healthapi "github.com/palantir/witchcraft-go-server/conjure/witchcraft/api/health"
	"github.com/palantir/witchcraft-go-server/status"
)

// CheckRegistry provides callers with the ability to dynamically add and remove health checks from the underlying
// registry. The registry is thread safe.
type CheckRegistry interface {
	status.HealthCheckSource

	Register(check TypedCheck)
	RegisterNamedCheck(checkType healthapi.CheckType, check Check)
	Unregister(check healthapi.CheckType)
}

type Check interface {
	HealthCheckResult(ctx context.Context) healthapi.HealthCheckResult
}

type TypedCheck interface {
	Check
	HealthCheckType() healthapi.CheckType
}

var _ CheckRegistry = (*DefaultHealthCheckRegistry)(nil)

type DefaultHealthCheckRegistry struct {
	m        *sync.Mutex
	registry map[healthapi.CheckType]Check
}

// NewCheckRegistry creates a new health check registry; an implementation of status.HealthCheckSource
// which can dynamically add and remove TypedCheck's from it's internal store.
func NewCheckRegistry() *DefaultHealthCheckRegistry {
	return &DefaultHealthCheckRegistry{
		m:        &sync.Mutex{},
		registry: make(map[healthapi.CheckType]Check),
	}
}

func (d *DefaultHealthCheckRegistry) HealthStatus(ctx context.Context) healthapi.HealthStatus {
	d.m.Lock()
	defer d.m.Unlock()

	result := healthapi.HealthStatus{
		Checks: map[healthapi.CheckType]healthapi.HealthCheckResult{},
	}
	for checkType, check := range d.registry {
		result.Checks[checkType] = check.HealthCheckResult(ctx)
	}
	return result
}

func (d *DefaultHealthCheckRegistry) Register(check TypedCheck) {
	d.m.Lock()
	defer d.m.Unlock()

	d.registry[check.HealthCheckType()] = check
}

func (d *DefaultHealthCheckRegistry) RegisterNamedCheck(checkType healthapi.CheckType, check Check) {
	d.m.Lock()
	defer d.m.Unlock()

	d.registry[checkType] = check
}

func (d *DefaultHealthCheckRegistry) Unregister(check healthapi.CheckType) {
	d.m.Lock()
	defer d.m.Unlock()

	delete(d.registry, check)
}
