package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	k := flag.Int("k", 5, "максимальное число одновременных запросов")
	q := flag.String("q", "Go", "искомая строка")
	flag.Parse()

	if *k <= 0 {
		fmt.Fprintln(os.Stderr, "ошибка: флаг -k должен быть больше 0")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	sem := make(chan struct{}, *k)
	var wg sync.WaitGroup
	var printMu sync.Mutex
	var total atomic.Int64

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rawURL := strings.TrimSpace(scanner.Text())
		if rawURL == "" {
			continue
		}

		// Занимаем слот в семафоре до запуска горутины,
		// чтобы не плодить тысячи ожидающих горутин при большом вводе
		sem <- struct{}{}
		wg.Add(1)

		go func(targetURL string) {
			defer wg.Done()
			defer func() { <-sem }() // Освобождаем слот после завершения

			count, err := fetchAndCount(client, targetURL, *q)
			if err != nil {
				printMu.Lock()
				fmt.Fprintf(os.Stderr, "Error for %s: %v\n", targetURL, err)
				printMu.Unlock()
				return
			}

			total.Add(int64(count))

			printMu.Lock()
			fmt.Printf("Count for %s: %d\n", targetURL, count)
			printMu.Unlock()
		}(rawURL)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
	}

	wg.Wait()
	fmt.Printf("Total: %d\n", total.Load())
}

func fetchAndCount(client *http.Client, targetURL, query string) (int, error) {
	resp, err := client.Get(targetURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return strings.Count(string(body), query), nil
}
