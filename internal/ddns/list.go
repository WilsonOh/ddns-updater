package ddns

import (
	"context"
	"ddns-go/internal/config"
	"strings"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
)

func FetchARecordsToUpdate(ctx context.Context, client *cloudflare.Client, cfg *config.Config) ([]dns.RecordResponse, error) {
	resp, err := client.DNS.Records.List(ctx,
		dns.RecordListParams{
			ZoneID: cloudflare.String(cfg.ZoneID),
			Comment: cloudflare.F(dns.RecordListParamsComment{
				Contains: cloudflare.String("self host"),
			}),
			Type: cloudflare.F(dns.RecordListParamsTypeA),
		},
	)

	if err != nil {
		return nil, err
	}

	var filteredRecords []dns.RecordResponse
	for domain, domainCfg := range cfg.Domains {
		for _, subdomain := range domainCfg.SubDomains {
			fullDomain := subdomain + "." + domain
			if subdomain == "@" {
				fullDomain = domain
			}
			for _, record := range resp.Result {
				if strings.EqualFold(fullDomain, record.Name) {
					filteredRecords = append(filteredRecords, record)
				}
			}
		}
	}

	return filteredRecords, nil
}
