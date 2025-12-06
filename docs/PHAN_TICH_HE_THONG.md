# PHÂN TÍCH HỆ THỐNG - MQTT BRIDGE (IoT Gateway)

## 📋 TỔNG QUAN CHUNG

Đây là một **MQTT-to-HTTP Bridge** được viết bằng **Go**, đóng vai trò gateway kết nối giữa các thiết bị IoT (ESP32, Arduino) sử dụng giao thức MQTT với hệ thống backend HTTP/REST API. Bridge nhận dữ liệu từ MQTT broker, parse và forward đến API endpoint để xử lý tự động.

### Tên Dự Án
- **Module**: github.com/Caxtiq/pmmnm-bridge
- **Language**: Go 1.23
- **License**: Apache License 2.0

### Mục Đích
- Kết nối thiết bị IoT (ESP32) với backend qua MQTT
- Forward sensor data từ MQTT sang HTTP API
- Hỗ trợ nhiều sensors với dynamic routing
- Auto-reconnect và error handling
- Logging chi tiết cho monitoring

---

## 🏗️ CẤU TRÚC DỰ ÁN

```
bridge/
├── main.go                  # Entry point, initialization
├── config.go                # Configuration management
├── mqtt.go                  # MQTT client & message handling
├── api.go                   # HTTP API client
│
├── config.yaml              # Active configuration file
├── config.example.yaml      # Configuration template
│
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
│
└── .git                     # Git metadata (submodule?)
```

**File Sizes:**
- main.go: 1.9 KB (76 lines)
- mqtt.go: 5.4 KB (196 lines)
- api.go: 3.8 KB (140 lines)
- config.go: 2.2 KB (90 lines)

**Total**: ~13 KB của Go code, rất nhẹ và hiệu quả

---

## 💻 CÔNG NGHỆ SỬ DỤNG

### Core Language
- **Go 1.23** - Compiled, high-performance language
  - Fast startup time
  - Low memory footprint
  - Built-in concurrency
  - Cross-platform compilation

### Dependencies

#### MQTT Client
- **github.com/eclipse/paho.mqtt.golang v1.4.3**
  - Official Eclipse Paho MQTT client
  - Support MQTT 3.1.1
  - Auto-reconnect
  - QoS levels 0, 1, 2
  - TLS/SSL support

#### Logging
- **github.com/sirupsen/logrus v1.9.3**
  - Structured logging
  - Multiple output formats (text, JSON)
  - Log levels: debug, info, warn, error
  - Fields support

#### Configuration
- **gopkg.in/yaml.v3 v3.0.1**
  - YAML parsing
  - Schema validation
  - Type conversion

#### Indirect Dependencies
- gorilla/websocket v1.5.0 (for WebSocket in paho.mqtt)
- golang.org/x/net v0.8.0 (networking)
- golang.org/x/sync v0.1.0 (synchronization primitives)
- golang.org/x/sys v0.6.0 (OS-specific)

---

## 🗄️ CẤU TRÚC DỮ LIỆU

### 1. Configuration Structure (config.go)

```go
type Config struct {
    API     APIConfig     // API endpoint settings
    MQTT    MQTTConfig    // MQTT broker settings
    Topics  []TopicMap    // Topic mappings
    Logging LoggingConfig // Logging settings
}

type APIConfig struct {
    Endpoint string        // API URL
    Timeout  time.Duration // HTTP timeout
}

type MQTTConfig struct {
    Broker       string // MQTT broker URL
    ClientID     string // Client identifier
    Username     string // Auth username (optional)
    Password     string // Auth password (optional)
    QoS          byte   // Quality of Service (0, 1, 2)
    CleanSession bool   // Clean session flag
}

type TopicMap struct {
    MQTTTopic           string // MQTT topic to subscribe
    SensorIDFromPayload bool   // Extract sensor ID from JSON
    Description         string // Human-readable description
}

type LoggingConfig struct {
    Level  string // debug, info, warn, error
    Format string // text or json
}
```

### 2. MQTT Message Format (mqtt.go)

**SensorPayload từ ESP32:**
```json
{
  "sensorId": "sensor-water-level-01",
  "value": 125.5,
  "timestamp": 1704556789000
}
```

**Fields:**
- `sensorId`: Unique sensor identifier (required in payload-based mode)
- `value`: Sensor reading (float64)
- `timestamp`: Unix timestamp in milliseconds (optional, auto-filled if missing)

