package converter_test

import (
	"math/big"
	"testing"

	"github.com/somebadcode/exchange/converter"
)

func TestParseRat(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Rat
		wantErr bool
	}{
		{
			args: args{
				s: "12.34",
			},
			want: big.NewRat(1234, 100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ParseRat(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Cmp(tt.want) != 0 {
				t.Errorf("ParseRat() got = %v, want %v", got.FloatString(2), tt.want.FloatString(2))
			}
		})
	}
}
