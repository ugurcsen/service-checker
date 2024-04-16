package opensearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/opensearch-project/opensearch-go/v3"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
	"github.com/ugurcsen/service-checker/types"
	"net"
	"net/http"
	"time"
)

func SendToOpenSearch(conf types.ServerCheckConfig, result types.ResultStruct) error {
	ctx := context.Background()
	oc, err := newClient(conf)
	if err != nil {
		return err
	}

	bfr := bytes.NewBuffer(nil)
	req := opensearchapi.IndexReq{
		Index: conf.OpenSearch.Index,
		Body:  bfr,
	}

	err = json.NewEncoder(bfr).Encode(result)
	if err != nil {
		return err
	}

	_, err = oc.Index(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func newClient(conf types.ServerCheckConfig) (*opensearchapi.Client, error) {
	oc, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: conf.OpenSearch.Hosts,
			Username:  conf.OpenSearch.Username,
			Password:  conf.OpenSearch.Password,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second,
				DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS11,
				},
			},
		},
	})
	return oc, err
}