### 3. API Request Format (api.go)

**Request to Backend (POST /api/sensor-data):**
```json
{
  "sensorId": "sensor-water-level-01",
  "value": 125.5,
  "timestamp": 1704556789000
}
```

**Success Response (201 Created):**
```json
{
  "success": true,
  "dataId": "data-123456",
  "thresholdExceeded": true,
  "automation": {
    "rulesChecked": 3,
    "rulesTriggered": 2,
    "zonesCreated": [
      "auto-zone-1234567890",
      "auto-zone-1234567891"
    ],
    "message": "Automation triggered successfully"
  }
}
```

**Warning Response (202 Accepted):**
```json
{
  "warning": "Sensor not found, data saved but not processed"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Invalid request format: sensorId is required"
}
```

### 4. Configuration File (config.yaml)

```yaml
# API Configuration
api:
  endpoint: "http://localhost:3001/api/sensor-data"
  timeout: 10s

# MQTT Broker Configuration
mqtt:
  broker: "tcp://localhost:1883"
  client_id: "pmmnm-bridge"
  username: ""
  password: ""
  qos: 1
  clean_session: true

# Topic Mappings
topics:
  - mqtt_topic: "sensors/data"
    sensor_id_from_payload: true
    description: "All sensors - ID from payload"

# Logging Configuration
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json or text
```

---

## 🔄 LUỒNG HOẠT ĐỘNG

### A. Startup Flow (main.go)

```
1. Parse command line flags
   └─ --config <path> (default: config.yaml)

2. Initialize logger (logrus)
   ├─ Set formatter (Text/JSON)
   ├─ Set timestamp format
   └─ Configure log level

3. Load configuration
   ├─ Read YAML file
   ├─ Parse with yaml.v3
   ├─ Apply defaults
   └─ Validate required fields

4. Create components
   ├─ API Client (with HTTP timeout)
   └─ MQTT Bridge (with config, API client, logger)

5. Connect to MQTT Broker
   ├─ Create Paho client with options
   ├─ Set handlers (message, connect, disconnect)
   ├─ Connect with auto-reconnect
   └─ Subscribe to configured topics

6. Wait for interrupt
   ├─ Listen for Ctrl+C (SIGINT)
   └─ Listen for SIGTERM

7. Graceful shutdown
   ├─ Disconnect MQTT (250ms grace)
   └─ Log shutdown message
```

### B. MQTT Message Processing Flow (mqtt.go)

```
[ESP32 Device]
     │
     │ Publish to MQTT topic
     ▼
[MQTT Broker]
     │
     │ Forward to subscribers
     ▼
[Bridge - messageHandler]
     │
     ├─1. Receive MQTT message
     │   └─ Extract topic and payload
     │
     ├─2. Parse JSON payload
     │   ├─ Unmarshal to SensorPayload struct
     │   └─ Handle parse errors
     │
     ├─3. Determine sensor ID
     │   ├─ Match topic to configured mappings
     │   ├─ Check wildcard patterns (+, #)
     │   └─ Extract sensorId from payload
     │
     ├─4. Add timestamp if missing
     │   └─ time.Now().UnixMilli()
     │
     ├─5. Forward to API
     │   ├─ Call apiClient.SendSensorData()
     │   ├─ Handle HTTP request
     │   └─ Process response
     │
     └─6. Log results
         ├─ Success: Info log
         └─ Error: Error log with details
```

### C. API Communication Flow (api.go)

```
[Bridge]
     │
     ├─1. Prepare request
     │   ├─ Create SensorDataRequest struct
     │   ├─ Marshal to JSON
     │   └─ Set Content-Type header
     │
     ├─2. Send HTTP POST
     │   ├─ Use http.Client with timeout
     │   ├─ POST to configured endpoint
     │   └─ Wait for response
     │
     ├─3. Handle response
     │   ├─ Read response body
     │   └─ Check status code
     │
     ├─4. Process by status
     │   ├─ 200/201: Success
     │   │   ├─ Parse SensorDataResponse
     │   │   ├─ Log automation results
     │   │   └─ Check thresholdExceeded
     │   │
     │   ├─ 202: Accepted (warning)
     │   │   └─ Log warning message
     │   │
     │   ├─ 400: Bad Request
     │   │   └─ Return error
     │   │
     │   └─ Other: Unexpected
     │       └─ Return error with body
     │
     └─5. Return result
         ├─ nil on success
         └─ error on failure
```

