# Kubernetes — UAV Store (local / Minikube)

Manifest deploy toàn bộ stack vào namespace **`uav-store`**. Thứ tự file có prefix số (`00-`, `01-`, …) để `kubectl apply -f deployment/k8s/` apply đúng thứ tự phụ thuộc cơ bản.

## Yêu cầu

- [Docker](https://docs.docker.com/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)

## Cấu trúc manifest

| File                         | Mô tả                                                          |
| ---------------------------- | -------------------------------------------------------------- |
| `00-namespace.yaml`          | Namespace `uav-store`                                          |
| `01-secrets.yaml`            | Secret (tạo từ file mẫu, **không commit** bản có giá trị thật) |
| `01-secrets.example.yaml`    | Mẫu cấu trúc Secret — copy và điền giá trị                     |
| `02-configmap.yaml`          | Config không nhạy cảm (URL nội bộ, MinIO, `ENV` cho migration) |
| `03-postgres-user.yaml`      | PVC + Postgres user DB + Service                               |
| `04-user-migrations.yaml`    | ConfigMap chứa SQL migration user                              |
| `05-minio.yaml`              | PVC + MinIO + Service (có NodePort console)                    |
| `06-user-service.yaml`       | User service + init (chờ DB, chạy migration khi `ENV=local`)   |
| `07-api-gateway.yaml`        | API Gateway + Service **NodePort** `30080`                     |
| `08-postgres-product.yaml`   | PVC + Postgres product DB + Service                            |
| `09-product-migrations.yaml` | ConfigMap migration product                                    |
| `10-file-service.yaml`       | File service (MinIO + JWT)                                     |
| `11-product-service.yaml`    | Product service                                                |
| `12-postgres-order.yaml`       | PVC + Postgres order DB + Service                              |
| `13-order-migrations.yaml`     | ConfigMap migration order                                      |
| `14-order-service.yaml`        | Order service                                                  |

Tài liệu học chi tiết hơn: [`plan.md`](./plan.md) (có thể lệch tên namespace / file so với repo — ưu tiên bảng trên).

## Chuẩn bị Secret

1. Tạo file thật (đã có trong `.gitignore`):

   ```bash
   cp deployment/k8s/01-secrets.example.yaml deployment/k8s/01-secrets.yaml
   ```

2. Điền Secret. Cách đơn giản: dùng `stringData` (plain text), ví dụ:

   ```yaml
   stringData:
     user-db-password: "postgres"
     product-db-password: "postgres"
     order-db-password: "postgres"
     minio-root-user: "minio"
     minio-root-password: "minio123"
     jwt-secret: "<chuỗi-dài-giống-nhau-trên-mọi-service>"
   ```

   Key `order-db-password` bắt buộc nếu deploy `12-postgres-order.yaml` / `14-order-service.yaml`.

   Hoặc giữ `data:` và mỗi giá trị là **base64** (một lần encode):

   ```bash
   echo -n 'postgres' | base64
   ```

3. Các key phải khớp manifest: `user-db-password`, `product-db-password`, `order-db-password`, `minio-root-user`, `minio-root-password`, `jwt-secret`.

## Build image vào Minikube (local)

Image tự build (`*:local`) phải nằm trong **Docker daemon của Minikube**, không phải Docker Desktop (trừ khi dùng driver đặc biệt).

```bash
minikube start

# Trỏ docker CLI vào Minikube (Git Bash)
eval $(minikube docker-env --shell bash)

# Từ thư mục gốc repo
docker build -f services/user/Dockerfile    -t user-service:local .
docker build -f services/file/Dockerfile     -t file-service:local .
docker build -f services/product/Dockerfile  -t product-service:local .
docker build -f api-gateway/Dockerfile       -t api-gateway:local ./api-gateway
```

Trong manifest, các service dùng Go từ root cần **build context là `.`** (đúng như lệnh trên).

## Apply toàn bộ

```bash
kubectl apply -f deployment/k8s/
```

Kiểm tra:

```bash
kubectl get pods,svc -n uav-store
kubectl logs -f deployment/api-gateway -n uav-store
```

## Truy cập API Gateway từ máy host

- **NodePort** trong manifest: `30080` (service `api-gateway`).

Trên **Windows + Minikube driver Docker**, IP `minikube ip` thường **không** vào được trực tiếp từ host. Dùng một trong hai:

```bash
# Tunnel — giữ terminal mở; sau đó thử http://127.0.0.1:30080
minikube tunnel

# Hoặc port-forward cố định
kubectl port-forward -n uav-store service/api-gateway 8080:8080
```

Hoặc:

```bash
minikube service api-gateway -n uav-store --url
```

Test health:

```bash
curl http://127.0.0.1:8080/health    # nếu dùng port-forward 8080:8080
```

## Gỡ / làm sạch namespace (cẩn thận mất PVC)

```bash
kubectl delete namespace uav-store
```

## Ghi chú production

- Đổi `image: *:local` + `imagePullPolicy: Never` thành image trên registry + `IfNotPresent` / `Always`.
- Secret không lưu trong Git; dùng CI/CD, Sealed Secrets, hoặc secret manager.
- `ENV` trong ConfigMap: `local` bật migration init; môi trường prod thường chuyển sang CI/CD Job migration.
