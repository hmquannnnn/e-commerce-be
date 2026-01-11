# E-Commerce Microservices Platform

Hệ thống E-Commerce theo kiến trúc Microservices cho đồ án tốt nghiệp.

## 📚 Tài Liệu Thiết Kế

### 1. [Thiết Kế Tổng Quan Hệ Thống](./SYSTEM_DESIGN.md)

Tài liệu này cung cấp cái nhìn tổng quan về toàn bộ hệ thống, bao gồm:

- Kiến trúc tổng quan (Architecture Overview)
- Các công nghệ sử dụng (Technology Stack)
- Communication patterns (Sync vs Async)
- Deployment architecture
- Development workflow

**👉 Đọc tài liệu này trước để hiểu big picture của hệ thống.**

---

### 2. [User Service - Chi Tiết](./docs/USER_SERVICE_DETAIL.md)

Thiết kế chi tiết cho User Service - service cốt lõi nhất:

- Database schema (Users, Shops, Roles, Permissions)
- Authentication & Authorization (OAuth2 + JWT)
- gRPC API definitions
- Redis caching strategy
- RabbitMQ events
- Security best practices

**👉 Service quan trọng nhất, cần implement đầu tiên.**

---

### 3. [API Gateway - Chi Tiết](./docs/API_GATEWAY_DESIGN.md)

Thiết kế cho API Gateway - entry point của hệ thống:

- Request routing
- Middleware chain (Auth, Rate limiting, CORS, etc.)
- gRPC client management
- Circuit breaker pattern
- Response standardization
- Health checks & monitoring

**👉 Làm sau User Service, là cầu nối giữa client và services.**

---

### 4. [Other Services - Chi Tiết](./docs/OTHER_SERVICES_DESIGN.md)

Thiết kế cho các services còn lại:

- **Product Service**: Products, categories, inventory, search
- **Order Service**: Cart, orders, payment, shipping
- **Notification Service**: Email, SMS, push notifications
- **Chat Service**: Realtime messaging với WebSocket
- **Analytic Service**: Data analytics & reporting

**👉 Làm sau khi hoàn thành User Service và API Gateway.**

---

## 🏗️ Kiến Trúc Tổng Quan

```
┌─────────────────────────────────────────────┐
│           Client Applications               │
│     (Web, Mobile, Admin Dashboard)          │
└──────────────────┬──────────────────────────┘
                   │ HTTPS/REST
                   ▼
┌─────────────────────────────────────────────┐
│              API Gateway                    │
│  - Authentication & Authorization           │
│  - Rate Limiting                            │
│  - Request Routing                          │
└──────────────────┬──────────────────────────┘
                   │ gRPC
                   ▼
┌─────────────────────────────────────────────┐
│           Microservices Layer               │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │   User   │  │ Product  │  │  Order   │ │
│  │ Service  │  │ Service  │  │ Service  │ │
│  └──────────┘  └──────────┘  └──────────┘ │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │   Chat   │  │   Noti   │  │ Analytic │ │
│  │ Service  │  │ Service  │  │ Service  │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────┘
         │                    │
         ▼                    ▼
┌──────────────┐    ┌──────────────────┐
│   RabbitMQ   │    │      Redis       │
│   (Events)   │    │     (Cache)      │
└──────────────┘    └──────────────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│              Data Layer                     │
│  PostgreSQL × 5    MongoDB × 1              │
└─────────────────────────────────────────────┘
```

## 🛠️ Technology Stack

| Component            | Technology               | Purpose                              |
| -------------------- | ------------------------ | ------------------------------------ |
| **Backend**          | Go 1.21+                 | Service implementation               |
| **API Protocol**     | gRPC, REST               | Inter-service & client communication |
| **API Gateway**      | Go (Gin/Fiber)           | HTTP router & middleware             |
| **Databases**        | PostgreSQL 15, MongoDB 7 | Data persistence                     |
| **Cache**            | Redis 7                  | Caching & session storage            |
| **Message Queue**    | RabbitMQ 3.12            | Event-driven communication           |
| **Authentication**   | OAuth2 + JWT             | Auth & authorization                 |
| **Containerization** | Docker, Docker Compose   | Deployment                           |
| **Monitoring**       | Prometheus + Grafana     | Metrics & visualization              |

## 📦 Services Overview

### 1. User Service 👤

**Trách nhiệm**: Authentication, authorization, user management, shop management

**Tech**: Go + PostgreSQL + Redis

**Port**: 50051 (gRPC)

**Key Features**:

- JWT-based authentication
- OAuth2 social login (Google, Facebook)
- RBAC (Role-Based Access Control)
- User & Shop CRUD
- Session management

---

### 2. Product Service 📦

**Trách nhiệm**: Product catalog, inventory, search, reviews

**Tech**: Go + PostgreSQL + Redis

**Port**: 50052 (gRPC)

**Key Features**:

- Product CRUD with variants
- Category & brand management
- Full-text search
- Inventory tracking
- Product reviews & ratings

---

