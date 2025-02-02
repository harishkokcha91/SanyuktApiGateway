# Golang Authentication Microservice

## 📌 Overview
This is a lightweight and scalable **Authentication Microservice** built using **Golang, Gin, GORM, PostgreSQL, and JWT**. It provides secure user authentication with JWT-based token management.

## 🚀 Features
- 🔑 User registration with **hashed passwords** (bcrypt)
- 🔓 User login with **JWT token generation**
- 🔒 Middleware for **protected routes**
- 🗄️ Database integration with **PostgreSQL**
- ⚡ Built using **Gin** for high performance

## 📁 Project Structure
```
auth-service/
│── main.go
│── config/
│   ├── db.go
│── models/
│   ├── user.go
│── controllers/
│   ├── auth_controller.go
│── routes/
│   ├── auth_routes.go
│── middleware/
│   ├── auth_middleware.go
│── utils/
│   ├── jwt.go
│── go.mod
│── go.sum
```

## 🛠 Installation & Setup

### 1️⃣ Install Dependencies
```sh
go mod init auth-service
go get -u github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/postgres github.com/dgrijalva/jwt-go golang.org/x/crypto/bcrypt
```

### 2️⃣ Configure Database
Edit `config/db.go` and set your PostgreSQL credentials:
```go
dsn := "host=localhost user=postgres password=yourpassword dbname=authdb port=5432 sslmode=disable"
```

### 3️⃣ Run the Microservice
```sh
go run main.go
```

## 📡 API Endpoints

### 🔹 Register User
```http
POST http://localhost:8080/register
Content-Type: application/json

{
  "username": "harish",
  "password": "password123"
}
```

### 🔹 Login User
```http
POST http://localhost:8080/login
Content-Type: application/json

{
  "username": "harish",
  "password": "password123"
}
```
_Response:_
```json
{
  "token": "your_jwt_token_here"
}
```

### 🔹 Access Protected Route
```http
GET http://localhost:8080/auth/protected
Authorization: Bearer your_jwt_token_here
```
_Response:_
```json
{
  "message": "Welcome harish"
}
```

## 📜 License
This project is licensed under the **MIT License**.

---

🔥 **Happy Coding!** 🚀

