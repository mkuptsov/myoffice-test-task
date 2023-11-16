package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/mkuptsov/myoffice-test-task/internal/process"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "url-processor [file with urls]",
	Short: "url-processor sends url requests and processes responses",
	Long: `url-processor reads urls from file, validates them, sends requests, 
	outputs the size of response content and processing time`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		r := bufio.NewScanner(file)
		inputCh := make(chan string, 1)
		semaphore := make(chan struct{}, simultaneousReqNum)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				close(inputCh)
				close(semaphore)
			}()

			for r.Scan() {
				rawURL := r.Text()
				_, err := url.ParseRequestURI(rawURL)

				if err != nil {
					fmt.Println(err)
				} else {
					semaphore <- struct{}{}
					inputCh <- rawURL
				}
			}
		}()

		client := &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		}

		for url := range inputCh {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				defer func() {
					<-semaphore
				}()

				res, err := process.Process(client, url)
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Println(res)

			}(url)
		}

		wg.Wait()

		fmt.Println("\nAll urls processed")
	},
}

var simultaneousReqNum int
var requestTimeout int // in seconds

func init() {
	rootCmd.PersistentFlags().IntVarP(&simultaneousReqNum, "req-num", "n", 50, "Number of simultaneous requests that app sends")
	rootCmd.PersistentFlags().IntVarP(&requestTimeout, "req-timeout", "t", 5, "How long response will be waited")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
