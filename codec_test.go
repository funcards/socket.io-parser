package sio_parser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	type testCase struct {
		arg  Packet
		want any
	}
	cases := []testCase{
		{Packet{Type: Connect, Nsp: "/foo"}, []any{"0/foo,"}},
		{Packet{Type: Event, Nsp: "/", Data: []string{"message", "hello"}}, []any{"2[\"message\",\"hello\"]"}},
		{Packet{Type: Event, Nsp: "/boo", Data: map[string]string{"token": "123"}}, []any{"2/boo,{\"token\":\"123\"}"}},
		{Packet{Type: Disconnect}, []any{"1"}},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			got := tc.arg.Encode()
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("Expected '%v', but got '%v'", tc.want, got)
			}
		})
	}
}
