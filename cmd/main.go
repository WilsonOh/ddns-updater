package main

import (
	"context"
	"ddns-go/internal/config"
	"ddns-go/internal/ddns"
	"ddns-go/internal/logger"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func handleSigs(cancel func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	cancel()
}

func getTimeTaken(start time.Time) string {
	timeTaken := time.Since(start)
	return fmt.Sprintf("%fs", timeTaken.Seconds())
}

func main() {
	logger.InitLogger()

	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("failed to retrieve config: %v", err)
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(cfg.APIToken),
		option.WithAPIEmail(cfg.Email),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go handleSigs(cancel)

	var records []dns.RecordResponse
	var ipAddr string

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		start := time.Now()
		ipAddr, err = ddns.GetPublicIPAddress()
		if err != nil {
			slog.Error("failed to get IP", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("fetched current public IP address", slog.String("ip address", ipAddr), slog.String("time_taken", getTimeTaken(start)))
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		records, err = ddns.FetchARecordsToUpdate(ctx, client, cfg)
		if err != nil {
			slog.Error("failed to fetch A records", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("fetched records to update", slog.Any("records", records), slog.String("time_taken", getTimeTaken(start)))
	}()
	wg.Wait()

	params := ddns.AssembleBatchUpdateParams(cfg, records, ipAddr)
	if len(params) == 0 {
		slog.Info("skipping batch update as no records need to be updated")
		os.Exit(0)
	}
	res, err := ddns.BatchUpdateDDNSRecords(ctx, cfg, client, params)
	if err != nil {
		slog.Error("failed to make batch patch request", slog.Any("error", err))
	}

	start := time.Now()
	formattedResp, _ := json.MarshalIndent(res, "", "  ")
	slog.Info("Batch update successful.", slog.Any("records", formattedResp), slog.String("time_taken", getTimeTaken(start)))
}
