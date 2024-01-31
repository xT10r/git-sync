/*
Пакет metrics отвечает за сбор числовых метрик работы приложения и экспорт их в формате Prometheus.
Другие пакеты вызывают RecordSync() для учета метрик. Функция RegisterMetrics() регистрирует коллекторы в реестре Prometheus.
*/

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	syncCount = prometheus.NewCounter(prometheus.CounterOpts{}) // счетчик синхронизаций

	syncDuration = prometheus.NewHistogram(prometheus.HistogramOpts{}) // гистограмма времени синхронизации
)

// Увеличиваем счетчик при синхронизации
func RecordSync(duration time.Duration) {

	syncCount.Inc()
	syncDuration.Observe(duration.Seconds())

}

// Регистрируем метрики в реестре Prometheus
func RegisterMetrics() {

	prometheus.MustRegister(syncCount)
	prometheus.MustRegister(syncDuration)

}
