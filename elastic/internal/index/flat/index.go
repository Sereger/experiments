package flat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type Index struct {
	client *elasticsearch.Client
}

func NewIndex(client *elasticsearch.Client) *Index {
	return &Index{client: client}
}

func (idx *Index) CreateIndex(ctx context.Context, force bool) error {
	initCtx, cnlFn := context.WithTimeout(ctx, time.Second*300)
	defer cnlFn()

	hasIdx, err := idx.client.Indices.Get(
		[]string{indexName},
		idx.client.Indices.Get.WithContext(initCtx),
	)
	if err == nil {
		hasIdx.Body.Close()
	}

	if err == nil && !hasIdx.IsError() && force {
		_, err = idx.client.Indices.Delete(
			[]string{indexName},
			idx.client.Indices.Delete.WithContext(initCtx),
			idx.client.Indices.Delete.WithExpandWildcards("all"),
		)
		if err != nil {
			data, _ := ioutil.ReadAll(hasIdx.Body)
			return fmt.Errorf("index delete [%s]: %w", string(data), err)
		}
	} else if err == nil && !hasIdx.IsError() {
		return nil
	}

	crI, err := idx.client.Indices.Create(
		indexName,
		idx.client.Indices.Create.WithContext(initCtx),
		idx.client.Indices.Create.WithBody(strings.NewReader(config)),
	)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer crI.Body.Close()
	if crI.IsError() {
		data, _ := ioutil.ReadAll(crI.Body)
		return fmt.Errorf("cant create index: [%s]", string(data))
	}

	return nil
}

func (idx *Index) Add(ctx context.Context, docs []*Doc) error {
	buff := &bytes.Buffer{}

	const packSize = 1024
	var i int
	for _, doc := range docs {
		fmt.Fprintf(buff, `{"index":{"_id":"%d"}}%s`, doc.ID, "\n")
		json.NewEncoder(buff).Encode(doc)
		buff.WriteByte('\n')
		i++

		if i%packSize == 0 {
			err := idx.writeBulk(ctx, buff)
			if err != nil {
				return fmt.Errorf("write bulk: %w", err)
			}

			buff.Reset()
		}
	}

	if buff.Len() == 0 {
		return nil
	}

	err := idx.writeBulk(ctx, buff)
	if err != nil {
		return fmt.Errorf("write bulk: %w", err)
	}

	return nil
}

func (idx *Index) writeBulk(ctx context.Context, data io.Reader) error {
	resp, err := idx.client.Bulk(
		data,
		idx.client.Bulk.WithContext(ctx),
		idx.client.Bulk.WithIndex(indexName),
	)
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}

	resp.Body.Close()

	if resp.IsError() {
		return errors.New("bulk error")
	}

	return nil
}