### D. Connection Management

**Auto-Reconnect:**
```
[Connected]
     │
     ▼
[Connection Lost] ──────┐
     │                  │
     ▼                  │
[connectionLostHandler] │
     │                  │
     ├─ Log warning     │
     └─ Paho auto-retry │
         │              │
         ▼              │
    [Reconnecting]      │
         │              │
         ├─ Backoff     │
         └─ Max 1 min   │
             │          │
             ▼          │
        [Connected] ────┘
             │
             ▼
    [onConnectHandler]
             │
             └─ Re-subscribe all topics
```

### E. Topic Matching

**Wildcard Support:**
- `+`: Single-level wildcard
  - `sensors/+/data` matches:
    - `sensors/water/data` ✓
    - `sensors/temperature/data` ✓
    - `sensors/water/level/data` ✗

- `#`: Multi-level wildcard
  - `sensors/#` matches:
    - `sensors/water/data` ✓
    - `sensors/water/level/data` ✓
    - `sensors/temp/room1/data` ✓

**Matching Algorithm:**
```go
func matchTopic(pattern, topic string) bool {
    patternParts := strings.Split(pattern, "/")
    topicParts := strings.Split(topic, "/")
    
    for i, part := range patternParts {
        if part == "#" {
            return true  // Match rest
        }
        if i >= len(topicParts) {
            return false  // Pattern longer
        }
        if part == "+" {
            continue  // Wildcard
        }
        if part != topicParts[i] {
            return false  // Mismatch
        }
    }
    
    return len(patternParts) == len(topicParts)
}
```

---

## 🔧 CÁC FILE CHI TIẾT

### 1. main.go (76 lines)

**Responsibilities:**
- Application entry point
- Command-line argument parsing
- Component initialization
- Graceful shutdown handling

**Key Functions:**
```go
func main() {
    // 1. Parse flags
    configFile := flag.String("config", "config.yaml", "...")
    flag.Parse()
    
    // 2. Setup logger
    logger := logrus.New()
    
    // 3. Load and validate config
    config, err := LoadConfig(*configFile)
    config.Validate()
    
    // 4. Create components
    apiClient := NewAPIClient(...)
    bridge := NewMQTTBridge(...)
    
    // 5. Connect
    bridge.Connect()
    
    // 6. Wait for signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    // 7. Cleanup
    bridge.Disconnect()
}
```

**Command-line Usage:**
```bash
# Default config
./bridge

# Custom config
./bridge --config production.yaml

# Help
./bridge --help
```

### 2. config.go (90 lines)

**Responsibilities:**
- Configuration structure definitions
- YAML parsing
- Default value setting
- Configuration validation

**Key Functions:**
```go
func LoadConfig(filename string) (*Config, error) {
    // Read YAML file
    data, err := os.ReadFile(filename)
    
    // Parse YAML
    var config Config
    yaml.Unmarshal(data, &config)
    
    // Set defaults
    if config.API.Timeout == 0 {
        config.API.Timeout = 10 * time.Second
    }
    if config.MQTT.ClientID == "" {
        config.MQTT.ClientID = "pmmnm-bridge"
    }
    
    return &config, nil
}

func (c *Config) Validate() error {
    // Check required fields
    if c.API.Endpoint == "" {
        return fmt.Errorf("api.endpoint is required")
    }
    if c.MQTT.Broker == "" {
        return fmt.Errorf("mqtt.broker is required")
    }
    if len(c.Topics) == 0 {
        return fmt.Errorf("at least one topic mapping required")
    }
    return nil
}
```

**Validation Rules:**
- API endpoint: Required, non-empty
- MQTT broker: Required, non-empty
- Topics: At least 1 mapping
- Timeout: Positive duration (default 10s)

### 3. mqtt.go (196 lines)

**Responsibilities:**
- MQTT client management
- Message handling and parsing
- Topic matching with wildcards
- Connection lifecycle

**Key Structures:**
```go
type MQTTBridge struct {
    config    *Config
    client    mqtt.Client      // Paho client
    apiClient *APIClient
    logger    *logrus.Logger
    topicMap  map[string]TopicMap  // Fast lookup
}

type SensorPayload struct {
    SensorID  string  `json:"sensorId,omitempty"`
    Value     float64 `json:"value"`
    Timestamp int64   `json:"timestamp,omitempty"`
}
```

