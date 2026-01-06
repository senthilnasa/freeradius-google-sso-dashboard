package models

import (
	"database/sql"
	"time"
)

// VLANRule represents a VLAN assignment rule
type VLANRule struct {
	ID        int       `db:"id" json:"id"`
	Domain    string    `db:"domain" json:"domain"`
	Key       string    `db:"key" json:"key"`
	UserType  string    `db:"user_type" json:"user_type"`
	VLANId    int       `db:"vlan_id" json:"vlan_id"`
	Priority  int       `db:"priority" json:"priority"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID           int64          `db:"id" json:"id"`
	SessionID    string         `db:"session_id" json:"session_id"`
	Username     string         `db:"username" json:"username"`
	Email        string         `db:"email" json:"email"`
	EmailDomain  string         `db:"email_domain" json:"email_domain"`
	VLANId       int            `db:"vlan_id" json:"vlan_id"`
	UserType     string         `db:"user_type" json:"user_type"`
	IPAddress    sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	MACAddress   sql.NullString `db:"mac_address" json:"mac_address,omitempty"`
	NASIp        sql.NullString `db:"nas_ip" json:"nas_ip,omitempty"`
	UserAgent    sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	LastActivity time.Time      `db:"last_activity" json:"last_activity"`
	ExpiresAt    sql.NullTime   `db:"expires_at" json:"expires_at,omitempty"`
	IsActive     bool           `db:"is_active" json:"is_active"`
}

// RadAcct represents a RADIUS accounting record
type RadAcct struct {
	RadAcctId         int64          `db:"radacctid" json:"radacctid"`
	AcctSessionId     string         `db:"acctsessionid" json:"acctsessionid"`
	AcctUniqueId      string         `db:"acctuniqueid" json:"acctuniqueid"`
	Username          string         `db:"username" json:"username"`
	Realm             sql.NullString `db:"realm" json:"realm,omitempty"`
	NASIPAddress      string         `db:"nasipaddress" json:"nasipaddress"`
	NASPortId         sql.NullString `db:"nasportid" json:"nasportid,omitempty"`
	NASPortType       sql.NullString `db:"nasporttype" json:"nasporttype,omitempty"`
	AcctStartTime     sql.NullTime   `db:"acctstarttime" json:"acctstarttime,omitempty"`
	AcctUpdateTime    sql.NullTime   `db:"acctupdatetime" json:"acctupdatetime,omitempty"`
	AcctStopTime      sql.NullTime   `db:"acctstoptime" json:"acctstoptime,omitempty"`
	AcctSessionTime   sql.NullInt64  `db:"acctsessiontime" json:"acctsessiontime,omitempty"`
	AcctInputOctets   sql.NullInt64  `db:"acctinputoctets" json:"acctinputoctets,omitempty"`
	AcctOutputOctets  sql.NullInt64  `db:"acctoutputoctets" json:"acctoutputoctets,omitempty"`
	CalledStationId   string         `db:"calledstationid" json:"calledstationid"`
	CallingStationId  string         `db:"callingstationid" json:"callingstationid"`
	AcctTerminateCause string        `db:"acctterminatecause" json:"acctterminatecause"`
	FramedIPAddress   string         `db:"framedipaddress" json:"framedipaddress"`
	VLANId            sql.NullInt64  `db:"vlan_id" json:"vlan_id,omitempty"`
	UserType          sql.NullString `db:"user_type" json:"user_type,omitempty"`
}

// RadPostAuth represents a RADIUS post-authentication record
type RadPostAuth struct {
	ID           int64          `db:"id" json:"id"`
	Username     string         `db:"username" json:"username"`
	Pass         string         `db:"pass" json:"-"` // Don't include in JSON
	Reply        string         `db:"reply" json:"reply"`
	AuthDate     time.Time      `db:"authdate" json:"authdate"`
	NASIPAddress string         `db:"nasipaddress" json:"nasipaddress"`
	NASPortId    sql.NullString `db:"nasportid" json:"nasportid,omitempty"`
	VLANId       sql.NullInt64  `db:"vlan_id" json:"vlan_id,omitempty"`
	UserType     sql.NullString `db:"user_type" json:"user_type,omitempty"`
	EmailDomain  sql.NullString `db:"email_domain" json:"email_domain,omitempty"`
	ErrorMessage sql.NullString `db:"error_message" json:"error_message,omitempty"`
}

// AuthFailure represents an authentication failure record
type AuthFailure struct {
	ID            int64          `db:"id" json:"id"`
	Username      string         `db:"username" json:"username"`
	Email         sql.NullString `db:"email" json:"email,omitempty"`
	EmailDomain   sql.NullString `db:"email_domain" json:"email_domain,omitempty"`
	FailureReason string         `db:"failure_reason" json:"failure_reason"`
	ErrorType     string         `db:"error_type" json:"error_type"`
	IPAddress     sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	NASIp         sql.NullString `db:"nas_ip" json:"nas_ip,omitempty"`
	UserAgent     sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	Details       sql.NullString `db:"details" json:"details,omitempty"` // JSON field
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

// SystemLog represents a system log entry
type SystemLog struct {
	ID        int64          `db:"id" json:"id"`
	Level     string         `db:"level" json:"level"`
	Component string         `db:"component" json:"component"`
	Message   string         `db:"message" json:"message"`
	Context   sql.NullString `db:"context" json:"context,omitempty"` // JSON field
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

// UserInfo represents Google OAuth user information
type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	HD            string `json:"hd"` // Hosted domain
}

// VLANAssignment represents the result of VLAN assignment
type VLANAssignment struct {
	VLANId   int    `json:"vlan_id"`
	UserType string `json:"user_type"`
	Domain   string `json:"domain"`
	Key      string `json:"key"`
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	IPAddress   string `json:"ip_address,omitempty"`
	MACAddress  string `json:"mac_address,omitempty"`
	NASIp       string `json:"nas_ip,omitempty"`
	NASPortId   string `json:"nas_port_id,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message"`
	VLANId       int            `json:"vlan_id,omitempty"`
	UserType     string         `json:"user_type,omitempty"`
	SessionID    string         `json:"session_id,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	ErrorCode    string         `json:"error_code,omitempty"`
}

