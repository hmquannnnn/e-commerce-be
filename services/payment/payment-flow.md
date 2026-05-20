# Payment Flow — PayOS Integration

Tài liệu mô tả luồng thanh toán qua PayOS — từ UI người dùng tương tác, đến data flow phía backend.

---

## 1. Tổng quan kiến trúc

```
Browser (Next.js FE)
    │
    ▼  JWT Bearer token
API Gateway  (port 8080)
    │  inject X-User-ID / X-User-Role / X-User-Email
    ▼
Payment Service  (port 8086)
    ├── Order Service (internal HTTP, port 8085)  ← validate order
    └── PayOS API    (HTTPS, payment link + webhook)
         │
         ▼
    PostgreSQL payment DB
      ├── payments            (+ payos_order_code column)
      ├── payment_attempts
      └── payment_webhook_events
```

**Hai phương thức thanh toán:** CASH (COD) và QR_CODE (PayOS).

---

## 2. UI Flow (Frontend)

### 2.1 Trang Checkout (`CheckoutPage.tsx`)

| Lựa chọn UI | `PaymentMethod` | Xử lý |
|---|---|---|
| Tiền mặt (COD) | `CASH` | Tạo đơn hàng → chuyển về trang chi tiết đơn |
| QR Code (PayOS) | `QR_CODE` | Tạo đơn hàng → tạo payment → redirect sang PayOS |

**Bước thực hiện (QR_CODE):**

```
1. User chọn QR Code, submit checkout
2. POST /api/orders  { payment_method: "QR_CODE", items: [...] }
3. POST /api/payments {
     provider: "payos",
     payment_method: "qr_code",
     amount: Math.round(order.total_price),
     currency: "VND",
     return_url: .../payment/result?orderId=...&state=success,
     cancel_url: .../payment/result?orderId=...&state=cancel
   }
4. Nhận { checkout_url } → window.location.assign(checkout_url)
5. User thanh toán trên trang PayOS (scan QR / chuyển khoản)
6. PayOS redirect browser về return_url hoặc cancel_url
```

### 2.2 Trang kết quả (`PaymentResultPage.tsx`)

Route: `/[locale]/payment/result?orderId=<uuid>&state=<success|cancel>`

FE poll `GET /api/payments/order/:orderId` để lấy trạng thái payment thực từ DB.

---

## 3. Backend Data Flow

### 3.1 `POST /api/payments` — Tạo payment

```
Handler (bind + validate: provider=payos, payment_method=qr_code)
    │
    ▼
Service.CreatePayment
    ├─ Validate: order_id, amount > 0
    ├─ Currency: default "VND"
    │
    ▼
Order validation → GET {ORDER_SERVICE_URL}/internal/orders/{order_id}
    ├─ 404 → ErrOrderNotFound
    ├─ order.user_id ≠ caller → ErrForbidden
    ├─ order.status ≠ "PENDING" → ErrOrderNotPayable
    └─ |round(total_price) - amount| > 1 → ErrAmountMismatch
    │
    ▼
NextPayosOrderCode → SELECT nextval('payos_order_code_seq')
    │
    ▼
PayOS.CreateCheckout → PaymentRequests.Create(
    orderCode: payos_order_code (int64),
    amount, description, returnUrl, cancelUrl
)
    → trả về { checkoutUrl, paymentLinkId }
    │
    ▼
Repository.CreatePaymentAndAttempt (1 transaction)
    ├─ INSERT payments (payos_order_code = generated code)
    └─ INSERT payment_attempts (attempt_no=1, status="initiated")
    │
    ▼
Trả về PaymentResponse { id, checkout_url, status="pending", ... }
```

### 3.2 Webhook PayOS — `POST /api/payments/webhook/payos`

*Không yêu cầu auth. PayOS gọi trực tiếp (qua ngrok hoặc public URL).*

