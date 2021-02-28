package etcd

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func Test(t *testing.T) {
	type (
		subCfg struct {
			XXX string
		}
		cfg struct {
			StrVal      string
			IntVal      int
			Int64Val    int64
			Int32Val    int32
			Float64Val  float64
			StrSliceVal []string
			IntSliceVal []int
			StructValue subCfg
			PtrValue    *subCfg
			DurationVal time.Duration
		}
		testcase struct {
			name        string
			key         string
			etcdValue   string
			expectValue cfg
			wantErr     string
		}
	)
	tests := []*testcase{
		{
			name:      "string value",
			key:       "TEST_STRVAL",
			etcdValue: "abc",
			expectValue: cfg{
				StrVal:      "abc",
				IntSliceVal: []int{},
				StrSliceVal: []string{},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "int value",
			key:       "TEST_INTVAL",
			etcdValue: "42",
			expectValue: cfg{
				IntVal:      42,
				IntSliceVal: []int{},
				StrSliceVal: []string{},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "float value",
			key:       "TEST_FLOAT64VAL",
			etcdValue: "42.0987",
			expectValue: cfg{
				Float64Val:  42.0987,
				IntSliceVal: []int{},
				StrSliceVal: []string{},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "slices int value",
			key:       "TEST_INTSLICEVAL",
			etcdValue: "1, 2,3\t,\n99",
			expectValue: cfg{
				IntSliceVal: []int{1, 2, 3, 99},
				StrSliceVal: []string{},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "slices str value",
			key:       "TEST_STRSLICEVAL",
			etcdValue: "xxx, yyy, \nabc,\tAAA\t",
			expectValue: cfg{
				IntSliceVal: []int{},
				StrSliceVal: []string{"xxx", "yyy", "abc", "AAA"},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "sub struct value",
			key:       "TEST_STRUCTVALUE_XXX",
			etcdValue: "zzz",
			expectValue: cfg{
				StructValue: subCfg{XXX: "zzz"},
				IntSliceVal: []int{},
				StrSliceVal: []string{},
				PtrValue:    &subCfg{},
			},
		},
		{
			name:      "sub struct ptr value",
			key:       "TEST_PTRVALUE_XXX",
			etcdValue: "zzz",
			expectValue: cfg{
				PtrValue:    &subCfg{XXX: "zzz"},
				IntSliceVal: []int{},
				StrSliceVal: []string{},
			},
		},
		{
			name:      "duration value",
			key:       "TEST_DURATIONVAL",
			etcdValue: "1s",
			expectValue: cfg{
				PtrValue:    &subCfg{},
				IntSliceVal: []int{},
				StrSliceVal: []string{},
				DurationVal: time.Second,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testConfig := &cfg{}
			params, _ := parseCfg(reflect.ValueOf(testConfig), "TEST")
			err := setValue(params[test.key], test.etcdValue)
			if test.wantErr != "" {
				require.Contains(t, err.Error(), test.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expectValue, *testConfig)
		})
	}
}
