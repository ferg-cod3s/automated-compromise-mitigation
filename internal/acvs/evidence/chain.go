// Package evidence provides evidence chain generation and verification for ACVS.
// Evidence chains create cryptographically-signed audit trails that prove
// good-faith compliance with Terms of Service.
package evidence

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChainGenerator implements the EvidenceChainGenerator interface.
// It creates a Merkle-tree-like chain of evidence entries, each cryptographically
// signed and linked to the previous entry.
type ChainGenerator struct {
	mu         sync.RWMutex
	entries    map[string]*acmv1.EvidenceChainEntry
	chain      []string // Ordered list of entry IDs
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	chainHead  string // ID of most recent entry
}

// NewChainGenerator creates a new evidence chain generator.
// It generates a new Ed25519 key pair for signing evidence.
func NewChainGenerator() (*ChainGenerator, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}

	return &ChainGenerator{
		entries:    make(map[string]*acmv1.EvidenceChainEntry),
		chain:      make([]string, 0),
		publicKey:  pub,
		privateKey: priv,
	}, nil
}

// NewChainGeneratorWithKeys creates a generator with existing keys.
func NewChainGeneratorWithKeys(publicKey, privateKey ed25519.PrivateKey) (*ChainGenerator, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	pub := privateKey.Public().(ed25519.PublicKey)

	return &ChainGenerator{
		entries:    make(map[string]*acmv1.EvidenceChainEntry),
		chain:      make([]string, 0),
		publicKey:  pub,
		privateKey: privateKey,
	}, nil
}

