package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
	"github.com/nikomkinds/SchoolSchedule/internal/utils"
)

type AuthService struct {
	authRepo  repositories.AuthRepository
	db        *sql.DB //  DB for inline join query
	jwtSecret string
}

// NewAuthService takes *sql.DB explicitly to avoid casting
func NewAuthService(authRepo repositories.AuthRepository, db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		authRepo:  authRepo,
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID.String(), user.Email, user.Role, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Getting name inline
	displayName := s.getDisplayName(ctx, user.ID)

	resp := &models.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
	resp.User.ID = user.ID.String()
	resp.User.Email = user.Email
	resp.User.Name = displayName

	return resp, nil
}

// getDisplayName executes inline-query to teachers, using *sql.DB
func (s *AuthService) getDisplayName(ctx context.Context, userID uuid.UUID) string {
	const query = `
		SELECT t.first_name, t.last_name, t.patronymic
		FROM teachers t
		WHERE t.user_id = $1
	`
	var firstName, lastName, patronymic sql.NullString
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&firstName, &lastName, &patronymic)
	if err != nil {
		if err == sql.ErrNoRows {
			// Not a teacher - fallback
			return "Пользователь"
		}
		// Log error but don’t fail login
		return "Пользователь"
	}

	name := lastName.String
	if firstName.String != "" {
		name += " " + firstName.String
	}
	if patronymic.String != "" {
		name += " " + patronymic.String
	}
	return name
}
