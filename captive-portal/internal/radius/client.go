package radius

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/config"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2868"
)

// Client handles RADIUS communication
type Client struct {
	server   string
	secret   []byte
	authPort int
	acctPort int
	timeout  time.Duration
	retries  int
}

// NewClient creates a new RADIUS client
func NewClient(cfg *config.RadiusConfig) *Client {
	return &Client{
		server:   cfg.Server,
		secret:   []byte(cfg.Secret),
		authPort: cfg.AuthPort,
		acctPort: cfg.AcctPort,
		timeout:  cfg.Timeout,
		retries:  cfg.Retries,
	}
}

// Authenticate sends an Access-Request to RADIUS server
func (c *Client) Authenticate(ctx context.Context, username string, vlanID int, userType string) error {
	packet := radius.New(radius.CodeAccessRequest, c.secret)

	// Set username
	if err := rfc2865.UserName_SetString(packet, username); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}

	// Set Tunnel attributes for dynamic VLAN
	// Tunnel-Type = VLAN (13)
	if err := rfc2868.TunnelType_Add(packet, 0, rfc2868.TunnelType(13)); err != nil {
		return fmt.Errorf("failed to set tunnel type: %w", err)
	}

	// Tunnel-Medium-Type = IEEE-802 (6)
	if err := rfc2868.TunnelMediumType_Add(packet, 0, rfc2868.TunnelMediumType_Value_IEEE802); err != nil {
		return fmt.Errorf("failed to set tunnel medium type: %w", err)
	}

	// Tunnel-Private-Group-Id = VLAN ID
	vlanStr := fmt.Sprintf("%d", vlanID)
	if err := rfc2868.TunnelPrivateGroupID_AddString(packet, 0, vlanStr); err != nil {
		return fmt.Errorf("failed to set VLAN ID: %w", err)
	}

	// Set Class attribute for user type
	if err := rfc2865.Class_AddString(packet, userType); err != nil {
		return fmt.Errorf("failed to set class: %w", err)
	}

	// Send packet
	addr := fmt.Sprintf("%s:%d", c.server, c.authPort)

	log.Debug().
		Str("username", username).
		Int("vlan_id", vlanID).
		Str("user_type", userType).
		Str("radius_server", addr).
		Msg("Sending RADIUS Access-Request")

	response, err := c.sendPacket(ctx, packet, addr)
	if err != nil {
		log.Error().Err(err).Msg("RADIUS authentication failed")
		return fmt.Errorf("RADIUS authentication failed: %w", err)
	}

	// Check response code
	switch response.Code {
	case radius.CodeAccessAccept:
		log.Info().
			Str("username", username).
			Int("vlan_id", vlanID).
			Msg("RADIUS Access-Accept received")
		return nil

	case radius.CodeAccessReject:
		log.Warn().
			Str("username", username).
			Msg("RADIUS Access-Reject received")
		return fmt.Errorf("access rejected by RADIUS server")

	case radius.CodeAccessChallenge:
		log.Warn().
			Str("username", username).
			Msg("RADIUS Access-Challenge received (not supported)")
		return fmt.Errorf("access challenge not supported")

	default:
		log.Error().
			Str("username", username).
			Int("code", int(response.Code)).
			Msg("Unexpected RADIUS response code")
		return fmt.Errorf("unexpected response code: %d", response.Code)
	}
}