// AddEntry adds a new entry to the evidence chain.
func (g *ChainGenerator) AddEntry(ctx context.Context, entry *Entry) (string, error) {
	if entry == nil {
		return "", fmt.Errorf("entry cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Generate entry ID
	entryID := g.generateEntryID(entry)

	// Get previous entry ID
	previousID := g.chainHead

	// Compute chain hash (links to previous entry)
	chainHash := g.computeChainHash(entryID, previousID)

	// Serialize evidence data
	evidenceDataJSON, err := json.Marshal(entry.EvidenceData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize evidence data: %w", err)
	}

	// Create the proto message
	protoEntry := &acmv1.EvidenceChainEntry{
		Id:                entryID,
		Timestamp:         timestamppb.New(time.Now()),
		EventType:         entry.EventType,
		Site:              entry.Site,
		CredentialIdHash:  entry.CredentialIDHash,
		Action:            entry.Action,
		ValidationResult:  entry.ValidationResult,
		CrcId:             entry.CRCID,
		AppliedRuleIds:    entry.AppliedRuleIDs,
		EvidenceData:      string(evidenceDataJSON),
		PreviousEntryId:   previousID,
		ChainHash:         chainHash,
	}

	// Sign the entry
	signature, err := g.signEntry(protoEntry)
	if err != nil {
		return "", fmt.Errorf("failed to sign entry: %w", err)
	}

	protoEntry.Signature = signature

	// Store entry
	g.entries[entryID] = protoEntry
	g.chain = append(g.chain, entryID)
	g.chainHead = entryID

	return entryID, nil
}

// GetEntry retrieves an evidence entry by ID.
func (g *ChainGenerator) GetEntry(ctx context.Context, entryID string) (*acmv1.EvidenceChainEntry, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	entry, exists := g.entries[entryID]
	if !exists {
		return nil, fmt.Errorf("entry not found: %s", entryID)
	}

	return entry, nil
}

// Export exports evidence entries for a time range or credential.
func (g *ChainGenerator) Export(ctx context.Context, req *ExportRequest) ([]*acmv1.EvidenceChainEntry, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var results []*acmv1.EvidenceChainEntry

	for _, entryID := range g.chain {
		entry := g.entries[entryID]

		// Apply filters
		if req.CredentialID != "" && entry.CredentialIdHash != req.CredentialID {
			continue
		}

		if !req.StartTime.IsZero() && entry.Timestamp.AsTime().Before(req.StartTime) {
			continue
		}

		if !req.EndTime.IsZero() && entry.Timestamp.AsTime().After(req.EndTime) {
			continue
		}

		results = append(results, entry)
	}

	return results, nil
}

// Verify verifies the integrity of an evidence entry.
func (g *ChainGenerator) Verify(ctx context.Context, entry *acmv1.EvidenceChainEntry) (bool, error) {
	if entry == nil {
		return false, fmt.Errorf("entry cannot be nil")
	}

	// Verify signature
	sigBytes, err := hex.DecodeString(entry.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature encoding: %w", err)
	}

	message := g.constructSignatureMessage(entry)
	valid := ed25519.Verify(g.publicKey, message, sigBytes)

	return valid, nil
}

// VerifyChain verifies the integrity of the entire evidence chain.
func (g *ChainGenerator) VerifyChain(ctx context.Context) (bool, []string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var errors []string

	// Verify each entry's signature
	for i, entryID := range g.chain {
		entry := g.entries[entryID]

		// Verify signature
		valid, err := g.Verify(ctx, entry)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Entry %s: signature verification error: %v", entryID, err))
			continue
		}

		if !valid {
			errors = append(errors, fmt.Sprintf("Entry %s: invalid signature", entryID))
		}

		// Verify chain linkage (except for first entry)
		if i > 0 {
			expectedPrevious := g.chain[i-1]
			if entry.PreviousEntryId != expectedPrevious {
				errors = append(errors, fmt.Sprintf("Entry %s: broken chain link (expected %s, got %s)",
					entryID, expectedPrevious, entry.PreviousEntryId))
			}

			// Verify chain hash
			expectedHash := g.computeChainHash(entryID, expectedPrevious)
			if entry.ChainHash != expectedHash {
				errors = append(errors, fmt.Sprintf("Entry %s: invalid chain hash", entryID))
			}
		}
	}

	return len(errors) == 0, errors, nil
}

// GetChainHead returns the most recent evidence entry ID.
func (g *ChainGenerator) GetChainHead(ctx context.Context) (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.chainHead == "" {
		return "", fmt.Errorf("chain is empty")
	}

	return g.chainHead, nil
}

// generateEntryID generates a unique ID for an evidence entry.
func (g *ChainGenerator) generateEntryID(entry *Entry) string {
	// Format: EVD-{timestamp}-{short-hash}
	now := time.Now().Unix()
	data := fmt.Sprintf("%d:%s:%s:%v", now, entry.Site, entry.CredentialIDHash, entry.EventType)
	hash := sha256.Sum256([]byte(data))
	shortHash := hex.EncodeToString(hash[:8])
	return fmt.Sprintf("EVD-%d-%s", now, shortHash)
}

// computeChainHash computes the chain hash linking this entry to the previous one.
func (g *ChainGenerator) computeChainHash(entryID, previousID string) string {
	if previousID == "" {
		// First entry in chain
		return g.hashString(entryID)
	}

	// Hash(currentID + previousID)
	data := entryID + previousID
	return g.hashString(data)
}

// hashString computes SHA-256 hash of a string.
func (g *ChainGenerator) hashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// signEntry signs an evidence entry using Ed25519.
func (g *ChainGenerator) signEntry(entry *acmv1.EvidenceChainEntry) (string, error) {
	message := g.constructSignatureMessage(entry)
	signature := ed25519.Sign(g.privateKey, message)
	return hex.EncodeToString(signature), nil
}

// constructSignatureMessage constructs the message to be signed.
// Format: entryID|timestamp|site|credentialHash|eventType|validationResult|crcID|chainHash
func (g *ChainGenerator) constructSignatureMessage(entry *acmv1.EvidenceChainEntry) []byte {
	message := fmt.Sprintf("%s|%d|%s|%s|%s|%s|%s|%s",
		entry.Id,
		entry.Timestamp.AsTime().Unix(),
		entry.Site,
		entry.CredentialIdHash,
		entry.EventType.String(),
		entry.ValidationResult.String(),
		entry.CrcId,
		entry.ChainHash,
	)
	return []byte(message)
}

// GetPublicKey returns the public key for signature verification.
func (g *ChainGenerator) GetPublicKey() []byte {
	return g.publicKey
}

// GetChainLength returns the number of entries in the chain.
func (g *ChainGenerator) GetChainLength() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.chain)
}

// GetEntryAtIndex returns the entry at the specified index (0-based).
func (g *ChainGenerator) GetEntryAtIndex(index int) (*acmv1.EvidenceChainEntry, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if index < 0 || index >= len(g.chain) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}

	entryID := g.chain[index]
	return g.entries[entryID], nil
}

// ExportToJSON exports the entire evidence chain to JSON format.
func (g *ChainGenerator) ExportToJSON() (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	type chainExport struct {
		PublicKey string                        `json:"public_key"`
		Chain     []*acmv1.EvidenceChainEntry   `json:"entries"`
		Length    int                           `json:"length"`
		ExportedAt time.Time                    `json:"exported_at"`
	}

	entries := make([]*acmv1.EvidenceChainEntry, 0, len(g.chain))
	for _, entryID := range g.chain {
		entries = append(entries, g.entries[entryID])
	}

	export := chainExport{
		PublicKey:  hex.EncodeToString(g.publicKey),
		Chain:      entries,
		Length:     len(g.chain),
		ExportedAt: time.Now(),
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal chain: %w", err)
	}

	return string(data), nil
}

// Clear removes all entries from the chain.
// WARNING: This is destructive and should only be used for testing or when
// the user explicitly disables ACVS with clear_cache=true.
func (g *ChainGenerator) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.entries = make(map[string]*acmv1.EvidenceChainEntry)
	g.chain = make([]string, 0)
	g.chainHead = ""
}
