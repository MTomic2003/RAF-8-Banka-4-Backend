package model

import "time"

// InboundMessage records a message that has been received and processed from
// another bank. Idempotency is enforced via the composite key
// (peer_routing_number, locally_generated_key). The cached response is
// returned on retransmissions so the peer sees the same outcome it would
// have seen the first time.
type InboundMessage struct {
	PeerRoutingNumber   int       `gorm:"primaryKey;column:peer_routing_number"`
	LocallyGeneratedKey string    `gorm:"primaryKey;size:64;column:locally_generated_key"`
	MessageType         string    `gorm:"not null;size:32;column:message_type"`
	RequestBody         []byte    `gorm:"type:jsonb;not null;column:request_body"`
	ResponseStatus      int       `gorm:"not null;column:response_status"`
	ResponseBody        []byte    `gorm:"type:jsonb;column:response_body"`
	ProcessedAt         time.Time `gorm:"not null;column:processed_at"`
	CreatedAt           time.Time
}

func (InboundMessage) TableName() string { return "interbank_inbound_messages" }
