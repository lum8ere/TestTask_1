package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	var (
		sqlQuery   string
		dbDSN      string
		durationMs int
		numWorkers int
	)

	flag.StringVar(&sqlQuery, "query", "", "Оцениваемый SQL запрос")
	flag.StringVar(&dbDSN, "dsn", "", "DSN PostgreSQL")
	flag.IntVar(&durationMs, "duration", 1000, "Время проведения измерения RPS в миллисекундах")
	flag.IntVar(&numWorkers, "workers", 100, "Количество конкурентных воркеров")
	flag.Parse()

	if sqlQuery == "" || dbDSN == "" {
		fmt.Println("Необходимо указать SQL запрос и DSN базы данных.")
		flag.Usage()
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		fmt.Printf("Ошибка подключения к базе данных: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	totalQueries, elapsed, rps := RunBenchmark(sqlQuery, db, durationMs, numWorkers)

	fmt.Printf("Всего выполнено запросов: %d\n", totalQueries)
	fmt.Printf("Общее время: %v\n", elapsed)
	fmt.Printf("RPS (запросов в секунду): %.2f\n", rps)
}

func RunBenchmark(sqlQuery string, db *sql.DB, durationMs int, numWorkers int) (int64, time.Duration, float64) {
	duration := time.Duration(durationMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var queryCount int64
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					err := executeQuery(ctx, db, sqlQuery)
					if err != nil {
						fmt.Printf("Ошибка выполнения запроса: %v\n", err)
						continue
					}
					atomic.AddInt64(&queryCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	totalQueries := atomic.LoadInt64(&queryCount)
	rps := float64(totalQueries) / elapsed.Seconds()

	return totalQueries, elapsed, rps
}

func executeQuery(ctx context.Context, db *sql.DB, query string) error {
	_, err := db.ExecContext(ctx, query)
	return err
}
