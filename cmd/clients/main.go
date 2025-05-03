package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Адрес прокси-сервера
	proxyAddr := "http://localhost:8080" // ваш прокси должен слушать этот адрес

	// Количество клиентов и запросов
	numClients := 10
	requestsPerClient := 50

	// Создаем WaitGroup для ожидания завершения всех клиентов
	var wg sync.WaitGroup
	wg.Add(numClients)

	// Запускаем клиентов
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			// Создаем HTTP-клиента
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			// Отправляем несколько запросов
			for j := 0; j < requestsPerClient; j++ {
				start := time.Now()

				// Создаем URL для запроса (можете изменить на нужный вам путь)
				reqURL := fmt.Sprintf("%s/api/data?client=%d&req=%d", proxyAddr, clientID, j)

				// Создаем запрос
				req, err := http.NewRequest("GET", reqURL, nil)
				if err != nil {
					log.Printf("Client %d: error creating request %d: %v", clientID, j, err)
					continue
				}

				// Отправляем запрос
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Client %d: request %d failed: %v", clientID, j, err)
					continue
				}
				defer resp.Body.Close()

				// Логируем результат
				duration := time.Since(start)
				log.Printf("Client %d: request %d completed in %v - Status: %d",
					clientID, j, duration, resp.StatusCode)

				// Небольшая пауза между запросами
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Ждем завершения всех клиентов
	wg.Wait()
	log.Println("All clients completed")
}
