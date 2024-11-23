package converter_test

import (
	"context"
	_ "embed"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/somebadcode/exchange/converter"
)

//go:embed testdata/base_usd_payload_1732374021.json
var baseUSDPayload1732374021 []byte

func StartServer(t *testing.T) (*http.Client, *url.URL, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(baseUSDPayload1732374021)
	}))

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	return server.Client(), serverURL, server.Close
}

func TestConverter_ConvertTo(t *testing.T) {
	type fields struct {
		TimeToLive   time.Duration
		Endpoint     *url.URL
		BaseCurrency string
	}
	type args struct {
		ctx      context.Context
		currency string
		qty      *big.Rat
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Rat
		wantErr bool
	}{
		{
			fields: fields{
				BaseCurrency: "USD",
			},
			args: args{
				ctx:      context.Background(),
				currency: "EUR",
				qty:      big.NewRat(1, 1),
			},
			want: big.NewRat(959877, 1000000),
		},
	}

	client, serverURL, serverClose := StartServer(t)
	defer serverClose()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &converter.Converter{
				Client:         client,
				TimeToLive:     tt.fields.TimeToLive,
				Endpoint:       serverURL,
				BaseCurrency:   tt.fields.BaseCurrency,
				CacheDirectory: t.TempDir(),
			}

			got, err := c.ConvertTo(tt.args.ctx, tt.args.currency, tt.args.qty)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Cmp(tt.want) != 0 {
				t.Errorf("ConvertTo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
