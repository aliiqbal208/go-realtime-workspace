package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"go-realtime-workspace/config"
	"go-realtime-workspace/models"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// MessageRepository handles chat message storage in Redis.
type MessageRepository struct {
	client *redis.Client
	cfg    config.RedisConfig
}

// NewMessageRepository creates a new message repository.
func NewMessageRepository(client *redis.Client, cfg config.RedisConfig) *MessageRepository {
	return &MessageRepository{
		client: client,
		cfg:    cfg,
	}
}

// Save stores a chat message in Redis.
func (r *MessageRepository) Save(ctx context.Context, msg models.ChatMessage) error {
	// Generate ID if not provided
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Serialize message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	// Create Redis key for the group's message list
	key := fmt.Sprintf("messages:%s:%s", msg.OrgID, msg.GroupID)

	// Use a pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Add message to sorted set (score is timestamp for ordering)
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(msg.Timestamp.Unix()),
		Member: data,
	})

	// Trim to keep only MaxMessages
	pipe.ZRemRangeByRank(ctx, key, 0, -r.cfg.MaxMessages-1)

	// Set TTL on the key
	pipe.Expire(ctx, key, r.cfg.MessageTTL)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}

	return nil
}

// GetHistory retrieves message history for a group.
func (r *MessageRepository) GetHistory(ctx context.Context, orgID, groupID string, limit int64) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > r.cfg.MaxMessages {
		limit = r.cfg.MaxMessages
	}

	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)

	// Get messages in reverse chronological order (most recent first)
	results, err := r.client.ZRevRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting message history: %w", err)
	}

	messages := make([]models.ChatMessage, 0, len(results))
	for _, data := range results {
		var msg models.ChatMessage
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			// Skip malformed messages
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetHistoryAfter retrieves messages after a specific timestamp.
func (r *MessageRepository) GetHistoryAfter(ctx context.Context, orgID, groupID string, after time.Time, limit int64) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > r.cfg.MaxMessages {
		limit = r.cfg.MaxMessages
	}

	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)

	// Get messages with score (timestamp) greater than 'after'
	results, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   fmt.Sprintf("%d", after.Unix()),
		Max:   "+inf",
		Count: limit,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("error getting messages after timestamp: %w", err)
	}

	messages := make([]models.ChatMessage, 0, len(results))
	for _, data := range results {
		var msg models.ChatMessage
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetHistoryBetween retrieves messages between two timestamps.
func (r *MessageRepository) GetHistoryBetween(ctx context.Context, orgID, groupID string, start, end time.Time, limit int64) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > r.cfg.MaxMessages {
		limit = r.cfg.MaxMessages
	}

	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)

	results, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   fmt.Sprintf("%d", start.Unix()),
		Max:   fmt.Sprintf("%d", end.Unix()),
		Count: limit,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("error getting messages between timestamps: %w", err)
	}

	messages := make([]models.ChatMessage, 0, len(results))
	for _, data := range results {
		var msg models.ChatMessage
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Count returns the total number of messages in a group.
func (r *MessageRepository) Count(ctx context.Context, orgID, groupID string) (int64, error) {
	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)
	return r.client.ZCard(ctx, key).Result()
}

// DeleteOld deletes messages older than the specified duration.
func (r *MessageRepository) DeleteOld(ctx context.Context, orgID, groupID string, olderThan time.Duration) (int64, error) {
	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)
	cutoff := time.Now().Add(-olderThan).Unix()

	return r.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff)).Result()
}

// DeleteGroup deletes all messages for a group.
func (r *MessageRepository) DeleteGroup(ctx context.Context, orgID, groupID string) error {
	key := fmt.Sprintf("messages:%s:%s", orgID, groupID)
	return r.client.Del(ctx, key).Err()
}