**Key Functions:**
```go
func (b *MQTTBridge) Connect() error {
    // Create Paho client options
    opts := mqtt.NewClientOptions()
    opts.AddBroker(b.config.MQTT.Broker)
    opts.SetClientID(b.config.MQTT.ClientID)
    opts.SetCleanSession(b.config.MQTT.CleanSession)
    
    // Set auth if configured
    if b.config.MQTT.Username != "" {
        opts.SetUsername(b.config.MQTT.Username)
        opts.SetPassword(b.config.MQTT.Password)
    }
    
    // Set handlers
    opts.SetDefaultPublishHandler(b.messageHandler)
    opts.SetConnectionLostHandler(b.connectionLostHandler)
    opts.SetOnConnectHandler(b.onConnectHandler)
    
    // Enable auto-reconnect
    opts.SetAutoReconnect(true)
    opts.SetMaxReconnectInterval(1 * time.Minute)
    
    // Connect
    b.client = mqtt.NewClient(opts)
    token := b.client.Connect()
    token.Wait()
    
    return token.Error()
}

func (b *MQTTBridge) messageHandler(client mqtt.Client, msg mqtt.Message) {
    // 1. Parse JSON payload
    var payload SensorPayload
    json.Unmarshal(msg.Payload(), &payload)
    
    // 2. Determine sensor ID
    sensorID := b.getSensorID(msg.Topic(), &payload)
    
    // 3. Add timestamp if missing
    if payload.Timestamp == 0 {
        payload.Timestamp = time.Now().UnixMilli()
    }
    
    // 4. Forward to API
    b.apiClient.SendSensorData(sensorID, payload.Value, payload.Timestamp)
}
```

**MQTT QoS Levels:**
- QoS 0: At most once (fire and forget)
- QoS 1: At least once (acknowledged delivery) - **Recommended**
- QoS 2: Exactly once (4-way handshake)

### 4. api.go (140 lines)

**Responsibilities:**
- HTTP client management
- API request formatting
- Response parsing
- Error handling

**Key Structures:**
```go
type APIClient struct {
    endpoint   string
    httpClient *http.Client  // With timeout
    logger     *logrus.Logger
}

type SensorDataRequest struct {
    SensorID  string  `json:"sensorId"`
    Value     float64 `json:"value"`
    Timestamp int64   `json:"timestamp,omitempty"`
}

type SensorDataResponse struct {
    Success           bool   `json:"success"`
    DataID            string `json:"dataId"`
    ThresholdExceeded bool   `json:"thresholdExceeded"`
    Automation        struct {
        RulesChecked   int      `json:"rulesChecked"`
        RulesTriggered int      `json:"rulesTriggered"`
        ZonesCreated   []string `json:"zonesCreated"`
        Message        string   `json:"message"`
    } `json:"automation"`
}
```

**Key Functions:**
```go
func (c *APIClient) SendSensorData(sensorID string, value float64, timestamp int64) error {
    // 1. Create request
    req := SensorDataRequest{
        SensorID:  sensorID,
        Value:     value,
        Timestamp: timestamp,
    }
    body, _ := json.Marshal(req)
    
    // 2. Send POST request
    httpReq, _ := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    resp, _ := c.httpClient.Do(httpReq)
    defer resp.Body.Close()
    
    // 3. Handle response by status code
    switch resp.StatusCode {
    case http.StatusCreated, http.StatusOK:
        // Parse success response
        var apiResp SensorDataResponse
        json.Unmarshal(respBody, &apiResp)
        
        // Log automation results
        if apiResp.ThresholdExceeded {
            c.logger.Warn("Threshold exceeded - automation triggered")
        }
        return nil
        
    case http.StatusAccepted:
        // Warning (sensor not found)
        return nil
        
    case http.StatusBadRequest:
        // Bad request
        return fmt.Errorf("bad request: %s", errResp.Error)
        
    default:
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
}
```

---

## 🚀 HƯỚNG DẪN SỬ DỤNG

### Setup & Build

```bash
cd bridge/

# Install dependencies
go mod download

# Build for current platform
go build -o bridge

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o bridge-linux-amd64
GOOS=windows GOARCH=amd64 go build -o bridge-windows-amd64.exe
GOOS=linux GOARCH=arm GOARM=7 go build -o bridge-linux-arm7  # Raspberry Pi

# Build with optimization
go build -ldflags="-s -w" -o bridge  # Smaller binary
```

