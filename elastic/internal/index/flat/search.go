package flat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type (
	hit struct {
		DocId string  `json:"_id"`
		Score float64 `json:"_score"`
		Doc   *Doc    `json:"_source"`
	}

	hits struct {
		Total struct {
			Val int `json:"value"`
		} `json:"total"`
		Results []*hit `json:"hits"`
	}

	SearchResult struct {
		Time int   `json:"took"`
		Hits *hits `json:"hits"`
	}
)

func (idx *Index) Search(ctx context.Context, phrase string) ([]*Doc, error) {
	q := fmt.Sprintf(query, phrase)
	resp, err := idx.client.Search(
		idx.client.Search.WithIndex(indexName),
		idx.client.Search.WithBody(strings.NewReader(q)),
		idx.client.Search.WithTrackScores(true),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	result := new(SearchResult)

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	dosc := make([]*Doc, len(result.Hits.Results))
	for i, hit := range result.Hits.Results {
		dosc[i] = hit.Doc
	}
	return dosc, nil
}
