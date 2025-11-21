package user

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var SECRET_KEY = []byte(os.Getenv("SECRET_KEY"))

func GenerateHash(password string) string {
	passwordBytes := []byte(password)
	passwordHash, _ := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(passwordHash)
}

func IsPasswordValid(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err != nil
}

func GenerateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(SECRET_KEY)
}

func VerifyToken(tokenString string) (jwt.MapClaims, error) {

	foundToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return SECRET_KEY, nil
	})

	if err != nil {
		return nil, err
	}

	if !foundToken.Valid {
		return nil, errors.New("token is invalid")
	}

	claims := foundToken.Claims.(jwt.MapClaims)

	return claims, nil

}

func Verify(c *fiber.Ctx) error {
	auth := c.Get("Authorization")

	if auth == "" {
		return c.Status(401).JSON(fiber.Map{"msg": "authentication required"})
	}

	tokenString := strings.TrimPrefix(auth, "Bearer ")

	if tokenString == "" {
		return c.Status(401).JSON(fiber.Map{"msg": "invalid key"})
	}

	fmt.Println(tokenString)

	claims, err := VerifyToken(tokenString)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"msg": "user sessions expired"})
	}

	return c.JSON(fiber.Map{"msg": "valid session ongoing", "data": claims})

}

func Login(c *fiber.Ctx, db *gorm.DB) error {

	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to verify the parsed information"})
	}

	var foundUser User
	if err := db.Where("email = ?", user.Email).First(&foundUser).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"msg": "user not found"})

	}

	if IsPasswordValid(foundUser.Password, user.Password) {
		return c.Status(401).JSON(fiber.Map{"msg": "invalid password"})
	}

	token, tokenErr := GenerateToken(foundUser)

	if tokenErr != nil {
		c.Status(400).JSON(fiber.Map{"msg": "failed to verify session"})
	}

	foundUser.Password = ""

	var data = map[string]any{
		"token": token,
		"user":  foundUser,
	}

	return c.JSON(fiber.Map{"msg": "logged in successfully", "data": data})

}

func SignUp(c *fiber.Ctx, db *gorm.DB) error {

	var user User

	// Parse incoming JSON first
	if err := c.BodyParser(&user); err != nil {
		return c.Status(401).JSON(fiber.Map{"msg": "invalid credentials"})
	}

	// Hash the user-provided password
	hashed := GenerateHash(user.Password)
	user.Password = hashed

	if user.Email == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "provide the email"})
	}

	if user.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "provide your full name"})
	}

	if user.Role == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "select your role as either company_admin or university_admin"})
	}

	var tmpUser User
	if err := db.Where("email = ?", user.Email).First(&tmpUser).Error; err == nil {
		return c.Status(402).JSON(fiber.Map{"msg": "user with email " + user.Email + " already exists"})
	}

	// Save to DB
	if err := db.Create(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid credentials submitted"})
	}

	// return c.Status(201).JSON(fiber.Map{"msg": "account created successfully"})

	return Login(c, db)

}
