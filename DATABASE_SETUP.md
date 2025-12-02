# Database Setup Guide

This guide will help you set up PostgreSQL and Redis for the Organization Hub application.

## Prerequisites

- Docker (recommended) OR
- PostgreSQL 14+ and Redis 7+ installed locally

## Option 1: Using Docker (Recommended)

### 1. Create a docker-compose.yml file

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: realtime_workspace-postgres
    environment:
      POSTGRES_DB: realtime_workspace
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: realtime_workspace-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
```

### 2. Start the services

```bash
docker-compose up -d
```

### 3. Verify the services are running

```bash
# Check PostgreSQL
docker exec -it realtime_workspace-postgres psql -U postgres -d realtime_workspace -c "SELECT version();"

# Check Redis
docker exec -it realtime_workspace-redis redis-cli ping
# Should return: PONG
```

### 4. Initialize the database schema

The schema will be automatically initialized on first run using the mounted schema.sql file.
To manually run it:

```bash
docker exec -i realtime_workspace-postgres psql -U postgres -d realtime_workspace < database/schema.sql
```

## Option 2: Local Installation

### PostgreSQL Setup

#### macOS (using Homebrew)

```bash
# Install PostgreSQL
brew install postgresql@15

# Start PostgreSQL
brew services start postgresql@15

# Create database
createdb realtime_workspace

# Create user (optional, if not using default)
psql postgres -c "CREATE USER postgres WITH PASSWORD 'postgres';"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE realtime_workspace TO postgres;"

# Initialize schema
psql -U postgres -d realtime_workspace -f database/schema.sql
```

#### Linux (Ubuntu/Debian)

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql -c "CREATE DATABASE realtime_workspace;"
sudo -u postgres psql -c "CREATE USER postgres WITH PASSWORD 'postgres';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE realtime_workspace TO postgres;"

# Initialize schema
sudo -u postgres psql -d realtime_workspace -f database/schema.sql
```

### Redis Setup

#### macOS (using Homebrew)

```bash
# Install Redis
brew install redis

# Start Redis
brew services start redis

# Test connection
redis-cli ping
# Should return: PONG
```

#### Linux (Ubuntu/Debian)

```bash
# Install Redis
sudo apt update
sudo apt install redis-server

# Start Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Test connection
redis-cli ping
# Should return: PONG
```

## Configuration

Update the `config/config.go` file if your database settings differ from the defaults:

```go
PostgreSQL: PostgreSQLConfig{
    Host:         "localhost",  // Change if using remote database
    Port:         5432,
    User:         "postgres",
    Password:     "postgres",   // Change to your password
    Database:     "realtime_workspace",
    SSLMode:      "disable",    // Use "require" in production
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    MaxLifetime:  5 * time.Minute,
},
Redis: RedisConfig{
    Host:        "localhost",   // Change if using remote Redis
    Port:        6379,
    Password:    "",            // Set if Redis requires authentication
    DB:          0,
    MaxRetries:  3,
    PoolSize:    10,
    MessageTTL:  7 * 24 * time.Hour,  // Messages expire after 7 days
    MaxMessages: 1000,                 // Keep last 1000 messages per group
},
```

## Verify Setup

### Test PostgreSQL Connection

```bash
# Using psql
psql -U postgres -h localhost -d realtime_workspace -c "\dt"

# Should show tables: users, tasks
```

### Test Redis Connection

```bash
# Using redis-cli
redis-cli
127.0.0.1:6379> ping
PONG
127.0.0.1:6379> exit
```

### Run Application Health Check

Start the application and check the health endpoint:

```bash
# Start the application
go run main.go

# In another terminal
curl http://localhost:8080/health
# Should return: OK
```

## Troubleshooting

### PostgreSQL Connection Issues

**Error: "connection refused"**
- Ensure PostgreSQL is running: `pg_isadmin` (macOS) or `sudo systemctl status postgresql` (Linux)
- Check the port is correct (default: 5432)
- Verify firewall settings

**Error: "password authentication failed"**
- Check username and password in config
- Reset password: `psql postgres -c "ALTER USER postgres WITH PASSWORD 'newpassword';"`

**Error: "database does not exist"**
- Create the database: `createdb realtime_workspace`

### Redis Connection Issues

**Error: "connection refused"**
- Ensure Redis is running: `redis-cli ping`
- Check the port is correct (default: 6379)

**Error: "NOAUTH Authentication required"**
- Redis has authentication enabled
- Set the password in config or disable authentication

### Schema Initialization Issues

If the schema wasn't initialized automatically:

```bash
# For Docker
docker exec -i realtime_workspace-postgres psql -U postgres -d realtime_workspace < database/schema.sql

# For local installation
psql -U postgres -d realtime_workspace -f database/schema.sql
```

## Production Considerations

### PostgreSQL

1. **Enable SSL**: Change `SSLMode` to "require" or "verify-full"
2. **Use strong passwords**: Never use default passwords in production
3. **Configure backups**: Set up regular automated backups
4. **Tune performance**: Adjust `max_connections`, `shared_buffers`, etc.
5. **Use connection pooling**: Consider PgBouncer for better connection management

### Redis

1. **Enable authentication**: Set `requirepass` in redis.conf
2. **Configure persistence**: Enable RDB and/or AOF persistence
3. **Set memory limits**: Configure `maxmemory` and `maxmemory-policy`
4. **Use Redis Sentinel or Cluster**: For high availability
5. **Monitor memory usage**: Set up alerts for memory consumption

### Security

1. **Network isolation**: Run databases in private networks
2. **Firewall rules**: Restrict access to database ports
3. **Regular updates**: Keep PostgreSQL and Redis updated
4. **Audit logs**: Enable logging for security monitoring
5. **Encryption**: Use encryption at rest and in transit

## Useful Commands

### PostgreSQL

```bash
# Connect to database
psql -U postgres -d realtime_workspace

# List all tables
\dt

# Describe table structure
\d users
\d tasks

# View table data
SELECT * FROM users LIMIT 10;
SELECT * FROM tasks WHERE status = 'pending';

# Check database size
SELECT pg_size_pretty(pg_database_size('realtime_workspace'));

# Drop and recreate database (CAUTION: deletes all data)
DROP DATABASE realtime_workspace;
CREATE DATABASE realtime_workspace;
\c realtime_workspace
\i database/schema.sql
```

### Redis

```bash
# Connect to Redis
redis-cli

# Check memory usage
INFO memory

# List all keys (CAUTION: expensive on large datasets)
KEYS *

# Check specific group messages
ZRANGE messages:org1:group1 0 -1

# Count messages in a group
ZCARD messages:org1:group1

# Clear all data (CAUTION: deletes everything)
FLUSHALL

# Monitor real-time commands
MONITOR
```

## Data Migration

### Backup

```bash
# PostgreSQL backup
pg_dump -U postgres -d realtime_workspace > realtime_workspace_backup.sql

# Redis backup
redis-cli --rdb realtime_workspace_backup.rdb
```

### Restore

```bash
# PostgreSQL restore
psql -U postgres -d realtime_workspace < realtime_workspace_backup.sql

# Redis restore
redis-cli SHUTDOWN SAVE
cp realtime_workspace_backup.rdb /var/lib/redis/dump.rdb
redis-server
```
