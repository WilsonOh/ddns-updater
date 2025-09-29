package main

import (
	"context"
	"ddns-go/internal/config"
	"ddns-go/internal/ddns"
	"ddns-go/internal/logger"
	"ddns-go/internal/metrics"
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
	// Initialize logging
	logger.Init()

	// Initialize metrics
	metricsConfig := metrics.LoadConfigFromEnv()
	metricsRecorder := metrics.NewRecorder(metricsConfig)

	// Set build information
	metrics.InitializeBuildInfo(metricsRecorder, metricsConfig.Version)

	// Ensure metrics are pushed at the end
	defer func() {
		if err := metricsRecorder.Push(); err != nil {
			slog.Error("Failed to push metrics", slog.Any("error", err))
		}
	}()

	// Track overall run status
	var runStatus string
	defer func() {
		metricsRecorder.RecordRun(runStatus)
	}()

	// Load configuration
	cfg, err := config.Get()
	if err != nil {
		runStatus = "failure"
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

	// Fetch public IP address with metrics
	go func() {
		defer wg.Done()
		start := time.Now()

		ipAddr, err = ddns.GetPublicIPAddress()
		duration := time.Since(start)

		if err != nil {
			metricsRecorder.RecordIPFetch(duration, false)
			runStatus = "failure"
			slog.Error("failed to get IP", slog.Any("error", err))
			os.Exit(1)
		}

		metricsRecorder.RecordIPFetch(duration, true)
		slog.Info("fetched current public IP address", slog.String("ip address", ipAddr), slog.String("time_taken", getTimeTaken(start)))
	}()

	// Fetch DNS records with metrics
	go func() {
		defer wg.Done()
		start := time.Now()

		records, err = ddns.FetchARecordsToUpdate(ctx, client, cfg)
		duration := time.Since(start)

		if err != nil {
			metricsRecorder.RecordCloudflareAPI("list_records", duration, err)
			runStatus = "failure"
			slog.Error("failed to fetch A records", slog.Any("error", err))
			os.Exit(1)
		}

		metricsRecorder.RecordCloudflareAPI("list_records", duration, nil)
		slog.Info("fetched records to update", slog.Any("records", records), slog.String("time_taken", getTimeTaken(start)))
	}()
	wg.Wait()

	// Process DNS records and track what we're doing
	params := ddns.AssembleBatchUpdateParams(cfg, records, ipAddr)

	// Track record processing
	recordsToUpdate := len(params)

	// Record individual record metrics
	for _, record := range records {
		domain := record.Name
		if record.Content == ipAddr {
			metricsRecorder.RecordDNSRecord("skipped", domain)
		} else {
			metricsRecorder.RecordDNSRecord("updated", domain)
		}
	}

	if len(params) == 0 {
		runStatus = "skipped"
		slog.Info("skipping batch update as no records need to be updated")
		return
	}

	// Perform DNS update with metrics
	start := time.Now()
	res, err := ddns.BatchUpdateDDNSRecords(ctx, cfg, client, params)
	duration := time.Since(start)

	if err != nil {
		metricsRecorder.RecordDNSUpdate(duration, false)
		metricsRecorder.RecordCloudflareAPI("batch_update", duration, err)
		runStatus = "failure"
		slog.Error("failed to make batch patch request", slog.Any("error", err))
		return
	}

	metricsRecorder.RecordDNSUpdate(duration, true)
	metricsRecorder.RecordCloudflareAPI("batch_update", duration, nil)

	// Check if we detected an actual IP change
	if recordsToUpdate > 0 {
		metricsRecorder.RecordIPChange("previous", ipAddr)
	}

	runStatus = "success"

	start = time.Now()
	formattedResp, _ := json.MarshalIndent(res, "", "  ")
	slog.Info("Batch update successful.", slog.Any("records", formattedResp), slog.String("time_taken", getTimeTaken(start)))
}
