# FORMS-Bridge -- MQTT to HTTP Gateway

**FORMS-Bridge** là một giải pháp Gateway trung gian hiệu năng cao được
phát triển bằng **Go 1.23**.\
Ứng dụng đóng vai trò cầu nối giữa thiết bị IoT (ESP32, Arduino, Sensor)
và hệ thống Backend thông qua mô hình:

**MQTT → Bridge xử lý → HTTP REST API**

Hệ thống được thiết kế hướng đến hiệu năng, ổn định và khả năng mở rộng
cho các dự án IoT quy mô lớn.

## 📋 Tính Năng Nổi Bật

### 🚀 Hiệu suất cao

-   Tận dụng **goroutine** và **channel** để xử lý song song.
-   Dễ dàng đạt **hàng nghìn messages/giây** với độ trễ thấp.

### 🪶 Tối ưu tài nguyên

-   Binary nhỏ gọn.
-   RAM tiêu thụ chỉ **10--20MB**.
-   Phù hợp chạy trên **Raspberry Pi**, **OpenWRT**, hoặc Docker
    container.

### 🔁 Tự động khôi phục kết nối

-   Tự động reconnect MQTT Broker khi mất mạng.

### 🧭 Định tuyến động (Dynamic Routing)

-   Cấu hình qua file **YAML**, không cần sửa code.
-   Hỗ trợ wildcard trong topic.

### 📜 Logging chi tiết

-   Định dạng `text` hoặc `json`.
-   Hỗ trợ các mức: `debug`, `info`, `warn`, `error`.

## 🏗 Kiến Trúc Hệ Thống

``` mermaid
graph LR
    A[Device/Sensor] -- MQTT Publish --> B[MQTT Broker]
    B -- Subscribe --> C[PMMNM Bridge]
    C -- Parse & Validate --> C
    C -- HTTP POST --> D[Backend API]
    D -- Response --> C
```

## 🛠 Công Nghệ Sử Dụng
```
  Thành phần    Công nghệ
  ------------- --------------------------
  Ngôn ngữ      Go 1.23
  MQTT Client   eclipse/paho.mqtt.golang
  Logging       logrus
  Config        yaml.v3
```
## 🚀 Hướng Dẫn Cài Đặt (Quick Start)

### 1. Yêu cầu hệ thống

-   Go **1.23** trở lên
-   MQTT Broker (Mosquitto, EMQX,...)
-   Backend REST API endpoint

### 2. Cài đặt & Chạy

#### **Cách 1: Chạy trực tiếp**

``` bash
git clone https://github.com/Caxtiq/pmmnm-bridge.git
cd pmmnm-bridge
go mod download
go run main.go
```

#### **Cách 2: Build binary**

``` bash
go build -o bridge
./bridge
```

## ⚙️ Cấu Hình

Sao chép file mẫu:

``` bash
cp config.example.yaml config.yaml
```

### 📄 Ví dụ `config.yaml`

``` yaml
api:
  endpoint: "http://localhost:3001/api/sensor-data"
  timeout: 10s

mqtt:
  broker: "tcp://localhost:1883"
  client_id: "pmmnm-bridge-01"
  username: ""
  password: ""
  qos: 1
  clean_session: true

topics:
  - mqtt_topic: "sensors/+/data"
    sensor_id_from_payload: true
    description: "Kênh dữ liệu tổng hợp từ các cảm biến"

logging:
  level: "info"
  format: "json"
```

## 📡 Định Dạng Payload MQTT

### Ví dụ topic:

    sensors/water/data

### Payload JSON hợp lệ:

``` json
{
  "sensorId": "sensor-water-01",
  "value": 125.5,
  "timestamp": 1704556789000
}
```

> 📝 *Nếu không có timestamp, Bridge sẽ tự bổ sung timestamp hiện tại.*

## 🐳 Triển Khai Bằng Docker

### Build Image:

``` bash
docker build -t pmmnm-bridge .
```

### Chạy container:

``` bash
docker run -d   --name mqtt-bridge   -v $(pwd)/config.yaml:/root/config.yaml   --restart unless-stopped   pmmnm-bridge
```

## 🤝 Đóng Góp (Contributing)

Mọi đóng góp để cải thiện dự án đều được hoan nghênh.\
Tạo **Pull Request** hoặc mở **Issue** trên GitHub để thảo luận.

## 📜 Giấy Phép

Dự án được phân phối theo **Apache License 2.0**.

**Maintainer:** PKA-OpenLD\
**Year:** 2025
