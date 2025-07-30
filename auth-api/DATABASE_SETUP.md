# Database Setup for Planning Poker Backend

This backend now uses PostgreSQL for persistent storage, enabling real-time synchronization between multiple users.

## Quick Start with Docker (Recommended)

1. **Start PostgreSQL using Docker Compose:**
   ```bash
   docker-compose up -d postgres
   ```

2. **Update your `.env` file:**
   Make sure the database configuration in `.env` matches:
   ```env
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASSWORD=your_password_here
   DB_NAME=planning_poker
   DB_PORT=5432
   ```

3. **Start the backend server:**
   ```bash
   ./poker-server
   ```

The database tables will be automatically created when the server starts (auto-migration).

## Optional: PostgreSQL Admin UI

To view and manage your database with a GUI:

1. **Start pgAdmin:**
   ```bash
   docker-compose up -d pgadmin
   ```

2. **Access pgAdmin:**
   - Open: http://localhost:5050
   - Login: admin@example.com / admin
   - Add server connection:
     - Host: postgres (or host.docker.internal on macOS)
     - Port: 5432
     - Database: planning_poker
     - Username: postgres
     - Password: your_password_here

## Database Schema

The backend automatically creates these tables:

- **session_dbs**: Planning poker sessions
- **player_dbs**: Players in sessions  
- **story_dbs**: Stories/issues for estimation
- **chat_message_dbs**: Chat messages
- **vote_dbs**: Player votes on stories
- **jira_connection_dbs**: JIRA OAuth connections

## Features Enabled

✅ **Multi-user sessions**: Multiple users can join the same session
✅ **Real-time sync**: WebSocket + Database for live updates  
✅ **Persistent data**: Sessions survive server restarts
✅ **Vote tracking**: All votes are stored and synchronized
✅ **Chat history**: Persistent chat messages
✅ **JIRA integration**: OAuth tokens stored securely

## Alternative Setup (Local PostgreSQL)

If you prefer to install PostgreSQL locally:

1. **Install PostgreSQL:**
   ```bash
   # macOS with Homebrew
   brew install postgresql
   brew services start postgresql
   
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   ```

2. **Create database:**
   ```bash
   createdb planning_poker
   ```

3. **Update `.env` accordingly**

## Troubleshooting

- **Connection refused**: Make sure PostgreSQL is running (`docker-compose ps`)
- **Permission denied**: Check your `.env` file database credentials
- **Port conflicts**: If port 5432 is busy, change it in `docker-compose.yml`
