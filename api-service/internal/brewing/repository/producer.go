package repository

import (
	dto "alla/shared/DTO"
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

const brewTopic = "brew-jobs"

type brewMessage struct {
	JobUUID string `json:"job_uuid"`
}

func (r *BrewingRepo) Boiled(ctx context.Context, req dto.JobUUIDDTO) error {

	msg, err := json.Marshal(brewMessage{JobUUID: req.JobUUID})
	if err != nil {
		return fmt.Errorf("Boiled: marshal: %w", err)
	}

	r.kafkaClient.ProduceSync(ctx, &kgo.Record{
		Topic: brewTopic,
		Value: msg,
	})
	return nil
}