// SendAccounting sends an Accounting-Request to RADIUS server
func (c *Client) SendAccounting(ctx context.Context, req *models.AccountingRequest) error {
	var statusType rfc2866.AcctStatusType

	switch req.AcctStatusType {
	case models.AcctStatusStart:
		statusType = rfc2866.AcctStatusType_Value_Start
	case models.AcctStatusStop:
		statusType = rfc2866.AcctStatusType_Value_Stop
	case models.AcctStatusUpdate:
		statusType = rfc2866.AcctStatusType_Value_InterimUpdate
	default:
		return fmt.Errorf("invalid accounting status type: %s", req.AcctStatusType)
	}

	packet := radius.New(radius.CodeAccountingRequest, c.secret)

	// Set username
	if err := rfc2865.UserName_SetString(packet, req.Username); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}

	// Set Acct-Status-Type
	if err := rfc2866.AcctStatusType_Set(packet, statusType); err != nil {
		return fmt.Errorf("failed to set status type: %w", err)
	}

	// Set Acct-Session-Id
	if err := rfc2866.AcctSessionID_SetString(packet, req.SessionID); err != nil {
		return fmt.Errorf("failed to set session ID: %w", err)
	}

	// Set NAS-IP-Address
	nasIP := net.ParseIP(req.NASIPAddress)
	if nasIP != nil {
		if err := rfc2865.NASIPAddress_Set(packet, nasIP); err != nil {
			return fmt.Errorf("failed to set NAS IP: %w", err)
		}
	}

	// Set optional attributes
	if req.NASPortId != "" {
		if err := rfc2865.NASIdentifier_SetString(packet, req.NASPortId); err != nil {
			log.Warn().Err(err).Msg("Failed to set NAS Identifier")
		}
	}

	if req.FramedIPAddress != "" {
		framedIP := net.ParseIP(req.FramedIPAddress)
		if framedIP != nil {
			if err := rfc2865.FramedIPAddress_Set(packet, framedIP); err != nil {
				log.Warn().Err(err).Msg("Failed to set Framed IP")
			}
		}
	}

	if req.CalledStationId != "" {
		if err := rfc2865.CalledStationID_SetString(packet, req.CalledStationId); err != nil {
			log.Warn().Err(err).Msg("Failed to set Called Station ID")
		}
	}

	if req.CallingStationId != "" {
		if err := rfc2865.CallingStationID_SetString(packet, req.CallingStationId); err != nil {
			log.Warn().Err(err).Msg("Failed to set Calling Station ID")
		}
	}

	// Set session time for Stop and Update
	if statusType == rfc2866.AcctStatusType_Value_Stop || statusType == rfc2866.AcctStatusType_Value_InterimUpdate {
		if err := rfc2866.AcctSessionTime_Set(packet, rfc2866.AcctSessionTime(req.AcctSessionTime)); err != nil {
			log.Warn().Err(err).Msg("Failed to set session time")
		}

		if err := rfc2866.AcctInputOctets_Set(packet, rfc2866.AcctInputOctets(req.AcctInputOctets)); err != nil {
			log.Warn().Err(err).Msg("Failed to set input octets")
		}

		if err := rfc2866.AcctOutputOctets_Set(packet, rfc2866.AcctOutputOctets(req.AcctOutputOctets)); err != nil {
			log.Warn().Err(err).Msg("Failed to set output octets")
		}
	}

	// Set Acct-Terminate-Cause for Stop
	if statusType == rfc2866.AcctStatusType_Value_Stop && req.TerminateCause != "" {
		// Map terminate cause string to RADIUS value
		terminateCause := mapTerminateCause(req.TerminateCause)
		if err := rfc2866.AcctTerminateCause_Set(packet, terminateCause); err != nil {
			log.Warn().Err(err).Msg("Failed to set terminate cause")
		}
	}

	// Send packet
	addr := fmt.Sprintf("%s:%d", c.server, c.acctPort)

	log.Debug().
		Str("username", req.Username).
		Str("session_id", req.SessionID).
		Str("status_type", req.AcctStatusType).
		Str("radius_server", addr).
		Msg("Sending RADIUS Accounting-Request")

	response, err := c.sendPacket(ctx, packet, addr)
	if err != nil {
		log.Error().Err(err).Msg("RADIUS accounting failed")
		return fmt.Errorf("RADIUS accounting failed: %w", err)
	}

	// Check response code
	if response.Code != radius.CodeAccountingResponse {
		log.Error().
			Int("code", int(response.Code)).
			Msg("Unexpected RADIUS accounting response")
		return fmt.Errorf("unexpected response code: %d", response.Code)
	}

	log.Info().
		Str("username", req.Username).
		Str("status_type", req.AcctStatusType).
		Msg("RADIUS Accounting-Response received")

	return nil
}

// sendPacket sends a RADIUS packet and waits for response
func (c *Client) sendPacket(ctx context.Context, packet *radius.Packet, addr string) (*radius.Packet, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			log.Debug().
				Int("attempt", attempt+1).
				Int("max_retries", c.retries).
				Msg("Retrying RADIUS request")
		}

		// Send packet
		response, err := radius.Exchange(ctx, packet, addr)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on context errors
		if ctx.Err() != nil {
			break
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			// Continue to next retry
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.retries, lastErr)
}

// mapTerminateCause maps terminate cause string to RADIUS value
func mapTerminateCause(cause string) rfc2866.AcctTerminateCause {
	switch cause {
	case "User-Request":
		return rfc2866.AcctTerminateCause_Value_UserRequest
	case "Idle-Timeout":
		return rfc2866.AcctTerminateCause_Value_IdleTimeout
	case "Session-Timeout":
		return rfc2866.AcctTerminateCause_Value_SessionTimeout
	case "Admin-Reset":
		return rfc2866.AcctTerminateCause_Value_AdminReset
	case "NAS-Reboot":
		return rfc2866.AcctTerminateCause_Value_NASReboot
	case "NAS-Error":
		return rfc2866.AcctTerminateCause_Value_NASError
	case "Lost-Carrier":
		return rfc2866.AcctTerminateCause_Value_LostCarrier
	default:
		return rfc2866.AcctTerminateCause_Value_UserRequest
	}
}

// StartAccounting sends an accounting start packet
func (c *Client) StartAccounting(ctx context.Context, username, sessionID, nasIP string) error {
	req := &models.AccountingRequest{
		Username:        username,
		SessionID:       sessionID,
		AcctStatusType:  models.AcctStatusStart,
		NASIPAddress:    nasIP,
	}
	return c.SendAccounting(ctx, req)
}

// StopAccounting sends an accounting stop packet
func (c *Client) StopAccounting(ctx context.Context, username, sessionID, nasIP string, sessionTime, inputOctets, outputOctets int64) error {
	req := &models.AccountingRequest{
		Username:         username,
		SessionID:        sessionID,
		AcctStatusType:   models.AcctStatusStop,
		NASIPAddress:     nasIP,
		AcctSessionTime:  sessionTime,
		AcctInputOctets:  inputOctets,
		AcctOutputOctets: outputOctets,
		TerminateCause:   "User-Request",
	}
	return c.SendAccounting(ctx, req)
}

// UpdateAccounting sends an accounting interim-update packet
func (c *Client) UpdateAccounting(ctx context.Context, username, sessionID, nasIP string, sessionTime, inputOctets, outputOctets int64) error {
	req := &models.AccountingRequest{
		Username:         username,
		SessionID:        sessionID,
		AcctStatusType:   models.AcctStatusUpdate,
		NASIPAddress:     nasIP,
		AcctSessionTime:  sessionTime,
		AcctInputOctets:  inputOctets,
		AcctOutputOctets: outputOctets,
	}
	return c.SendAccounting(ctx, req)
}
