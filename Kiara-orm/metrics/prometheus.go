package metrics

import (
	"fmt"
	"sync"
	
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusExporter exporta métricas para o Prometheus
type PrometheusExporter struct {
	mu       sync.RWMutex
	counters map[string]*prometheus.CounterVec
	gauges   map[string]*prometheus.GaugeVec
	
	registry *prometheus.Registry
}

// NewPrometheusExporter cria um novo exportador Prometheus
func NewPrometheusExporter() *PrometheusExporter {
	exporter := &PrometheusExporter{
		counters:  make(map[string]*prometheus.CounterVec),
		gauges:    make(map[string]*prometheus.GaugeVec),
		registry:  prometheus.NewRegistry(),
	}
	
	// Registra métricas padrão
	exporter.registerDefaultMetrics()
	
	return exporter
}

// Export exporta uma métrica para o Prometheus
func (p *PrometheusExporter) Export(metric Metric) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	switch metric.Type {
	case QueryExecution:
		counter, ok := p.counters["query_total"]
		if ok {
			counter.With(metric.Labels).Add(metric.Value)
		}
		
	case CacheHit, CacheMiss:
		counter, ok := p.counters["cache_operations"]
		if ok {
			counter.With(metric.Labels).Add(metric.Value)
		}
		
	case ConnectionUsage:
		gauge, ok := p.gauges["connections_active"]
		if ok {
			gauge.With(metric.Labels).Set(metric.Value)
		}
		
	case ErrorCount:
		counter, ok := p.counters["errors_total"]
		if ok {
			counter.With(metric.Labels).Add(metric.Value)
		}
	}
	
	return nil
}

// registerDefaultMetrics registra as métricas padrão
func (p *PrometheusExporter) registerDefaultMetrics() {
	// Contador de queries
	queryCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orm_query_total",
			Help: "Total number of queries executed",
		},
		[]string{"type", "table"},
	)
	p.registry.MustRegister(queryCounter)
	p.counters["query_total"] = queryCounter
	
	// Contador de operações de cache
	cacheCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orm_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)
	p.registry.MustRegister(cacheCounter)
	p.counters["cache_operations"] = cacheCounter
	
	// Gauge de conexões ativas
	connectionsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "orm_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"pool"},
	)
	p.registry.MustRegister(connectionsGauge)
	p.gauges["connections_active"] = connectionsGauge
	
	// Contador de erros
	errorCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orm_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "operation"},
	)
	p.registry.MustRegister(errorCounter)
	p.counters["errors_total"] = errorCounter
}

// GetRegistry retorna o registro Prometheus
func (p *PrometheusExporter) GetRegistry() *prometheus.Registry {
	return p.registry
} 