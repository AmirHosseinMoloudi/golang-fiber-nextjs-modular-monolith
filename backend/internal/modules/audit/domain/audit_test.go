package domain_test

import (
	"testing"

	"github.com/golang-todo/modular-monolith/internal/modules/audit/domain"
)

func TestAudit_HashChain(t *testing.T) {
	prev := "0000000000000000000000000000000000000000000000000000000000000000"
	rec1 := domain.NewAuditRecord("tenant-1", "user-1", "admin@todo.local", domain.ActionCreated, "TODO", "todo-1", `{"title":"Task 1"}`, prev)

	if rec1.RecordHash == "" {
		t.Fatalf("record hash should not be empty")
	}

	rec2 := domain.NewAuditRecord("tenant-1", "user-1", "admin@todo.local", domain.ActionUpdated, "TODO", "todo-1", `{"title":"Task 1 updated"}`, rec1.RecordHash)

	if rec2.PreviousHash != rec1.RecordHash {
		t.Errorf("expected prev hash %s, got %s", rec1.RecordHash, rec2.PreviousHash)
	}

	// Verify tamper detection
	tamperedRec := *rec2
	tamperedRec.Payload = `{"title":"Hacked"}`
	calc := domain.CalculateHash(tamperedRec.PreviousHash, tamperedRec.Timestamp, tamperedRec.Action, tamperedRec.EntityType, tamperedRec.EntityID, tamperedRec.ActorID, tamperedRec.Payload)

	if calc == tamperedRec.RecordHash {
		t.Errorf("expected tamper detection to yield different hash, but hashes matched")
	}
}
