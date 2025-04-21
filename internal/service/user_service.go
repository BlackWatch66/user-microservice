package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/blackwatch66/user-microservice/config"
	"github.com/blackwatch66/user-microservice/internal/auth"
	"github.com/blackwatch66/user-microservice/internal/model"
	redisClient "github.com/blackwatch66/user-microservice/internal/redis"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// UserService defines the user service interface
type UserService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error) // Returns JWT
	GetUserProfile(ctx context.Context, userID uint) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userID uint, firstName, lastName string) (*model.User, error)
	GetUserAddresses(ctx context.Context, userID uint) ([]model.Address, error)
	AddUserAddress(ctx context.Context, userID uint, addr model.Address) (*model.Address, error)
	UpdateUserAddress(ctx context.Context, userID, addrID uint, addr model.Address) (*model.Address, error)
	DeleteUserAddress(ctx context.Context, userID, addrID uint) error
    ValidateToken(ctx context.Context, tokenString string) (*auth.Claims, error)
}

// userServiceImpl implements the UserService interface
type userServiceImpl struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

// NewUserService creates a new UserService instance
func NewUserService(db *gorm.DB, rdb *redis.Client, cfg *config.Config) UserService {
	return &userServiceImpl{db: db, rdb: rdb, cfg: cfg}
}

// Register handles user registration logic
func (s *userServiceImpl) Register(ctx context.Context, email, password string) (*model.User, error) {
	// Check if email already exists
	var existingUser model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error checking email: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := model.User{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Emit UserCreated event (optional)
	log.Printf("User registered: ID=%d, Email=%s\n", user.ID, user.Email)

	return &user, nil
}

// Login handles user login logic
func (s *userServiceImpl) Login(ctx context.Context, email, password string) (string, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", fmt.Errorf("database error finding user: %w", err)
	}

	// Check password
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT
	tokenString, err := auth.GenerateJWT(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Store JWT identifier in Redis (per document requirements)
    // Simple example using UserID as key, more complex strategies might be needed
    // E.g. storing JTI (JWT ID) or the token itself with expiry time matching JWT
	err = s.rdb.Set(redisClient.Ctx, fmt.Sprintf("jwt:%d", user.ID), tokenString, s.cfg.JWTExpiry).Err()
    if err != nil {
        // Even if Redis fails, we may allow login success but log the error
        log.Printf("WARN: Failed to store JWT identifier in Redis for user %d: %v\n", user.ID, err)
    }

	return tokenString, nil
}

// GetUserProfile retrieves user profile
func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	// Preload Addresses to get associated address information
	if err := s.db.WithContext(ctx).Preload("Addresses").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error finding user: %w", err)
	}
	// Clean sensitive information
    user.PasswordHash = ""
	return &user, nil
}

// UpdateUserProfile updates user profile
func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID uint, firstName, lastName string) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error finding user: %w", err)
	}

	user.FirstName = firstName
	user.LastName = lastName

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
    // Clean sensitive information
    user.PasswordHash = ""
	return &user, nil
}

// GetUserAddresses gets user address list
func (s *userServiceImpl) GetUserAddresses(ctx context.Context, userID uint) ([]model.Address, error) {
	var addresses []model.Address
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, fmt.Errorf("database error finding addresses: %w", err)
	}
	return addresses, nil
}

// AddUserAddress adds a user address
func (s *userServiceImpl) AddUserAddress(ctx context.Context, userID uint, addr model.Address) (*model.Address, error) {
	// Ensure address belongs to the user
	addr.UserID = userID
	// Clean possible client-provided ID and timestamps
    addr.ID = 0
    addr.CreatedAt = time.Time{}
    addr.UpdatedAt = time.Time{}

	if err := s.db.WithContext(ctx).Create(&addr).Error; err != nil {
		return nil, fmt.Errorf("failed to add address: %w", err)
	}
	return &addr, nil
}

// UpdateUserAddress updates a user address
func (s *userServiceImpl) UpdateUserAddress(ctx context.Context, userID, addrID uint, updatedAddr model.Address) (*model.Address, error) {
	var existingAddr model.Address
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", addrID, userID).First(&existingAddr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("address not found or does not belong to user")
		}
		return nil, fmt.Errorf("database error finding address: %w", err)
	}

	// Update fields
	existingAddr.Street = updatedAddr.Street
	existingAddr.City = updatedAddr.City
	existingAddr.State = updatedAddr.State
	existingAddr.PostalCode = updatedAddr.PostalCode
	existingAddr.Country = updatedAddr.Country
	existingAddr.IsDefault = updatedAddr.IsDefault

	if err := s.db.WithContext(ctx).Save(&existingAddr).Error; err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}
	return &existingAddr, nil
}

// DeleteUserAddress deletes a user address
func (s *userServiceImpl) DeleteUserAddress(ctx context.Context, userID, addrID uint) error {
    // First check if the address exists and belongs to the user
    var addr model.Address
    if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", addrID, userID).First(&addr).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Address doesn't exist or doesn't belong to user, can be treated as delete success or return specific error
            return errors.New("address not found or does not belong to user")
        }
        return fmt.Errorf("database error finding address before delete: %w", err)
    }

	// Execute delete
	if err := s.db.WithContext(ctx).Delete(&model.Address{}, addrID).Error; err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}
	return nil
}

// ValidateToken validates JWT Token (for gRPC use)
func (s *userServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
    claims, err := auth.ValidateJWT(tokenString, s.cfg.JWTSecret)
    if err != nil {
        return nil, fmt.Errorf("token validation failed: %w", err)
    }

    // Optional: Check if token identifier exists in Redis (for quick revocation)
    // _, err = s.rdb.Get(redisClient.Ctx, fmt.Sprintf("jwt:%d", claims.UserID)).Result()
    // if err == redis.Nil {
    //     return nil, errors.New("token has been revoked or expired from cache")
    // } else if err != nil {
    //     log.Printf("WARN: Redis error during token validation for user %d: %v\n", claims.UserID, err)
    //     // Depending on policy, may reject token due to Redis error
    // }

    return claims, nil
} 