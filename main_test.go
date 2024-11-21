package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

func TestRunBenchmark(t *testing.T) {
	testCases := []struct {
		name        string
		sqlQuery    string
		durationMs  int
		numWorkers  int
		expectError bool
		setupMock   func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "Valid SELECT query",
			sqlQuery:   "SELECT 1",
			durationMs: 1000,
			numWorkers: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:        "Invalid SQL query",
			sqlQuery:    "INVALID QUERY",
			durationMs:  1000,
			numWorkers:  10,
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INVALID QUERY").WillReturnError(fmt.Errorf("syntax error"))
			},
		},
		{
			name:       "Zero workers",
			sqlQuery:   "SELECT 1",
			durationMs: 1000,
			numWorkers: 0,
			setupMock: func(mock sqlmock.Sqlmock) {
				// No expectations since no workers will run
			},
		},
		{
			name:        "Negative duration",
			sqlQuery:    "SELECT 1",
			durationMs:  -1000,
			numWorkers:  10,
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				// No expectations since benchmark won't run
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Ошибка при создании mock базы данных: %v", err)
			}
			defer db.Close()

			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			totalQueries, elapsed, rps := RunBenchmark(tc.sqlQuery, db, tc.durationMs, tc.numWorkers)

			if tc.expectError {
				// Ожидается, что запросы не будут выполнены
				if totalQueries != 0 {
					t.Errorf("Ожидалось 0 выполненных запросов, получили %d", totalQueries)
				}
			} else {
				// Проверяем, что было выполнено хотя бы несколько запросов
				if totalQueries == 0 && tc.numWorkers > 0 && tc.durationMs > 0 {
					t.Error("Ожидалось, что будет выполнено хотя бы один запрос")
				}

				// Проверяем, что время выполнения не превышает заданное
				expectedDuration := time.Duration(tc.durationMs) * time.Millisecond
				if elapsed > expectedDuration+time.Millisecond*100 {
					t.Errorf("Время выполнения превышает ожидаемое: %v", elapsed)
				}

				// Проверяем, что RPS положительное число
				if rps < 0 {
					t.Errorf("Ожидалось положительное значение RPS, получили: %.2f", rps)
				}
			}

			// Проверяем, что все ожидания были выполнены
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Не все ожидания были выполнены: %v", err)
			}
		})
	}
}
