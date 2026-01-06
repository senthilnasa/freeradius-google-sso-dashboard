package vlan

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
)

// Service handles VLAN assignment logic
type Service struct {
	db         *sqlx.DB
	rulesCache []models.VLANRule
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
	lastUpdate time.Time
}

// NewService creates a new VLAN service
func NewService(db *sqlx.DB) *Service {
	s := &Service{
		db:       db,
		cacheTTL: 5 * time.Minute, // Cache rules for 5 minutes
	}

	// Load rules on initialization
	if err := s.loadRules(); err != nil {
		log.Error().Err(err).Msg("Failed to load VLAN rules on initialization")
	}

	return s
}

// AssignVLAN determines the appropriate VLAN for a given email address
func (s *Service) AssignVLAN(email string) (*models.VLANAssignment, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	// Parse email
	domain, username := parseEmail(email)
	if domain == "" {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	// Refresh rules cache if needed
	if err := s.refreshCacheIfNeeded(); err != nil {
		log.Warn().Err(err).Msg("Failed to refresh VLAN rules cache")
	}

	// Get rules from cache
	rules := s.getCachedRules()

	log.Debug().
		Str("email", email).
		Str("domain", domain).
		Str("username", username).
		Int("rules_count", len(rules)).
		Msg("Starting VLAN assignment")

	// Find matching rule
	var matchedRule *models.VLANRule
	var defaultRule *models.VLANRule

	for i := range rules {
		rule := &rules[i]

		// Check if domain matches
		if strings.EqualFold(rule.Domain, domain) {
			// Check for key match (suffix)
			if rule.Key != "" {
				// Non-empty key - check for suffix match
				if strings.Contains(strings.ToLower(username), strings.ToLower(rule.Key)) {
					// Found matching rule with suffix
					matchedRule = rule
					log.Info().
						Str("email", email).
						Str("domain", domain).
						Str("key", rule.Key).
						Int("vlan_id", rule.VLANId).
						Str("user_type", rule.UserType).
						Msg("VLAN assigned by suffix match")
					break
				}
			} else {
				// Empty key = default rule for this domain
				if defaultRule == nil || rule.Priority < defaultRule.Priority {
					defaultRule = rule
				}
			}
		}
	}

	// Use matched rule or default rule
	if matchedRule == nil {
		matchedRule = defaultRule
	}

	if matchedRule == nil {
		return nil, fmt.Errorf("no VLAN rule found for email: %s (domain: %s)", email, domain)
	}

	log.Info().
		Str("email", email).
		Int("vlan_id", matchedRule.VLANId).
		Str("user_type", matchedRule.UserType).
		Str("rule_key", matchedRule.Key).
		Msg("VLAN assignment completed")

	return &models.VLANAssignment{
		VLANId:   matchedRule.VLANId,
		UserType: matchedRule.UserType,
		Domain:   matchedRule.Domain,
		Key:      matchedRule.Key,
	}, nil
}

// GetAllRules returns all VLAN rules
func (s *Service) GetAllRules() ([]models.VLANRule, error) {
	if err := s.refreshCacheIfNeeded(); err != nil {
		log.Warn().Err(err).Msg("Failed to refresh cache")
	}
	return s.getCachedRules(), nil
}

// GetRuleByID retrieves a specific VLAN rule by ID
func (s *Service) GetRuleByID(id int) (*models.VLANRule, error) {
	var rule models.VLANRule
	query := `SELECT * FROM vlan_rules WHERE id = ? LIMIT 1`
	err := s.db.Get(&rule, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rule not found")
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return &rule, nil
}

// CreateRule creates a new VLAN rule
func (s *Service) CreateRule(rule *models.VLANRule) error {
	query := `
		INSERT INTO vlan_rules (domain, key, user_type, vlan_id, priority, enabled)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query,
		rule.Domain,
		rule.Key,
		rule.UserType,
		rule.VLANId,
		rule.Priority,
		rule.Enabled,
	)
	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}

	rule.ID = int(id)

	// Invalidate cache
	s.invalidateCache()

	log.Info().
		Int("id", rule.ID).
		Str("domain", rule.Domain).
		Str("key", rule.Key).
		Int("vlan_id", rule.VLANId).
		Msg("VLAN rule created")

	return nil
}

// UpdateRule updates an existing VLAN rule
func (s *Service) UpdateRule(rule *models.VLANRule) error {
	query := `
		UPDATE vlan_rules
		SET domain = ?, key = ?, user_type = ?, vlan_id = ?, priority = ?, enabled = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query,
		rule.Domain,
		rule.Key,
		rule.UserType,
		rule.VLANId,
		rule.Priority,
		rule.Enabled,
		rule.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}

	// Invalidate cache
	s.invalidateCache()

	log.Info().
		Int("id", rule.ID).
		Str("domain", rule.Domain).
		Int("vlan_id", rule.VLANId).
		Msg("VLAN rule updated")

	return nil
}

// DeleteRule deletes a VLAN rule
func (s *Service) DeleteRule(id int) error {
	query := `DELETE FROM vlan_rules WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}

	// Invalidate cache
	s.invalidateCache()

	log.Info().Int("id", id).Msg("VLAN rule deleted")

	return nil
}

// loadRules loads VLAN rules from database
func (s *Service) loadRules() error {
	var rules []models.VLANRule
	query := `
		SELECT * FROM vlan_rules
		WHERE enabled = 1
		ORDER BY priority ASC, domain ASC, ` + "`key`" + ` DESC
	`

	err := s.db.Select(&rules, query)
	if err != nil {
		return fmt.Errorf("failed to load VLAN rules: %w", err)
	}

	s.cacheMutex.Lock()
	s.rulesCache = rules
	s.lastUpdate = time.Now()
	s.cacheMutex.Unlock()

	log.Info().
		Int("count", len(rules)).
		Msg("VLAN rules loaded into cache")

	return nil
}

// refreshCacheIfNeeded refreshes the cache if TTL has expired
func (s *Service) refreshCacheIfNeeded() error {
	s.cacheMutex.RLock()
	needsRefresh := time.Since(s.lastUpdate) > s.cacheTTL
	s.cacheMutex.RUnlock()

	if needsRefresh {
		return s.loadRules()
	}

	return nil
}

// getCachedRules returns a copy of cached rules
func (s *Service) getCachedRules() []models.VLANRule {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	// Return a copy to prevent concurrent modification
	rules := make([]models.VLANRule, len(s.rulesCache))
	copy(rules, s.rulesCache)

	return rules
}

// invalidateCache forces a cache refresh on next access
func (s *Service) invalidateCache() {
	s.cacheMutex.Lock()
	s.lastUpdate = time.Time{} // Set to zero time to force refresh
	s.cacheMutex.Unlock()
}

// parseEmail extracts domain and username from email address
func parseEmail(email string) (domain, username string) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[1], parts[0]
}

// ValidateDomain checks if a domain is allowed
func (s *Service) ValidateDomain(domain string) (bool, error) {
	rules := s.getCachedRules()

	for _, rule := range rules {
		if strings.EqualFold(rule.Domain, domain) {
			return true, nil
		}
	}

	return false, fmt.Errorf("domain not allowed: %s", domain)
}

// GetVLANStats returns statistics about VLAN usage
func (s *Service) GetVLANStats() (map[string]interface{}, error) {
	type vlanStat struct {
		VLANId        int    `db:"vlan_id"`
		UserType      string `db:"user_type"`
		ActiveCount   int    `db:"active_count"`
		TotalSessions int    `db:"total_sessions"`
	}

	var stats []vlanStat
	query := `
		SELECT
			vr.vlan_id,
			vr.user_type,
			COUNT(DISTINCT CASE WHEN s.is_active = 1 THEN s.id END) as active_count,
			COUNT(DISTINCT s.id) as total_sessions
		FROM vlan_rules vr
		LEFT JOIN sessions s ON vr.vlan_id = s.vlan_id
		WHERE vr.enabled = 1
		GROUP BY vr.vlan_id, vr.user_type
		ORDER BY active_count DESC
	`

	err := s.db.Select(&stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get VLAN stats: %w", err)
	}

	result := make(map[string]interface{})
	result["stats"] = stats
	result["total_rules"] = len(s.getCachedRules())

	return result, nil
}
