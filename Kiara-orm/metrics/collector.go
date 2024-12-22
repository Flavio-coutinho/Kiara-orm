package metrics

import (
	"sync"
	"time"
)

// MetricType representa os tipos de métricas disponíveis
type MetricType string

const (
	QueryExecution  MetricType = "query_execution"
	CacheHit        MetricType = "cache_hit"
	CacheMiss       MetricType = "cache_miss"
	ConnectionUsage MetricType = "connection_usage"
	ErrorCount      MetricType = "error_count"
)

// Metric representa uma métrica coletada
type Metric struct {
	Type      MetricType
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// Collector coleta e armazena métricas do ORM
type Collector struct {
	mu      sync.RWMutex
	metrics []Metric
	
	// Callbacks para exportação de métricas
	exporters []MetricExporter
}

// MetricExporter define a interface para exportação de métricas
type MetricExporter interface {
	Export(metric Metric) error
}

// NewCollector cria uma nova instância do Collector
func NewCollector() *Collector {
	collector := &Collector{
		metrics:   make([]Metric, 0),
		exporters: make([]MetricExporter, 0),
	}
	
	// Inicia a rotina de limpeza de métricas antigas
	go collector.cleanup()
	
	return collector
}

// AddMetric adiciona uma nova métrica
func (c *Collector) AddMetric(metricType MetricType, value float64, labels map[string]string) {
	metric := Metric{
		Type:      metricType,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}
	
	c.mu.Lock()
	c.metrics = append(c.metrics, metric)
	c.mu.Unlock()
	
	// Exporta a métrica para todos os exporters
	for _, exporter := range c.exporters {
		go exporter.Export(metric)
	}
}

// GetMetrics retorna todas as métricas coletadas
func (c *Collector) GetMetrics() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	metrics := make([]Metric, len(c.metrics))
	copy(metrics, c.metrics)
	return metrics
}

// AddExporter adiciona um novo exportador de métricas
func (c *Collector) AddExporter(exporter MetricExporter) {
	c.mu.Lock()
	c.exporters = append(c.exporters, exporter)
	c.mu.Unlock()
}

// cleanup remove métricas antigas periodicamente
func (c *Collector) cleanup() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		c.mu.Lock()
		
		// Remove métricas mais antigas que 24 horas
		threshold := time.Now().Add(-24 * time.Hour)
		newMetrics := make([]Metric, 0)
		
		for _, metric := range c.metrics {
			if metric.Timestamp.After(threshold) {
				newMetrics = append(newMetrics, metric)
			}
		}
		
		c.metrics = newMetrics
		c.mu.Unlock()
	}
} 