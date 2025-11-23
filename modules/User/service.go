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

func FindOne(c *fiber.Ctx, db *gorm.DB, user_id uint) (User, error) {
	var user User
	if err := db.Where("id = ?", user_id).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func CreateGroup(c *fiber.Ctx, db *gorm.DB) error {

	var group Group

	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid group data"})
	}

	group.UserID = c.Locals("user_id").(uint)

	if err := db.Create(&group).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add group"})
	}

	var members []User

	for _, u := range group.Members {
		member, _ := FindOne(c, db, u.ID)
		members = append(members, member)
	}

	group.Members = members

	return c.Status(201).JSON(fiber.Map{"data": group})
}

func AddToGroup(c *fiber.Ctx, db *gorm.DB) error {

	type Body struct {
		User  uint `json:"user"`
		Group uint `json:"group"`
	}

	var body Body

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid input"})
	}

	var group Group
	if err := db.First(&group, body.Group).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get group"})
	}

	var user User
	if err := db.First(&user, body.User).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get user"})
	}

	if err := db.Model(&group).Association("Members").Append(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add user to group" + err.Error()})
	}

	return c.SendStatus(201)
}

func RemoveFromGroup(c *fiber.Ctx, db *gorm.DB) error {

	type Body struct {
		Group uint `json:"group"`
		User  uint `json:"user"`
	}

	var body Body
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid input"})
	}

	var group Group
	if err := db.First(&group, body.Group).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to load group"})
	}

	var user User
	if err := db.First(&user, body.User).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to load user"})
	}

	if err := db.Model(&group).Association("Members").Delete(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to remove user"})
	}

	return c.SendStatus(200)

}
