package configurd

import (
	"errors"
	"reflect"
	"testing"
)

func TestConfigurd_Service(t *testing.T) {
	type fields struct {
		Services []Service
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Service
		wantErr error
	}{
		{
			name:   "empty",
			fields: fields{[]Service{}},
			args:   args{name: "empty"},
			want:   Service{}, wantErr: ErrResourceNotFound("service", "empty"),
		},
		{
			name:   "found",
			fields: fields{[]Service{{Name: "found"}}},
			args:   args{name: "found"},
			want:   Service{Name: "found"}, wantErr: nil,
		},
		{
			name:   "not found",
			fields: fields{[]Service{{Name: "found"}}},
			args:   args{name: "not found"},
			want:   Service{}, wantErr: ErrResourceNotFound("service", "not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Configurd{
				Services: tt.fields.Services,
			}
			got, err := c.Service(tt.args.name)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Service() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service() got = %v, want %v", got, tt.want)
			}
		})
	}
}