### Configuration

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit configuration
nano config.yaml
```

**Key settings to change:**
```yaml
api:
  endpoint: "http://YOUR_SERVER_IP:3001/api/sensor-data"

mqtt:
  broker: "tcp://YOUR_MQTT_BROKER:1883"
  username: "your_username"  # If required
  password: "your_password"  # If required

topics:
  - mqtt_topic: "sensors/data"
    sensor_id_from_payload: true
```

### Running

```bash
# Run with default config
./bridge

# Run with custom config
./bridge --config production.yaml

# Run in background (Linux)
nohup ./bridge > bridge.log 2>&1 &

# Run as systemd service (Linux)
sudo systemctl start pmmnm-bridge
```

### Testing

**Test MQTT Connection:**
```bash
# Subscribe to topic (using mosquitto_sub)
mosquitto_sub -h localhost -t "sensors/data" -v

# Publish test message
mosquitto_pub -h localhost -t "sensors/data" \
  -m '{"sensorId":"sensor-test-001","value":25.5}'
```

**Test API Connection:**
```bash
# Check if API is reachable
curl -X POST http://localhost:3001/api/sensor-data \
  -H "Content-Type: application/json" \
  -d '{"sensorId":"test","value":10}'
```

**Check Logs:**
```bash
# Real-time logs
tail -f bridge.log

# Filter errors
grep "ERROR" bridge.log

# Count successful forwards
grep "Successfully forwarded" bridge.log | wc -l
```

---

## 📊 DEPLOYMENT PATTERNS

### 1. Single Raspberry Pi Setup

```
[ESP32 Devices]
     │
     ├─ WiFi
     │
     ▼
[Raspberry Pi]
├─ Mosquitto MQTT Broker
├─ Bridge (this app)
└─ (Optional) Backend API
     │
     ├─ Internet
     │
     ▼
[Cloud Backend]
```

**Raspberry Pi Setup:**
```bash
# Install Mosquitto
sudo apt install mosquitto mosquitto-clients

# Configure Mosquitto
sudo nano /etc/mosquitto/mosquitto.conf
# Add:
listener 1883
allow_anonymous true

# Restart
sudo systemctl restart mosquitto

# Install bridge
scp bridge pi@raspberrypi:/home/pi/
ssh pi@raspberrypi
chmod +x bridge
./bridge
```

### 2. Cloud MQTT Broker

```
[ESP32 Devices]
     │
     ├─ Internet
     │
     ▼
[Cloud MQTT Broker]
(HiveMQ, CloudMQTT, AWS IoT)
     │
     ▼
[Bridge on VPS/Docker]
     │
     ▼
[Backend API]
```

**config.yaml:**
```yaml
mqtt:
  broker: "tcp://mqtt.example.com:1883"
  username: "device_user"
  password: "secure_password"
  qos: 1
  clean_session: false  # Persistent session
```

### 3. Docker Deployment

**Dockerfile:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o bridge

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bridge .
COPY config.yaml .
CMD ["./bridge"]
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  mosquitto:
    image: eclipse-mosquitto:latest
    ports:
      - "1883:1883"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf

  bridge:
    build: .
    depends_on:
      - mosquitto
    environment:
      - TZ=Asia/Ho_Chi_Minh
    restart: unless-stopped
```

**Run:**
```bash
docker-compose up -d
docker-compose logs -f bridge
```

### 4. Systemd Service (Production)

**/etc/systemd/system/pmmnm-bridge.service:**
```ini
[Unit]
Description=PMMNM MQTT Bridge
After=network.target mosquitto.service

[Service]
Type=simple
User=pmmnm
WorkingDirectory=/opt/pmmnm-bridge
ExecStart=/opt/pmmnm-bridge/bridge --config config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Setup:**
```bash
# Create user
sudo useradd -r -s /bin/false pmmnm

# Copy files
sudo mkdir -p /opt/pmmnm-bridge
sudo cp bridge config.yaml /opt/pmmnm-bridge/
sudo chown -R pmmnm:pmmnm /opt/pmmnm-bridge

# Enable service
sudo systemctl daemon-reload
sudo systemctl enable pmmnm-bridge
sudo systemctl start pmmnm-bridge