```
Handler đọc raw JSON body
    │
    ▼
Webhooks.VerifyData(body)  ← SDK verify chữ ký HMAC
    ├─ invalid → 400
    └─ ok → parse data: orderCode, paymentLinkId, reference
    │
    ▼
Map webhook code:
    code="00" AND success=true → "succeeded"
    else                       → "failed"
    │
    ▼
Resolve payment: SELECT * FROM payments WHERE payos_order_code = ?
    │
    ▼
Repository.ApplyCallback (1 transaction)
    ├─ INSERT payment_webhook_events (UNIQUE provider + event_id)
    │   duplicate → return nil (idempotent)
    ├─ UPDATE payments SET status
    └─ INSERT payment_attempts (attempt_no=999)
    │
    ▼
notifyOrderPaid → POST /internal/orders/{id}/mark-paid
    │
    ▼
200 OK { success: true }
```

### 3.3 Key Challenge: orderCode mapping

PayOS yêu cầu `orderCode` là `int64`, hệ thống dùng UUID.

**Giải pháp:**
- PostgreSQL SEQUENCE `payos_order_code_seq` (start 100001)
- Cột `payos_order_code BIGINT UNIQUE` trong bảng `payments`
- Generate khi tạo payment, lookup khi nhận webhook

---

## 4. Vòng đời trạng thái payment

```
                  ┌──────────┐
      tạo mới ──► │ pending  │
                  └────┬─────┘
                       │ PayOS xử lý
              ┌────────┼────────┐
              ▼        ▼        ▼
       requires_action  processing  (canceled ◄─ user/admin hủy)
              │        │
              └────────┘
                   │
            ┌──────┼──────┐
            ▼      ▼      ▼
        succeeded failed expired
```

---

## 5. Routes

### API Gateway (`:8080`)

| Method | Path | Auth | Forward đến |
|---|---|---|---|
| `POST` | `/api/payments/webhook/payos` | — | payment-service |
| `POST` | `/api/payments` | JWT | payment-service |
| `GET` | `/api/payments/:id` | JWT | payment-service |
| `GET` | `/api/payments/order/:orderId` | JWT | payment-service |
| `POST` | `/api/payments/:id/cancel` | JWT | payment-service |

---

## 6. Database schema

### `payments`

| Cột | Kiểu | Ghi chú |
|---|---|---|
| `id` | UUID PK | |
| `order_id` | TEXT | UUID dạng string |
| `user_id` | UUID | |
| `provider` | ENUM | `payos` |
| `payment_method` | ENUM | `qr_code` |
| `payos_order_code` | BIGINT UNIQUE | int64 cho PayOS API |
| `status` | ENUM | 9 trạng thái (xem §4) |
| `amount` | BIGINT | VND |
| `currency` | TEXT | `VND` |
| `checkout_url` | TEXT | URL redirect sang PayOS |
| `provider_payment_id` | TEXT | PayOS paymentLinkId |
| `provider_transaction_ref` | TEXT | PayOS reference |

---

## 7. Biến môi trường

| Biến | Ví dụ | Ghi chú |
|---|---|---|
| `PAYOS_CLIENT_ID` | `abc123` | PayOS Dashboard → Payment Channel |
| `PAYOS_API_KEY` | `xyz789` | PayOS Dashboard → Payment Channel |
| `PAYOS_CHECKSUM_KEY` | `secret` | Checksum key để verify webhook |
| `PAYOS_RETURN_URL` | `http://localhost:3000/en/payment/result?state=success` | FE URL sau thanh toán thành công |
| `PAYOS_CANCEL_URL` | `http://localhost:3000/en/payment/result?state=cancel` | FE URL khi user hủy |
| `ORDER_SERVICE_URL` | `http://order-service:8085` | URL nội bộ order-service |

### Webhook setup (local development)

1. Chạy `ngrok http 8080` để expose API Gateway
2. Đăng ký webhook URL tại PayOS Dashboard: `https://<ngrok-domain>/api/payments/webhook/payos`
3. Hoặc gọi `client.Webhooks.Confirm(ctx, webhookUrl)` lúc khởi động
