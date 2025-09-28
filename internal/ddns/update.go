package ddns

import (
	"context"
	"ddns-go/internal/config"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
)

func AssembleBatchUpdateParams(cfg *config.Config, records []dns.RecordResponse, ipAddr string) []dns.BatchPatchUnionParam {
	ret := make([]dns.BatchPatchUnionParam, 0)

	for _, record := range records {
		param := &dns.BatchPatchARecordParam{
			ID: cloudflare.String(record.ID),
			ARecordParam: dns.ARecordParam{
				Content: cloudflare.String(ipAddr),
			},
		}
		ret = append(ret, param)
	}

	return ret
}

func BatchUpdateDDNSRecords(ctx context.Context, cfg *config.Config, client *cloudflare.Client, params []dns.BatchPatchUnionParam) ([]dns.RecordResponse, error) {
	res, err := client.DNS.Records.Batch(ctx, dns.RecordBatchParams{
		ZoneID:  cloudflare.String(cfg.ZoneID),
		Patches: cloudflare.F(params),
	})
	if err != nil {
		return nil, err
	}

	return res.Patches, nil
}