### 3. Order Service 🛒

**Trách nhiệm**: Shopping cart, order processing, payment

**Tech**: Go + PostgreSQL

**Port**: 50053 (gRPC)

**Key Features**:

- Shopping cart management
- Order lifecycle management
- Payment gateway integration
- Saga pattern for distributed transactions
- Refund processing

---

### 4. Notification Service 📧

**Trách nhiệm**: Multi-channel notifications

**Tech**: Go + PostgreSQL

**Port**: 50054 (gRPC)

**Key Features**:

- Email notifications (SendGrid)
- SMS notifications (Twilio)
- Push notifications (FCM)
- In-app notifications
- Template management

---

### 5. Chat Service 💬

**Trách nhiệm**: Realtime messaging between buyers & sellers

**Tech**: Go + MongoDB + WebSocket

**Port**: 50055 (gRPC), 8081 (WebSocket)

**Key Features**:

- Realtime chat via WebSocket
- Message history
- File/Image upload
- Typing indicators
- Read receipts

---

### 6. Analytic Service 📊

**Trách nhiệm**: Business analytics & reporting

**Tech**: Go + PostgreSQL (Time-series)

**Port**: 50056 (gRPC)

**Key Features**:

- Sales analytics
- Customer behavior tracking
- Product performance metrics
- Admin & shop dashboards
- Report generation

---

### 7. API Gateway 🚪

**Trách nhiệm**: Single entry point for all clients

**Tech**: Go (Gin framework)

**Port**: 8080 (HTTP/REST)

**Key Features**:

- REST to gRPC translation
- JWT validation
- Rate limiting
- CORS handling
- Request logging

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- PostgreSQL 15
- Redis 7
- RabbitMQ 3.12
- MongoDB 7

### 1. Clone Repository

```bash
git clone <repo-url>
cd e-commerce-be
```

### 2. Start Infrastructure

```bash
cd deployment
docker-compose up -d postgres-user postgres-product postgres-order mongodb-chat redis rabbitmq
```

### 3. Run Database Migrations

```bash
cd services/user
make migrate-up

cd ../product
make migrate-up

# ... repeat for other services
```

### 4. Start Services

```bash
# Terminal 1: User Service
cd services/user
go run main.go

# Terminal 2: Product Service
cd services/product
go run main.go

# Terminal 3: Order Service
cd services/order
go run main.go

# Terminal 4: API Gateway
cd api-gateway
go run main.go
```

### 5. Test API

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "full_name": "John Doe"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

---

## 📊 Database Ports

```
PostgreSQL (User):         5432
PostgreSQL (Product):      5433
PostgreSQL (Order):        5434
PostgreSQL (Notification): 5435
PostgreSQL (Analytic):     5436
MongoDB (Chat):            27017
Redis:                     6379
RabbitMQ:                  5672 (AMQP), 15672 (Management UI)
```

---

## 🔐 Security

- **Authentication**: OAuth2 + JWT (RS256)
- **Authorization**: RBAC with fine-grained permissions
- **Password Hashing**: bcrypt (cost factor 12)
- **Rate Limiting**: Configurable per endpoint
- **HTTPS**: Required in production
- **Input Validation**: All inputs validated
- **SQL Injection**: Parameterized queries only

---

## 📈 Development Phases

### ✅ Phase 1: Foundation (2 weeks)

- [x] Setup project structure
- [ ] Docker Compose environment
- [ ] User Service implementation
- [ ] API Gateway skeleton
- [ ] Basic authentication

### ⏳ Phase 2: Core Services (3 weeks)

- [ ] Product Service
- [ ] Order Service
- [ ] Notification Service
- [ ] Service integration tests

### ⏳ Phase 3: Advanced Features (2 weeks)

- [ ] Chat Service (WebSocket)
- [ ] Analytic Service
- [ ] Payment gateway integration

### ⏳ Phase 4: Polish & Deploy (1 week)

- [ ] End-to-end testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Deployment to cloud

---

## 🧪 Testing

```bash
# Unit tests
cd services/user
go test ./... -v

# Integration tests
cd services/user
go test ./tests/integration -v

# Load testing with k6
k6 run tests/load/auth-test.js
```

---

## 📝 API Documentation

API documentation available at:

- **Swagger UI**: http://localhost:8080/swagger
- **Postman Collection**: [Download](./docs/postman_collection.json)

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📞 Contact

**Author**: [Your Name]

**Email**: [your.email@example.com]

**GitHub**: [github.com/yourusername](https://github.com/yourusername)

---

## 📄 License

This project is for educational purposes (Thesis/Capstone project).

---

## 🙏 Acknowledgments

- Go gRPC team
- Docker team
- PostgreSQL community
- Redis team
- RabbitMQ team

---

## 📚 Further Reading

- [gRPC Documentation](https://grpc.io/docs/)
- [Microservices Patterns](https://microservices.io/patterns/)
- [OAuth2 Specification](https://oauth.net/2/)
- [Redis Best Practices](https://redis.io/docs/management/optimization/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html)
