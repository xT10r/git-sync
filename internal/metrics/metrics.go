package metrics

/*
package pg

import (
	"context"
	"fmt"
	"time"
)

func (pg *PGData) Add(ctx context.Context, item *models.MetricItem) (models.MetricItem, error) {

	if pg.pool == nil {
		return *item, fmt.Errorf("pg.pool is nil")
	}

	// Открываем соединение с базой данных
	conn, err := pg.pool.Conn(ctx)
	if err != nil {
		fmt.Printf("ошибка при открытии соединения с базой данных: %s\n", err.Error())
		return *item, err
	}
	defer conn.Close()

	// Выполняем SQL-запрос для добавления данных в таблицу metrics
	query := `INSERT INTO metrics (name, value, timestamp)
			  VALUES ($1, $2, $3)`
	_, err = conn.ExecContext(ctx, query, item.Name, item.Value, item.Timestamp)
	if err != nil {
		fmt.Printf("ошибка при добавлении метрики '%s' в таблицу metrics: %s\n", item.Name, err.Error())
		return *item, err
	}
	return *item, nil
}

func (pg *PGData) ListMetrics(ctx context.Context, name string, from, to time.Time) ([]models.MetricItem, error) {

	// Запрос для получения записей таблицы metrics
	query := "SELECT id, name, value, timestamp FROM metrics WHERE name = $1 AND timestamp >= $2 AND timestamp <= $3"

	// Проверим правильность указания дат (дата 'to' - должна быть больше даты 'from')
	if !from.Before(to) {
		return nil, fmt.Errorf("дата 'to' - должна быть больше даты 'from'")
	}

	from_formated := from.Format("2006-01-02 15:04:05.999999")
	to_formated := to.Format("2006-01-02 15:04:05.999999")

	// Выполняем SQL-запрос для получения метрик из таблицы metrics
	rows, err := pg.pool.QueryContext(ctx, query, name, from_formated, to_formated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.MetricItem
	for rows.Next() {
		var metric models.MetricItem
		err := rows.Scan(&metric.Id, &metric.Name, &metric.Value, &metric.Timestamp)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

// TODO: задание 8+. Доделать после сдачи заданий 6+
// func (pg *PGData) ListChecks(ctx context.Context, id string) ([]models.MetricCheck, error) {
// 	return nil
// }

/*

6.2.1

1.2. Работу с отдельной сущностью базы данных принято располагать в отдельном файле этого же пакета,
для метрик подошло бы имя ./Service/internal/checker/models/repository/pg/metrics.go,
создайте его и разместите в нем метод Add вашей структуры.

2.) В файле ./Service/internal/checker/models/repository/pg/metrics.go добавляем метод List,
что бы соответствовать интерфейсу models.RepositoryReadInterface

6.4.1

2.) В файле ./Service/internal/checker/models/repository/pg/metrics.go
добавляем метод List, что бы соответствовать интерфейсу models.RepositoryReadInterface

*/
