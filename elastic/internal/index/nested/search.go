package nested

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type (
	hit struct {
		DocId    string  `json:"_id"`
		Score    float64 `json:"_score"`
		Doc      *Doc    `json:"_source"`
		Children *struct {
			Items *struct {
				Hits *hits `json:"hits"`
			}
		} `json:"inner_hits"`
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
	q := fmt.Sprintf(query, phrase, phrase)
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
		if hit.Children == nil {
			continue
		}
		if hit.Children.Items == nil {
			continue
		}
		if hit.Children.Items.Hits == nil {
			continue
		}
		dosc[i].Items = make([]*Doc, len(hit.Children.Items.Hits.Results))
		for j, child := range hit.Children.Items.Hits.Results {
			dosc[i].Items[j] = child.Doc
		}
	}
	return dosc, nil
}
