package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
	"time"

	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const resetTokenTTL = time.Hour

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func generateResetToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	plain := hex.EncodeToString(bytes)
	hashed := hashToken(plain)
	return plain, hashed, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ForgotPassword handles password reset email requests.
func ForgotPassword(c *fiber.Ctx, db *gorm.DB) error {
	var req forgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Invalid request payload"})
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Email is required"})
	}

	var foundUser user.User
	if err := db.Where("email = ?", email).First(&foundUser).Error; err != nil {
		log.Printf("forgot password: no user found for %s", email)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "No account found for this email"})
	}

	plain, hashed, err := generateResetToken()
	if err != nil {
		log.Printf("forgot password: failed to generate token for %s: %v", email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "Unable to process request right now. Please try again later."})
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// Invalidate previous tokens for this user.
		if err := tx.Where("user_id = ?", foundUser.ID).Delete(&PasswordResetToken{}).Error; err != nil {
			return err
		}

		token := PasswordResetToken{
			UserID:    foundUser.ID,
			TokenHash: hashed,
			ExpiresAt: time.Now().Add(resetTokenTTL),
		}

		if err := tx.Create(&token).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Printf("forgot password: failed to store reset token for %s: %v", email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "Unable to process request right now. Please try again later."})
	}

	if err := SendPasswordResetEmail(foundUser.Email, plain, foundUser.Name); err != nil {
		log.Printf("forgot password: failed to send email to %s: %v", email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "Unable to send reset email at the moment. Please try again later."})
	}

	return c.JSON(fiber.Map{"msg": "If an account exists for this email, a password reset link has been sent"})
}

// ResetPassword validates a reset token and updates the user password.
func ResetPassword(c *fiber.Ctx, db *gorm.DB) error {
	var req resetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Invalid request payload"})
	}

	token := strings.TrimSpace(req.Token)
	password := strings.TrimSpace(req.Password)

	if token == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Reset token and password are required"})
	}

	if len(password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Password must be at least 8 characters"})
	}

	hashed := hashToken(token)

	var resetToken PasswordResetToken
	if err := db.Where("token_hash = ?", hashed).First(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "This reset link is invalid or has already been used."})
	}

	if resetToken.UsedAt != nil || time.Now().After(resetToken.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "This reset link has expired. Please request a new one."})
	}

	var targetUser user.User
	if err := db.First(&targetUser, resetToken.UserID).Error; err != nil {
		log.Printf("reset password: user not found for token %d: %v", resetToken.ID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "This reset link is invalid. Please request a new one."})
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&targetUser).Update("password", user.GenerateHash(password)).Error; err != nil {
			return err
		}

		now := time.Now()
		if err := tx.Model(&resetToken).Updates(map[string]interface{}{
			"used_at":    &now,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Printf("reset password: failed transaction for user %d: %v", targetUser.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "Unable to reset password right now. Please try again later."})
	}

	return c.JSON(fiber.Map{"msg": "Your password has been updated successfully."})
}
