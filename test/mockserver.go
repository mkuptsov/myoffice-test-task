package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mkuptsov/myoffice-test-task/internal/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	requestTimeout = 5
)

func runServer(t *testing.T) {
	srv := http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/resp200", resp200)
	http.HandleFunc("/resp404", resp404)
	http.HandleFunc("/respReqTimeoutExpired", respReqTimeoutExpired)

	go func() {
		err := srv.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	tests(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatalf("server shutdown: %v", err)
		srv.Close()
	}
}

func resp200(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "200 OK")
}

func resp404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func respReqTimeoutExpired(w http.ResponseWriter, r *http.Request) {
	time.Sleep(6 * time.Second)
	_, _ = io.WriteString(w, "200 OK")
}

func tests(t *testing.T) {
	type test struct {
		input    string
		expected string
	}

	cases := []test{
		{
			input:    "http://localhost:8080/resp200",
			expected: fmt.Sprintf("url: %s size: %d bytes", "http://localhost:8080/resp200", 6),
		},
		{
			input:    "http://localhost:8080/resp404",
			expected: fmt.Sprintf("client: url %s unavilble: status code: %d", "http://localhost:8080/resp404", 404),
		},
		{
			input:    "http://localhost:8080/respReqTimeoutExpired",
			expected: "",
		},
	}

	client := &http.Client{
		Timeout: time.Duration(requestTimeout) * time.Second,
	}

	t.Run("response 200", func(t *testing.T) {
		actual, err := process.Process(client, cases[0].input)
		require.Empty(t, err)
		require.Contains(t, actual, cases[0].expected)
	})

	t.Run("response 404", func(t *testing.T) {
		_, err := process.Process(client, cases[1].input)
		require.Contains(t, err.Error(), cases[1].expected)
	})

	t.Run("request timeout expired", func(t *testing.T) {
		_, err := process.Process(client, cases[2].input)
		c := assert.Comparison(func() bool {
			return os.IsTimeout(errors.Unwrap(err))
		})

		require.Condition(t, c)
	})
}
