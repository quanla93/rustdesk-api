package utils

import (
	"errors"
	"sync"
	"time"
)

// Security policy configuration
type SecurityPolicy struct {
	CaptchaThreshold int // Failed attempt threshold for captcha. Less than 0 disables it; 0 always requires it.
	BanThreshold     int // Failed attempt threshold for banning. 0 disables it.
	AttemptsWindow   time.Duration
	BanDuration      time.Duration
}

// Captcha provider interface
type CaptchaProvider interface {
	Generate() (id string, content string, answer string, err error)
	//Validate(ip, code string) bool
	Expiration() time.Duration           // Captcha expiration time, should be shorter than AttemptsWindow
	Draw(content string) (string, error) // Draw captcha
}

// Captcha metadata
type CaptchaMeta struct {
	Id        string
	Content   string
	Answer    string
	ExpiresAt time.Time
}

// IP ban record
type BanRecord struct {
	ExpiresAt time.Time
	Reason    string
}

// Login limiter
type LoginLimiter struct {
	mu          sync.Mutex
	policy      SecurityPolicy
	attempts    map[string][]time.Time //
	captchas    map[string]CaptchaMeta
	bannedIPs   map[string]BanRecord
	provider    CaptchaProvider
	cleanupStop chan struct{}
}

var defaultSecurityPolicy = SecurityPolicy{
	CaptchaThreshold: 3,
	BanThreshold:     5,
	AttemptsWindow:   5 * time.Minute,
	BanDuration:      30 * time.Minute,
}

func NewLoginLimiter(policy SecurityPolicy) *LoginLimiter {
	// Set defaults
	if policy.AttemptsWindow == 0 {
		policy.AttemptsWindow = 5 * time.Minute
	}
	if policy.BanDuration == 0 {
		policy.BanDuration = 30 * time.Minute
	}

	ll := &LoginLimiter{
		policy:      policy,
		attempts:    make(map[string][]time.Time),
		captchas:    make(map[string]CaptchaMeta),
		bannedIPs:   make(map[string]BanRecord),
		cleanupStop: make(chan struct{}),
	}
	go ll.cleanupRoutine()
	return ll
}

// Register captcha provider
func (ll *LoginLimiter) RegisterProvider(p CaptchaProvider) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	ll.provider = p
}

// isDisabled checks whether login limiting is disabled
func (ll *LoginLimiter) isDisabled() bool {
	return ll.policy.CaptchaThreshold < 0 && ll.policy.BanThreshold == 0
}

// Record failed login attempts
func (ll *LoginLimiter) RecordFailedAttempt(ip string) {
	if ll.isDisabled() {
		return
	}
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if banned, _ := ll.isBanned(ip); banned {
		return
	}

	now := time.Now()
	windowStart := now.Add(-ll.policy.AttemptsWindow)

	// Clean expired attempts
	validAttempts := ll.pruneAttempts(ip, windowStart)

	// Record a new attempt
	validAttempts = append(validAttempts, now)
	ll.attempts[ip] = validAttempts

	// Check ban conditions
	if ll.policy.BanThreshold > 0 && len(validAttempts) >= ll.policy.BanThreshold {
		ll.banIP(ip, "excessive failed attempts")
		return
	}

	return
}

// Generate captcha
func (ll *LoginLimiter) RequireCaptcha() (error, CaptchaMeta) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if ll.provider == nil {
		return errors.New("no captcha provider available"), CaptchaMeta{}
	}

	id, content, answer, err := ll.provider.Generate()
	if err != nil {
		return err, CaptchaMeta{}
	}

	// Store captcha
	ll.captchas[id] = CaptchaMeta{
		Id:        id,
		Content:   content,
		Answer:    answer,
		ExpiresAt: time.Now().Add(ll.provider.Expiration()),
	}

	return nil, ll.captchas[id]
}

// Verify captcha
func (ll *LoginLimiter) VerifyCaptcha(id, answer string) bool {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	// Find matching captcha
	if ll.provider == nil {
		return false
	}

	// Get and verify captcha
	captcha, exists := ll.captchas[id]
	if !exists {
		return false
	}

	// Clean expired captchas
	if time.Now().After(captcha.ExpiresAt) {
		delete(ll.captchas, id)
		return false
	}

	// Verify and clean state
	if answer == captcha.Answer {
		delete(ll.captchas, id)
		return true
	}

	return false
}

func (ll *LoginLimiter) DrawCaptcha(content string) (err error, str string) {
	str, err = ll.provider.Draw(content)
	return
}

// Clear record window
func (ll *LoginLimiter) RemoveAttempts(ip string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	_, exists := ll.attempts[ip]
	if exists {
		delete(ll.attempts, ip)
	}
}

// CheckSecurityStatus checks security status
func (ll *LoginLimiter) CheckSecurityStatus(ip string) (banned bool, captchaRequired bool) {
	if ll.isDisabled() {
		return
	}
	ll.mu.Lock()
	defer ll.mu.Unlock()

	// Check ban status
	if banned, _ = ll.isBanned(ip); banned {
		return
	}

	// Clean expired data
	ll.pruneAttempts(ip, time.Now().Add(-ll.policy.AttemptsWindow))

	// Check captcha requirement
	captchaRequired = len(ll.attempts[ip]) >= ll.policy.CaptchaThreshold

	return
}

// Background cleanup task
func (ll *LoginLimiter) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ll.cleanupExpired()
		case <-ll.cleanupStop:
			return
		}
	}
}

// Internal utility methods
func (ll *LoginLimiter) isBanned(ip string) (bool, BanRecord) {
	record, exists := ll.bannedIPs[ip]
	if !exists {
		return false, BanRecord{}
	}
	if time.Now().After(record.ExpiresAt) {
		delete(ll.bannedIPs, ip)
		return false, BanRecord{}
	}
	return true, record
}

func (ll *LoginLimiter) banIP(ip, reason string) {
	ll.bannedIPs[ip] = BanRecord{
		ExpiresAt: time.Now().Add(ll.policy.BanDuration),
		Reason:    reason,
	}
	delete(ll.attempts, ip)
	delete(ll.captchas, ip)
}

func (ll *LoginLimiter) pruneAttempts(ip string, cutoff time.Time) []time.Time {
	var valid []time.Time
	for _, t := range ll.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) == 0 {
		delete(ll.attempts, ip)
	} else {
		ll.attempts[ip] = valid
	}
	return valid
}

func (ll *LoginLimiter) pruneCaptchas(id string) {
	if captcha, exists := ll.captchas[id]; exists {
		if time.Now().After(captcha.ExpiresAt) {
			delete(ll.captchas, id)
		}
	}
}

func (ll *LoginLimiter) cleanupExpired() {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	now := time.Now()

	// Clean ban records
	for ip, record := range ll.bannedIPs {
		if now.After(record.ExpiresAt) {
			delete(ll.bannedIPs, ip)
		}
	}

	// Clean attempt records
	for ip := range ll.attempts {
		ll.pruneAttempts(ip, now.Add(-ll.policy.AttemptsWindow))
	}

	// Clean captchas
	for id := range ll.captchas {
		ll.pruneCaptchas(id)
	}
}
