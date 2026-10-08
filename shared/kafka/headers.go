package kafka

import "github.com/twmb/franz-go/pkg/kgo"

// recordHeaders lets OpenTelemetry read and write trace context in Kafka
// record headers, so a consumer's span joins the producer's trace.
type recordHeaders struct {
	record *kgo.Record
}

func (h recordHeaders) Get(key string) string {
	for _, header := range h.record.Headers {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

func (h recordHeaders) Set(key, value string) {
	for i, header := range h.record.Headers {
		if header.Key == key {
			h.record.Headers[i].Value = []byte(value)
			return
		}
	}
	h.record.Headers = append(h.record.Headers, kgo.RecordHeader{Key: key, Value: []byte(value)})
}

func (h recordHeaders) Keys() []string {
	keys := make([]string, len(h.record.Headers))
	for i, header := range h.record.Headers {
		keys[i] = header.Key
	}
	return keys
}