# Check status
sudo systemctl status pmmnm-bridge
sudo journalctl -u pmmnm-bridge -f
```

---

## 🔍 MONITORING & DEBUGGING

### Log Analysis

**JSON Logs (recommended for production):**
```json
{
  "level": "info",
  "msg": "Successfully forwarded sensor data to API",
  "sensor_id": "sensor-water-level-01",
  "time": "2025-01-06T14:30:52+07:00",
  "topic": "sensors/data",
  "value": 125.5
}
```

**Text Logs (human-readable):**
```
INFO[2025-01-06T14:30:52+07:00] Successfully forwarded sensor data to API
  sensor_id=sensor-water-level-01 topic=sensors/data value=125.5
```

**Useful Commands:**
```bash
# Count messages per sensor
jq -r '.sensor_id' bridge.log | sort | uniq -c

# Find errors
jq 'select(.level=="error")' bridge.log

# Average value by sensor
jq -r 'select(.sensor_id=="sensor-001") | .value' bridge.log | awk '{sum+=$1} END {print sum/NR}'

# Monitor in real-time
tail -f bridge.log | jq 'select(.threshold_exceeded==true)'
```

### Health Checks

**Monitoring Script:**
```bash
#!/bin/bash
# check_bridge.sh

# Check if process is running
if ! pgrep -f "bridge" > /dev/null; then
    echo "ERROR: Bridge is not running"
    exit 1
fi

# Check recent activity (last 5 minutes)
recent=$(find bridge.log -mmin -5 -type f)
if [ -z "$recent" ]; then
    echo "WARNING: No recent log activity"
    exit 1
fi

# Check error rate
errors=$(grep -c "ERROR" bridge.log)
if [ "$errors" -gt 100 ]; then
    echo "WARNING: High error count: $errors"
    exit 1
fi

echo "OK: Bridge is healthy"
exit 0
```

### Performance Metrics

**Memory Usage:**
```bash
# Go binary is very lightweight
ps aux | grep bridge
# Typical: 10-20 MB RAM
```

**Message Throughput:**
```bash
# Count messages per second
watch -n 1 'tail -100 bridge.log | grep "forwarded" | wc -l'
```

**Latency:**
- MQTT receive: < 10ms
- JSON parse: < 1ms
- API call: depends on network (typically 10-100ms)
- Total: < 150ms typical

---

## 🔐 BẢO MẬT & BEST PRACTICES

### MQTT Security

**1. Authentication:**
```yaml
mqtt:
  username: "bridge_user"
  password: "${MQTT_PASSWORD}"  # Use env vars
```

**2. TLS/SSL:**
```yaml
mqtt:
  broker: "ssl://mqtt.example.com:8883"
```

**Go code for TLS:**
```go
import "crypto/tls"

tlsConfig := &tls.Config{
    ClientAuth: tls.NoClientCert,
    ClientCAs:  nil,
    InsecureSkipVerify: false,  // Verify certificates
}
opts.SetTLSConfig(tlsConfig)
```

**3. Access Control (MQTT broker):**
```conf
# mosquitto.conf
allow_anonymous false
password_file /etc/mosquitto/passwd
acl_file /etc/mosquitto/acl

# acl file
user bridge_user
topic read sensors/#
topic write $SYS/#
```

### API Security

**1. HTTPS:**
```yaml
api:
  endpoint: "https://api.example.com/sensor-data"
```

**2. API Keys:**
```go
// Add to api.go
httpReq.Header.Set("Authorization", "Bearer " + config.API.APIKey)
httpReq.Header.Set("X-API-Key", config.API.APIKey)
```

**3. Rate Limiting:**
```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 20)  // 10 req/s, burst 20

func (c *APIClient) SendSensorData(...) error {
    if err := c.limiter.Wait(context.Background()); err != nil {
        return err
    }
    // ... rest of function
}
```

### Production Hardening

**1. Input Validation:**
```go
func (p *SensorPayload) Validate() error {
    if p.SensorID == "" {
        return fmt.Errorf("sensorId is required")
    }
    if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
        return fmt.Errorf("invalid value")
    }
    if p.Timestamp < 0 {
        return fmt.Errorf("invalid timestamp")
    }
    return nil
}
```

**2. Error Recovery:**
```go
// In main.go
for {
    if err := runBridge(); err != nil {
        logger.WithError(err).Error("Bridge crashed, restarting...")
        time.Sleep(5 * time.Second)
    }
}
```

**3. Resource Limits:**
```go
// Limit concurrent API calls
semaphore := make(chan struct{}, 10)  // Max 10 concurrent