// AccountingRequest represents a RADIUS accounting request
type AccountingRequest struct {
	Username         string `json:"username"`
	SessionID        string `json:"session_id"`
	AcctStatusType   string `json:"acct_status_type"` // Start, Stop, Update
	NASIPAddress     string `json:"nas_ip_address"`
	NASPortId        string `json:"nas_port_id,omitempty"`
	FramedIPAddress  string `json:"framed_ip_address,omitempty"`
	CalledStationId  string `json:"called_station_id,omitempty"`
	CallingStationId string `json:"calling_station_id,omitempty"`
	AcctSessionTime  int64  `json:"acct_session_time,omitempty"`
	AcctInputOctets  int64  `json:"acct_input_octets,omitempty"`
	AcctOutputOctets int64  `json:"acct_output_octets,omitempty"`
	TerminateCause   string `json:"terminate_cause,omitempty"`
}

// AccountingResponse represents a RADIUS accounting response
type AccountingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Error types for authentication failures
const (
	ErrorTypeInvalidDomain  = "INVALID_DOMAIN"
	ErrorTypeInvalidSuffix  = "INVALID_SUFFIX"
	ErrorTypeNoVLANRule     = "NO_VLAN_RULE"
	ErrorTypeOAuthError     = "OAUTH_ERROR"
	ErrorTypeRadiusError    = "RADIUS_ERROR"
	ErrorTypeOther          = "OTHER"
)

// RADIUS accounting status types
const (
	AcctStatusStart  = "Start"
	AcctStatusStop   = "Stop"
	AcctStatusUpdate = "Interim-Update"
)
