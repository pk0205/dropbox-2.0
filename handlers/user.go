package handlers

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pk0205/dropbox-2.0/models"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := conn.Query(context.Background(), "SELECT * FROM users")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}
		defer rows.Close()
		var users []models.User
		for rows.Next() {
			var u models.User
			err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Email, &u.Password)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Scan error: " + err.Error()})
			}
			users = append(users, u)
		}
		return c.Status(200).JSON(users)
	}
}

func SignUp(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := &models.User{}
		if err := c.BodyParser(user); err != nil {
			return c.Status(422).JSON(fiber.Map{"error": "Cannot parse JSON" + err.Error()})
		}
		user.ID = uuid.New().String()

		if (user.FirstName == "" || user.LastName == "" || user.Username == "" || user.Email == "" || user.Password == "") {
			return c.Status(400).JSON(fiber.Map{"error": "All fields are required"})
		}

		var exists bool
		err := conn.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM users WHERE username=$1)", user.Username).Scan(&exists)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}
		if exists {
			return c.Status(409).JSON(fiber.Map{"error": "Username already taken"})
		}

		err = conn.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)", user.Email).Scan(&exists)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}
		if exists {
			return c.Status(409).JSON(fiber.Map{"error": "Email already registered"})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error hashing password: " + err.Error()})
		}
		_, err = conn.Exec(context.Background(),
			"INSERT INTO users (id, firstName, lastName, username, email, password) VALUES ($1, $2, $3, $4, $5, $6)",
			user.ID, user.FirstName, user.LastName, user.Username, user.Email, hashedPassword)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.Username,
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error generating token: " + err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "AuthToken",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HTTPOnly: true,
		Secure: false,
		SameSite: "Lax",
		Path: "/",
		Domain: "",
	})
	
	// Clear password before returning user data
	user.Password = ""
	return c.Status(201).JSON(user)
}
}

func Login(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			EmailOrUsername string `json:"emailOrUsername"`
			Password        string `json:"password"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(422).JSON(fiber.Map{"error": "Cannot parse JSON: " + err.Error()})
		}

		if req.EmailOrUsername == "" || req.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Email/Username and password are required"})
		}

		// Query user by email OR username
		var user models.User
		var hashedPassword string
		err := conn.QueryRow(context.Background(),
			"SELECT * FROM users WHERE email=$1 OR username=$1",
			req.EmailOrUsername).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &hashedPassword)

		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}

		if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":    user.Username,
			"userId": user.ID,
			"exp":    time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error generating token: " + err.Error()})
		}

	c.Cookie(&fiber.Cookie{
		Name:     "AuthToken",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HTTPOnly: true,
		Secure: false,
		SameSite: "Lax",
		Path: "/",
		Domain: "",
	})

	// Password field is already empty, just return user
	return c.Status(200).JSON(user)
}
}

func GetMe(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get username from context (set by RequireAuth middleware)
		username, ok := c.Locals("userName").(string)
		if !ok || username == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// Query user by username
		var user models.User
		var hashedPassword string
		err := conn.QueryRow(context.Background(),
			"SELECT * FROM users WHERE username=$1", username).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &hashedPassword)

		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "User not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}

		// Password field is already empty, just return user
		return c.Status(200).JSON(user)
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
		})

		return c.Status(200).JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}
