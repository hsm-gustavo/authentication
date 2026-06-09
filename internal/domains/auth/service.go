package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/hsm-gustavo/authentication/internal/database"
	"github.com/hsm-gustavo/authentication/internal/domains/email"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
	"github.com/hsm-gustavo/authentication/shared/date"
	"github.com/hsm-gustavo/authentication/shared/hash"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	db *database.Queries
	jwtService *jwt.Service
	emailService *email.Service
}

func NewService(db *database.Queries, jwtService *jwt.Service, emailService *email.Service) *Service {
	return &Service{
		db:         db,
		jwtService: jwtService,
		emailService: emailService,
	}
}

func (s *Service) Register(ctx context.Context, dto RegisterDTO) error {
	encodedString, err := hash.HashSecret(dto.Password)
	if err != nil {
		return err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("falha ao gerar ID do usuário: %w", err)
	}

	_, err = s.db.CreateUser(ctx, database.CreateUserParams{
		ID: userID.String(),
		Email: dto.Email,
		PasswordHash: encodedString,
	})
	if err != nil {
		return fmt.Errorf("falha ao criar usuário: %w", err)
	}

	code, err := generateRecoveryCode()
	if err != nil {
		return err
	}

	s.db.CreateRecovery(ctx, database.CreateRecoveryParams{
		UserID: userID.String(),
		Email: dto.Email,
		Code: code,
		Type: "email_verification",
		ExpiresAt: time.Now().Add(15 * time.Minute).UTC(),
	})

	subject := "Bem-vindo à nossa plataforma!"
	body := fmt.Sprintf("\nOlá,\n\nObrigado por se registrar! Seu link de confirmação é: %s\n\nEste link é válido por 15 minutos.\n\nAtenciosamente,\nEquipe de Suporte", 
    "http://localhost:8080/auth/confirm-email?code="+code+"&email="+dto.Email)

	err = s.emailService.SendEmail(email.Message{
		To: dto.Email,
		Subject: subject,
		Body: body,
	})
	if err != nil {
		return fmt.Errorf("falha ao enviar email de confirmação: %w", err)
	}

	return nil
}

func (s *Service) Login(ctx context.Context, dto LoginDTO, userAgent, ipAddress string) (*LoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		fmt.Print("Erro ao buscar usuário por email: ", err)
		return nil, errors.New("credenciais inválidas")
	}

	if !user.IsVerified {
		return nil, errors.New("email não confirmado")
	}

	isValid, err := hash.VerifyArgon2Match(dto.Password, user.PasswordHash)
	if err != nil || !isValid {
		fmt.Print("Erro ao verificar senha: ", err, " - senha válida: ", isValid)
		return nil, errors.New("credenciais inválidas")
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar ID da sessão: %w", err)
	}

	secret, secretHash, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar segredo da sessão: %w", err)
	}
	ip, err := netip.ParseAddr(ipAddress)
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear endereço IP: %w", err)
	}

	_, err = s.db.CreateSession(ctx, database.CreateSessionParams{
		ID: sessionID.String(),
		UserID: user.ID,
		SecretHash: []byte(secretHash),
		IpAddress: &ip,
		UserAgent: pgtype.Text{Valid: true, String: userAgent},
	})
	
	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar token de acesso: %w", err)
	}

	return &LoginResponse{
		AccessToken: token,
		SessionSecret: secret,
		SessionID: sessionID.String(),
	}, nil
}

func (s *Service) ConfirmEmail(ctx context.Context, code string, email string) error {
	recovery, err := s.db.GetActiveRecoveryByCode(ctx, database.GetActiveRecoveryByCodeParams{
		Code: code,
		Email: email,
	})
	if err != nil {
		return errors.New("código de recuperação inválido")
	}

	if date.IsExpired(recovery.ExpiresAt) {
		return errors.New("código de recuperação expirado")
	}

	err = s.db.UpdateUserVerification(ctx, database.UpdateUserVerificationParams{
		ID: recovery.UserID,
		IsVerified: true,
	})
	if err != nil {
		return fmt.Errorf("falha ao confirmar email: %w", err)
	}

	return nil
}

func generateRecoveryCode() (string, error) {
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", fmt.Errorf("falha ao gerar código de recuperação: %w", err)
	}

	return hex.EncodeToString(codeBytes), nil
}

func generateSecret() (string, string, error) {
	sessionSecretBytes := make([]byte, 32)
	if _, err := rand.Read(sessionSecretBytes); err != nil {
    	return "", "", fmt.Errorf("falha ao gerar segredo da sessão: %w", err)
	}

	sessionSecret := base64.RawURLEncoding.EncodeToString(sessionSecretBytes)

	secretHash, err := hash.HashSecret(sessionSecret)
	if err != nil {
		return "", "", fmt.Errorf("falha ao hash do segredo da sessão: %w", err)
	}

	return sessionSecret, secretHash, nil
}