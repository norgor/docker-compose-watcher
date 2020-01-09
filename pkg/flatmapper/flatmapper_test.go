package flatmapper

import (
	"reflect"
	"testing"
)

func TestMapToStruct(t *testing.T) {
	type ts struct {
		NoTag      string
		One        string          `mtag:"1"`
		Two        int             `mtag:"2"`
		Three      bool            `mtag:"3"`
		Four       float64         `mtag:"4"`
		Empty      string          `mtag:""`
		Impossible struct{ x int } `mtag:"impossible"`
	}
	type args struct {
		tagKey string
		src    map[string]interface{}
		dst    interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal",
			args: args{
				tagKey: "mtag",
				src: map[string]interface{}{
					"1": "#1",
					"2": 2,
					"3": true,
					"4": 2.3,
					"":  "empty tag!",
				},
				dst: &ts{
					NoTag: "64",
					One:   "boo",
				},
			},
			want: &ts{
				NoTag: "64",
				One:   "#1",
				Two:   2,
				Three: true,
				Four:  2.3,
				Empty: "empty tag!",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapToStruct(tt.args.tagKey, tt.args.src, tt.args.dst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
