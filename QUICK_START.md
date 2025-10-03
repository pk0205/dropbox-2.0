# 🚀 Quick Start Guide

## Step-by-Step Setup

### 1. Prerequisites Check

```bash
# Check Go version (need 1.21+)
go version

# Check PostgreSQL (need 14+)
psql --version

# Check if PostgreSQL is running
pg_isready
```

### 2. Database Setup

```bash
# Create database
createdb dropbox

# Or with psql
psql -U postgres -c "CREATE DATABASE dropbox;"
```

### 3. Environment Configuration

```bash
# Copy example env file
cp .env.example .env

# Edit .env and update:
# - DATABASE_URL with your PostgreSQL credentials
# - SECRET_KEY with a random string
nano .env  # or your favorite editor
```

### 4. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify installation
go mod verify
```

### 5. Create Storage Directories

```bash
# Create necessary directories
mkdir -p storage/users storage/chunks uploads

# Set permissions
chmod 755 storage uploads
```

### 6. Start the Server

```bash
# Run the server
go run main.go

# You should see:
# Database connected ✓
# Database setup complete ✓
# Server listening on :3000 ✓
```

### 7. Test the Installation

#### Option A: Use the Web Client

1. Open `client-example.html` in your browser
2. Sign up for a new account
3. Upload a test file
4. Done!

#### Option B: Use curl

```bash
# 1. Sign up (cookies automatically saved)
curl -X POST http://localhost:3000/api/user/signup \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "firstName": "Test",
    "lastName": "User",
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Cookie is automatically saved to cookies.txt

# 2. Upload a file
curl -X POST http://localhost:3000/api/files/upload \
  -b cookies.txt \
  -F "file=@/path/to/your/file.pdf"

# 3. List files
curl http://localhost:3000/api/files \
  -b cookies.txt
```

**Note:** Use `-c cookies.txt` to save cookies and `-b cookies.txt` to send them!

## Common Issues

### Issue: "Error loading .env file"

**Solution:**

```bash
# Make sure .env exists
ls -la .env

# If not, copy from example
cp .env.example .env
```

### Issue: "Unable to connect to database"

**Solution:**

```bash
# Check PostgreSQL is running
systemctl status postgresql  # Linux
brew services list           # macOS

# Start PostgreSQL
systemctl start postgresql   # Linux
brew services start postgresql # macOS

# Verify connection string
psql postgresql://username:password@localhost:5432/dropbox
```

### Issue: "Permission denied" when uploading

**Solution:**

```bash
# Fix directory permissions
chmod -R 755 storage uploads

# If using Docker, check volume mounts
```

### Issue: Port already in use

**Solution:**

```bash
# Change port in .env
PORT=8080

# Or kill process using port 3000
lsof -ti:3000 | xargs kill
```

## Production Deployment

### Using Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
EXPOSE 3000
CMD ["./main"]
```

```bash
# Build and run
docker build -t dropbox-2.0 .
docker run -p 3000:3000 -v $(pwd)/storage:/root/storage dropbox-2.0
```

### Using systemd (Linux)

```bash
# Create service file
sudo nano /etc/systemd/system/dropbox.service
```

```ini
[Unit]
Description=Dropbox 2.0 File Storage
After=network.target postgresql.service

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/dropbox-2.0
ExecStart=/usr/local/go/bin/go run main.go
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable dropbox
sudo systemctl start dropbox
sudo systemctl status dropbox
```

### Using PM2 (Alternative)

```bash
# Install PM2
npm install -g pm2

# Start application
pm2 start main.go --interpreter="go" --interpreter-args="run"

# Save configuration
pm2 save
pm2 startup
```

## Performance Optimization

### For Development

```bash
# Use air for hot reload
go install github.com/cosmtrek/air@latest
air
```

### For Production

1. **Build optimized binary:**

```bash
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -o dropbox-server .
```

2. **Enable PostgreSQL connection pooling:**

```go
// Use pgxpool in production
import "github.com/jackc/pgx/v5/pgxpool"
```

3. **Add Redis caching:**

```bash
# Install Redis
sudo apt install redis-server

# Update code to cache file metadata
```

4. **Use reverse proxy (Nginx):**

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        client_max_body_size 100M;
    }
}
```

## Next Steps

1. ✅ Server running
2. 📱 Try the web client (`client-example.html`)
3. 📚 Read the [API Documentation](./API_DOCUMENTATION.md)
4. 🚀 Check [Performance Guide](./PERFORMANCE_GUIDE.md)
5. 🔧 Customize for your needs

## Getting Help

- 📖 Read documentation files
- 🐛 Check [Troubleshooting](#common-issues)
- 💬 Create an issue on GitHub
- 📧 Contact support

Happy file storing! 🎉
