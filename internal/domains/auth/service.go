package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hsm-gustavo/authentication/internal/database"
	"github.com/hsm-gustavo/authentication/internal/domains/email"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
	"github.com/hsm-gustavo/authentication/shared/date"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
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
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("falha ao gerar salt: %w", err)
	}

	hash := argon2.IDKey([]byte(dto.Password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	b64Salt  := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash  := base64.RawStdEncoding.EncodeToString(hash)
	encodedString := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)

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

func (s *Service) Login(ctx context.Context, dto LoginDTO) (*LoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		fmt.Print("Erro ao buscar usuário por email: ", err)
		return nil, errors.New("credenciais inválidas")
	}

	if !user.IsVerified {
		return nil, errors.New("email não confirmado")
	}

	isValid, err := verifyArgon2Match(dto.Password, user.PasswordHash)
	if err != nil || !isValid {
		fmt.Print("Erro ao verificar senha: ", err, " - senha válida: ", isValid)
		return nil, errors.New("credenciais inválidas")
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar token de acesso: %w", err)
	}

	return &LoginResponse{
		AccessToken: token,
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

func verifyArgon2Match(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("formato de hash inválido")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, errors.New("versão do argon2 incompatível")
	}

	var memory, time, threads uint32
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, errors.New("parâmetros do argon2 inválidos")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func generateRecoveryCode() (string, error) {
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", fmt.Errorf("falha ao gerar código de recuperação: %w", err)
	}

	return hex.EncodeToString(codeBytes), nil
}