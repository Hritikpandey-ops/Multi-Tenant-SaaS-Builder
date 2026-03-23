# API Testing Guide

This document provides sample requests and expected responses for the Multi-Tenant SaaS Platform API. Use these with **Postman**, **Insomnia**, or **curl**.

## 🚀 Setup Note
*   **Base URL:** `http://localhost:8080/api/v1` (API Gateway)
*   **Port:** The Gateway is running on port **8080**.
*   **Auth Prefix:** Authentication routes now include the `/auth` prefix (e.g., `/api/v1/auth/login`).

---

## 🔐 Authentication Endpoints

### 1. Register a New Tenant & Owner
**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "email": "hritik@example.com",
  "password": "SecurePassword123!",
  "first_name": "Hritik",
  "last_name": "Pandey",
  "tenant_name": "Innovitegra Solutions",
  "tenant_slug": "innovitegra"
}
```

**Expected Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4...",
  "expires_in": 900,
  "user": {
    "id": "user-uuid",
    "email": "hritik@example.com",
    "first_name": "Hritik",
    "last_name": "Pandey",
    "role": "OWNER"
  },
  "tenant": {
    "id": "tenant-uuid",
    "name": "Innovitegra Solutions",
    "slug": "innovitegra",
    "plan": "FREE"
  }
}
```

### 2. Login
**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "hritik@example.com",
  "password": "SecurePassword123!"
}
```

**Expected Response (200 OK):**
Same as Registration (returns JWT token and refresh token).

### 3. Get Current Profile (Protected)
**Endpoint:** `GET /auth/me`
**Header:** `Authorization: Bearer <your_token>`

**Expected Response (200 OK):**
```json
{
  "user": { ... },
  "tenant": { ... }
}
```

### 4. Refresh Token
**Endpoint:** `POST /auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "your-refresh-token-from-login"
}
```

---

## 👥 User Management (Protected)
*All require `Authorization: Bearer <token>`*

### 1. List All Users in Tenant
**Endpoint:** `GET /users`

### 2. Create/Invite New User
**Endpoint:** `POST /users`

**Request Body:**
```json
{
  "email": "team-member@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "MEMBER"
}
```

### 3. Delete User
**Endpoint:** `DELETE /users/:id`

---

## 🏢 Tenant Management (Protected)
*All require `Authorization: Bearer <token>`*

### 1. Get Tenant Details
**Endpoint:** `GET /tenant`

### 2. Update Tenant Settings (Owner Only)
**Endpoint:** `PUT /tenant`

**Request Body:**
```json
{
  "name": "Updated Company Name"
}
```

### 3. Get Tenant Usage
**Endpoint:** `GET /tenant/usage`

---

## 🛠️ Common Error Responses

### 401 Unauthorized
```json
{
  "code": "UNAUTHORIZED",
  "message": "Invalid credentials",
  "details": null
}
```

### 400 Validation Error
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request body",
  "details": "password must be at least 8 characters"
}
```
