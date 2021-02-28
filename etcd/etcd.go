package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ConfigWatcher struct {
	cli         *clientv3.Client
	serviceName string
	close       chan struct{}
}

func New(endpoints []string, service string, timeout time.Duration) (*ConfigWatcher, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	return &ConfigWatcher{cli: cli, serviceName: service, close: make(chan struct{})}, nil
}

func (cw *ConfigWatcher) Close() error {
	close(cw.close)
	return cw.cli.Close()
}

type watcher func(key string, newValue interface{})

func (cw *ConfigWatcher) SyncConfig(ctx context.Context, cfg interface{}, w watcher) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return errors.New("config should be pointer")
	}

	data, defVals := parseCfg(v, "")

	resp, err := cw.cli.Get(ctx, cw.serviceName+"/", clientv3.WithPrefix())
	if err != nil {
		return err
	}
	etcdData := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		etcdData[cw.key(kv.Key)] = string(kv.Value)
	}

	for k, val := range data {
		if v, ok := etcdData[k]; ok {
			err := setValue(val, v)
			if err != nil {
				return err
			}
		}
	}

	wch := cw.cli.Watch(ctx, cw.serviceName+"/", clientv3.WithPrefix())
	go func() {
		for {
			select {
			case newVal := <-wch:
				for _, e := range newVal.Events {
					key, val := cw.key(e.Kv.Key), string(e.Kv.Value)
					cfgAttr, ok := data[key]
					if !ok {
						continue
					}

					if val == "" {
						val = defVals[key]
					}
					err := setValue(cfgAttr, val)
					if err != nil {
						log.Printf("Set value err: %s", err)
						continue
					}

					w(key, val)
				}
			case <-cw.close:
				return
			}
		}
	}()

	return nil
}

func (cw *ConfigWatcher) key(k []byte) string {
	return strings.ToUpper(string(k)[len(cw.serviceName+"/"):])
}
func setValue(v reflect.Value, val string) error {
	switch v.Kind() {
	case reflect.Int:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int(i)))
	case reflect.Int32:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int32(i)))
	case reflect.Int64:
		if v.Type().String() == "time.Duration" {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}

			v.Set(reflect.ValueOf(d))
			return nil
		}
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(i))
	case reflect.String:
		v.Set(reflect.ValueOf(val))
	case reflect.Float64:
		i, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(i))
	case reflect.Slice:
		strSlice := strings.Split(val, ",")
		switch v.Type().Elem().Kind() {
		case reflect.Int:
			data := make([]int, len(strSlice))
			for j, c := range strSlice {
				c = strings.Trim(c, " \n\t")
				a, err := strconv.ParseInt(c, 10, 64)
				if err != nil {
					return err
				}

				data[j] = int(a)
			}
			v.Set(reflect.ValueOf(data))
		case reflect.Int32:
			data := make([]int32, len(strSlice))
			for j, c := range strSlice {
				c = strings.Trim(c, " \n\t")
				a, err := strconv.ParseInt(c, 10, 64)
				if err != nil {
					return err
				}

				data[j] = int32(a)
			}
			v.Set(reflect.ValueOf(data))
		case reflect.Int64:
			data := make([]int64, len(strSlice))
			for j, c := range strSlice {
				c = strings.Trim(c, " \n\t")
				a, err := strconv.ParseInt(c, 10, 64)
				if err != nil {
					return err
				}

				data[j] = a
			}
			v.Set(reflect.ValueOf(data))
		case reflect.String:
			data := make([]string, len(strSlice))
			for j, c := range strSlice {
				data[j] = strings.Trim(c, " \n\t")
			}
			v.Set(reflect.ValueOf(data))
		case reflect.Float64:
			data := make([]float64, len(strSlice))
			for j, c := range strSlice {
				c = strings.Trim(c, " \n\t")
				a, err := strconv.ParseFloat(c, 64)
				if err != nil {
					return err
				}

				data[j] = a
			}
			v.Set(reflect.ValueOf(data))
		}
	case reflect.Struct:
		_, ok := v.Interface().(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type: [%T]", v.Interface())
		}
		avalFormats := []string{time.RFC3339, "2006-01-02T15:04:05"}
		for _, f := range avalFormats {
			t, err := time.ParseInLocation(f, val, time.Local)
			if err == nil {
				v.Set(reflect.ValueOf(t))
				return nil
			}
		}
	case reflect.Bool:
		bv, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(bv))
	}

	return nil
}

func parseCfg(v reflect.Value, prefix string) (map[string]reflect.Value, map[string]string) {
	result := make(map[string]reflect.Value)
	defVals := make(map[string]string)

	switch v.Kind() {
	case reflect.Struct:
		if v.Type().String() == "time.Time" {
			result[prefix] = v
			defVals[prefix] = defValue(v)
			return result, defVals
		}

		for i := 0; i < v.NumField(); i++ {
			key, _ := v.Type().Field(i).Tag.Lookup("etcd")
			key = strings.ToUpper(key)
			switch key {
			case "-":
				continue
			case "":
				key = strings.TrimLeft(prefix+"_"+strings.ToUpper(v.Type().Field(i).Name), "_")
			}
			data, defV := parseCfg(v.Field(i), key)
			for k, val := range data {
				result[k] = val
			}
			for k, val := range defV {
				defVals[k] = val
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		data, defV := parseCfg(v.Elem(), prefix)
		for k, val := range data {
			result[k] = val
		}
		for k, val := range defV {
			defVals[k] = val
		}
	case reflect.Slice:
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		kind := v.Type().Elem().Kind()
		if kind >= reflect.Array && kind != reflect.String {
			break
		}
		result[prefix] = v
		defVals[prefix] = defValue(v)
	case reflect.Array:
		kind := v.Type().Elem().Kind()
		if kind >= reflect.Array && kind != reflect.String {
			break
		}
		result[prefix] = v
		defVals[prefix] = defValue(v)
	default:
		if v.Kind() < reflect.Array || v.Kind() == reflect.String {
			result[prefix] = v
			defVals[prefix] = defValue(v)
		}
	}

	return result, defVals
}

func defValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Float64, reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Int16:
			data := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				data[i] = strconv.FormatInt(v.Index(i).Int(), 10)
			}
			return strings.Join(data, ",")
		case reflect.Float64, reflect.Float32:
			data := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				data[i] = strconv.FormatFloat(v.Index(i).Float(), 'f', -1, 64)
			}
			return strings.Join(data, ",")
		case reflect.String:
			data := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				data[i] = v.Index(i).String()
			}
			return strings.Join(data, ",")
		case reflect.Bool:
			data := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				data[i] = strconv.FormatBool(v.Index(i).Bool())
			}
			return strings.Join(data, ",")
		}
	}
	switch v.Type().String() {
	case "time.Time":
		return v.Interface().(time.Time).Format(time.RFC3339)
	}
	return ""
}
