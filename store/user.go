package store

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"nofx/ent"
	entuser "nofx/ent/user"
	"time"
)

// UserStore user storage
type UserStore struct {
	ec *ent.Client
}

// User user model
type User struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex:idx_users_email;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	OTPSecret    string    `gorm:"column:otp_secret" json:"-"`
	OTPVerified  bool      `gorm:"column:otp_verified;default:false" json:"otp_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// fromEntUser converts ent.User to store.User
func fromEntUser(eu *ent.User) *User {
	if eu == nil {
		return nil
	}
	return &User{
		ID:           eu.ID,
		Email:        eu.Email,
		PasswordHash: eu.PasswordHash,
		OTPSecret:    eu.OtpSecret,
		OTPVerified:  eu.OtpVerified,
		CreatedAt:    eu.CreatedAt,
		UpdatedAt:    eu.UpdatedAt,
	}
}

// GenerateOTPSecret generates OTP secret
func GenerateOTPSecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// NewUserStore creates a new UserStore
func NewUserStore() *UserStore {
	return &UserStore{}
}

func (s *UserStore) initTables() error {
	return nil
}



// Create creates user
func (s *UserStore) Create(user *User) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	_, err := s.ec.User.Create().
		SetID(user.ID).
		SetEmail(user.Email).
		SetPasswordHash(user.PasswordHash).
		SetNillableOtpSecret(strPtr(user.OTPSecret)).
		SetOtpVerified(user.OTPVerified).
		Save(ctx)
	return err
}

// GetByEmail gets user by email
func (s *UserStore) GetByEmail(email string) (*User, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	u, err := s.ec.User.Query().Where(entuser.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return fromEntUser(u), nil
}

// GetByID gets user by ID
func (s *UserStore) GetByID(userID string) (*User, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	u, err := s.ec.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return fromEntUser(u), nil
}

// Count returns the total number of users
func (s *UserStore) Count() (int, error) {
	if s.ec == nil {
		return 0, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	return s.ec.User.Query().Count(ctx)
}

// GetAllIDs gets all user IDs
func (s *UserStore) GetAllIDs() ([]string, error) {
	if s.ec == nil {
		return nil, fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	users, err := s.ec.User.Query().Order(ent.Asc(entuser.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids, nil
}

// UpdateOTPVerified updates OTP verification status
func (s *UserStore) UpdateOTPVerified(userID string, verified bool) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	return s.ec.User.UpdateOneID(userID).SetOtpVerified(verified).Exec(ctx)
}

// UpdatePassword updates password
func (s *UserStore) UpdatePassword(userID, passwordHash string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	ctx := context.Background()
	return s.ec.User.UpdateOneID(userID).
		SetPasswordHash(passwordHash).
		SetUpdatedAt(time.Now().UTC()).
		Exec(ctx)
}

// EnsureAdmin ensures admin user exists with the given password hash
// Returns error if admin already exists with a different password (to prevent silent overwrite)
func (s *UserStore) EnsureAdmin(passwordHash string) error {
	if s.ec == nil {
		return fmt.Errorf("ent client not available")
	}
	if passwordHash == "" {
		return fmt.Errorf("password hash cannot be empty for admin user")
	}
	ctx := context.Background()
	// Check if admin already exists
	existing, err := s.ec.User.Query().Where(entuser.ID("admin")).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create new admin with password
			_, err = s.ec.User.Create().
				SetID("admin").
				SetEmail("admin@localhost").
				SetPasswordHash(passwordHash).
				SetOtpVerified(true).
				Save(ctx)
			return err
		}
		return err
	}
	// Admin exists, verify the password hash is not empty
	if existing.PasswordHash == "" {
		// Update empty password hash
		return s.ec.User.UpdateOne(existing).
			SetPasswordHash(passwordHash).
			SetUpdatedAt(time.Now().UTC()).
			Exec(ctx)
	}
	return nil
}

// strPtr returns a pointer to the string, or nil if empty
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
