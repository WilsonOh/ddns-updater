package main

import (
	"context"
	"ddns-go/internal/config"
	"ddns-go/internal/ddns"
	"encoding/json"
	"fmt"
	"log"
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

func reportTime(name string, start time.Time) {
	timeTaken := time.Since(start)
	fmt.Printf("%s took %fs to complete\n", name, timeTaken.Seconds())
}

func main() {
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
		defer reportTime("get public ip address", time.Now())
		ipAddr, err = ddns.GetPublicIPAddress()
		if err != nil {
			log.Fatalf("failed to get IP: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer reportTime("fetch A records", time.Now())
		records, err = ddns.FetchARecordsToUpdate(ctx, client, cfg)
		if err != nil {
			log.Fatalf("failed to fetch A records: %v", err)
		}
	}()
	wg.Wait()

	params := ddns.AssembleBatchUpdateParams(cfg, records, ipAddr)
	res, err := ddns.BatchUpdateDDNSRecords(ctx, cfg, client, params)
	if err != nil {
		log.Fatalf("failed to make batch patch request: %v", err)
	}

	defer reportTime("batch update records", time.Now())
	fmt.Printf("Batch update successful.")
	fmt.Printf("Updated records: \n")
	formatted, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("%s\n", string(formatted))
}
