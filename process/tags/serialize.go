package tags

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// Serialize will serialize tagsCount in a way that Elastic Search expects a bulk request
func (tc *tagsCount) Serialize() ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for tag, count := range tc.tags {
		if tag == "" {
			continue
		}

		base64Tag := base64.StdEncoding.EncodeToString([]byte(tag))
		meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, base64Tag, "\n"))
		serializedDataStr := fmt.Sprintf(`{"script": {"source": "ctx._source.count += params.count","lang": "painless","params": {"count": %d}},"upsert": {"count": %d}}`, count, count)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
