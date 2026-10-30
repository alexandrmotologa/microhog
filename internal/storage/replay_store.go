package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

// ReplayStore manages storage and retrieval of compressed session replay streams.
type ReplayStore struct {
	db         *DB
	eventStore *EventStore
	encoder    *zstd.Encoder
	decoder    *zstd.Decoder
	mu         sync.Mutex
}

// NewReplayStore creates a new ReplayStore with initialized zstd encoder and decoder.
func NewReplayStore(db *DB, eventStore *EventStore) (*ReplayStore, error) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}

	return &ReplayStore{
		db:         db,
		eventStore: eventStore,
		encoder:    enc,
		decoder:    dec,
	}, nil
}

// SaveChunk compresses and inserts a DOM mutation chunk.
func (r *ReplayStore) SaveChunk(sessionID string, chunkIndex int, rawJSON []byte, eventCount int) error {
	r.mu.Lock()
	compressed := r.encoder.EncodeAll(rawJSON, make([]byte, 0, len(rawJSON)/2))
	r.mu.Unlock()

	query := `
		INSERT INTO replays (session_id, chunk_index, event_count, compressed_data, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(session_id, chunk_index) DO UPDATE SET
			event_count = excluded.event_count,
			compressed_data = excluded.compressed_data
	`
	nowStr := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(query, sessionID, chunkIndex, eventCount, compressed, nowStr)
	if err != nil {
		return fmt.Errorf("failed to save replay chunk: %w", err)
	}

	// Update session status flag
	_ = r.eventStore.MarkSessionHasReplay(sessionID)
	return nil
}

// GetReplayEvents retrieves and decompresses all chunks for a session, merging them into a single JSON array.
func (r *ReplayStore) GetReplayEvents(sessionID string) ([]json.RawMessage, error) {
	query := `
		SELECT chunk_index, compressed_data
		FROM replays
		WHERE session_id = ?
		ORDER BY chunk_index ASC
	`
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query replay chunks: %w", err)
	}
	defer rows.Close()

	var allEvents []json.RawMessage

	for rows.Next() {
		var chunkIndex int
		var compressed []byte
		if err := rows.Scan(&chunkIndex, &compressed); err != nil {
			return nil, err
		}

		r.mu.Lock()
		decompressed, err := r.decoder.DecodeAll(compressed, nil)
		r.mu.Unlock()
		if err != nil {
			return nil, fmt.Errorf("failed to decompress chunk %d: %w", chunkIndex, err)
		}

		// Parse the chunk as an array of JSON objects
		var chunkEvents []json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(decompressed))
		if err := dec.Decode(&chunkEvents); err != nil {
			// If not a JSON array, treat whole decompressed payload as single raw message
			var singleEvent json.RawMessage
			if errSingle := json.Unmarshal(decompressed, &singleEvent); errSingle == nil {
				allEvents = append(allEvents, singleEvent)
			}
			continue
		}

		allEvents = append(allEvents, chunkEvents...)
	}

	return allEvents, nil
}
