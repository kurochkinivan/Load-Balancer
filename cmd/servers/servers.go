package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var serversCount, startPort int
	flag.IntVar(&serversCount, "servers", 5, "specify running servers count")
	flag.IntVar(&startPort, "start_port", 8090, "specify the first port for the servers")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, nil)
	log := slog.New(handler)

	for i := 0; i < serversCount; i++ {
		go startTestServer(log, strconv.Itoa(startPort+i))
	}

	log.Info("all test servers are initialized and started!",
		slog.String(
			"ports",
			fmt.Sprintf("%d-%d", startPort, startPort+serversCount-1),
		),
	)

	select {}
}

func startTestServer(log *slog.Logger, port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("incoming request",
			slog.String("url", r.URL.Path),
			slog.String("host_header", r.Host),
			slog.String("port", port),
		)
	})

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Error("failed to start server", slog.String("port", port), slog.String("error", err.Error()))
	}
}