func (c *APIClient) SendSensorData(...) error {
    semaphore <- struct{}{}        // Acquire
    defer func() { <-semaphore }() // Release
    
    // ... rest of function
}
```

**4. Graceful Degradation:**
```go
// Buffer messages when API is down
type MessageBuffer struct {
    queue []SensorPayload
    mu    sync.Mutex
}

func (b *MessageBuffer) Add(p SensorPayload) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.queue = append(b.queue, p)
}

// Retry later
```

---

## 📊 LUỒNG DỮ LIỆU TỔNG QUÁT

```
┌─────────────────┐
│  ESP32 Device   │
│  (Sensor)       │
└────────┬────────┘
         │
         │ WiFi
         │
         ▼
┌─────────────────┐
│  MQTT Broker    │
│  (Mosquitto)    │
└────────┬────────┘
         │
         │ TCP/MQTT
         │
         ▼
┌─────────────────┐
│  Bridge         │
│  ├─ MQTT Client │
│  ├─ Parser      │
│  └─ API Client  │
└────────┬────────┘
         │
         │ HTTP/JSON
         │
         ▼
┌─────────────────┐
│  Backend API    │
│  (/api/sensor-  │
│   data)         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Rule Engine    │
│  - Check rules  │
│  - Auto zones   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  WebSocket      │
│  Broadcast      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Frontend Map   │
│  Display zones  │
└─────────────────┘
```

---

## 🎯 USE CASES & SCENARIOS

### 1. Water Level Monitoring

**ESP32 Code:**
```cpp
#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>

WiFiClient espClient;
PubSubClient client(espClient);

void publishSensorData(float waterLevel) {
    StaticJsonDocument<200> doc;
    doc["sensorId"] = "sensor-water-level-bridge-01";
    doc["value"] = waterLevel;
    doc["timestamp"] = millis();
    
    char buffer[256];
    serializeJson(doc, buffer);
    
    client.publish("sensors/data", buffer);
}

void loop() {
    float level = readWaterLevel();
    publishSensorData(level);
    delay(30000);  // Every 30 seconds
}
```

**Bridge Config:**
```yaml
topics:
  - mqtt_topic: "sensors/data"
    sensor_id_from_payload: true
    description: "Water level sensors"
```

### 2. Multiple Sensor Types

**Topics Structure:**
```
sensors/
├── water/data          # Water level sensors
├── temperature/data    # Temperature sensors
├── traffic/data        # Traffic cameras
└── weather/data        # Weather stations
```

**Bridge Config:**
```yaml
topics:
  - mqtt_topic: "sensors/water/data"
    sensor_id_from_payload: true
    description: "Water level sensors"
    
  - mqtt_topic: "sensors/temperature/data"
    sensor_id_from_payload: true
    description: "Temperature sensors"
    
  - mqtt_topic: "sensors/#"  # Wildcard for all
    sensor_id_from_payload: true
    description: "All sensors fallback"
```

### 3. High-Frequency Data

**For high-frequency sensors:**
```yaml
mqtt:
  qos: 0  # Faster, less reliable
  
logging:
  level: "warn"  # Less verbose
```

**Batching in ESP32:**
```cpp
void loop() {
    // Collect 10 readings
    for (int i = 0; i < 10; i++) {
        readings[i] = readSensor();
        delay(1000);
    }
    
    // Send average
    float avg = calculateAverage(readings, 10);
    publishSensorData(avg);
}
```

---


## 📖 TÀI LIỆU THAM KHẢO

### MQTT
- Eclipse Paho Go: https://github.com/eclipse/paho.mqtt.golang
- MQTT Specification: https://mqtt.org/mqtt-specification/
- Mosquitto Broker: https://mosquitto.org/

### Go Libraries
- Logrus: https://github.com/sirupsen/logrus
- YAML v3: https://github.com/go-yaml/yaml
- Go Modules: https://go.dev/doc/modules

### ESP32 Integration
- PubSubClient: https://github.com/knolleary/pubsubclient
- ArduinoJson: https://arduinojson.org/

---

## 👥 TEAM & LICENSE

- **Module**: github.com/Caxtiq/pmmnm-bridge
- **Maintainer**: PKA-OpenLD
- **License**: Apache License 2.0
- **Language**: Go 1.23
- **Year**: 2025

---

_Tài liệu này được tạo tự động dựa trên phân tích source code._
_Cập nhật lần cuối: 2025-12-05_
