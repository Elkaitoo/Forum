# Forum Application

A modern web forum built with Go, featuring user authentication, post creation, and commenting functionality with a beautiful glassmorphism UI design.

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher
- Git

### Installation & Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd forum
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Run the application**
   ```bash
   go run ./cmd/main.go
   ```

4. **Open your browser**
   ```
   http://localhost:8080
   ```
   

That's it! The application will automatically create the database on first run.

## Screenshots
<img width="1237" height="816" alt="image" src="https://github.com/user-attachments/assets/5b29ac06-d56d-4a56-99e5-db9711928a8c" />
<img width="1237" height="816" alt="image" src="https://github.com/user-attachments/assets/08c42fdb-81e7-4862-bd33-060d426c4708" />
<img width="1237" height="816" alt="image" src="https://github.com/user-attachments/assets/c9231ad4-c78c-499e-8c89-d7c116642720" />
<img width="1237" height="816" alt="image" src="https://github.com/user-attachments/assets/0f9f5c0c-14fe-4bf7-a0c6-a5d9d8749575" />


## 📁 Project Structure

```
forum/
├── Dockerfile                  # Container image build
├── cmd/
│   └── main.go                 # Application entry point
├── go.mod
├── go.sum
├── internal/
│   ├── auth/                   # Authentication logic
│   │   ├── auth.go
│   │   ├── errorhandler.go
│   │   ├── middleware.go
│   │   └── sessions.go
│   ├── database/               # DB connection & queries
│   │   ├── db.go
│   │   ├── migrations.sql
│   │   ├── models.go
│   │   └── queries.go
│   ├── features/               # Business logic (posts, comments, likes)
│   │   ├── comments.go
│   │   ├── filters.go
│   │   ├── likes.go
│   │   └── posts.go
│   └── handlers/               # HTTP handlers
│       ├── auth_handlers.go
│       ├── filter_handlers.go
│       └── forum_handlers.go
├── web/
│   ├── static/
│   │   ├── css/                # Stylesheets (main.css, style.css, components...)
│   │   └── img/
│   │       └── reactions/      # reaction icons and title image
│   └── templates/              # HTML templates
│       ├── layout.html
│       ├── index.html
│       ├── create_post.html
│       ├── post_detail.html
│       ├── login.html
│       ├── register.html
│       └── error.html
├── forum.db                    # SQLite database (auto-created)
├── forum.db-wal                # SQLite WAL file (auto-created)
├── forum.db-shm                # SQLite shared memory (auto-created)
└── README.md                   # This file
```

## 🛠️ Development

### Running in Development Mode

```bash
# From project root directory
go run ./cmd/main.go
```

### Building for Production

```bash
# Build binary
go build -o forum ./cmd/main.go

# Run binary
./forum
```

### Environment Variables

The application uses these default settings (no environment variables required):

- **Port**: `8080`
- **Database**: `forum.db` (SQLite, auto-created)
- **Templates**: `web/templates/`
- **Static files**: `web/static/`

## 🎯 Features

### ✅ User Authentication
- User registration with email validation
- Secure login/logout
- Session management
- Password hashing with bcrypt

### ✅ Forum Functionality
- Create and view posts
- Comment on posts
- Category-based organization
- User-specific content

### ✅ Modern UI/UX
- **Glassmorphism design** with beautiful gradients
- **Responsive layout** for all devices
- **Smooth animations** and hover effects
- **Professional error pages** (404, 400, 500)

### ✅ Error Handling
- Custom HTTP error pages with styling
- Proper status codes (404, 400, 500)
- User-friendly error messages
- Fallback mechanisms

## 🔧 Database

The application uses **SQLite** with the following features:

- **Auto-creation**: Database is created automatically on first run
- **WAL mode**: Write-Ahead Logging for better performance
- **Automatic schema**: Tables are created automatically

### Database Files Explained
- `forum.db` - Main database file
- `forum.db-wal` - Write-Ahead Log file (auto-created)
- `forum.db-shm` - Shared memory file (auto-created)

*Note: The `.wal` and `.shm` files are automatically managed by SQLite and should not be deleted while the application is running.*

## 📊 API Endpoints

### Authentication
- `GET /login` - Login page
- `POST /login` - Login form submission
- `GET /register` - Registration page  
- `POST /register` - Registration form submission
- `GET /logout` - Logout user

### Forum
- `GET /` - Homepage with posts
- `GET /post/{id}` - View specific post with comments
- `GET /create-post` - Create post page
- `POST /create-post` - Submit new post
- `POST /comment` - Add comment to post

### Static Files
- `GET /static/` - CSS, JS, images

## 🐛 Troubleshooting

### Common Issues

**Issue: `database is locked`**
```bash
# Stop the application and restart
# Make sure only one instance is running
```

**Issue: `template not found`**
```bash
# Make sure you're running from the project root directory
cd /path/to/forum
go run ./cmd/main.go
```

**Issue: `404 errors not styled`**
- This is fixed! All invalid URLs now show beautiful error pages

**Issue: `port already in use`**
```bash
# Find and kill process using port 8080
lsof -ti:8080 | xargs kill -9
```
# Forum Application - Docker Setup

This is a simple Go forum application that has been dockerized for easy deployment.

## Building the Docker Image

```bash
docker build -t forum-app .
```

## Running the Application

### Basic Run
```bash
docker run -p 8080:8080 forum-app
```

### Run with Persistent Database
To persist the database between container restarts, mount a volume:

```bash
docker run -p 8080:8080 -v $(pwd)/data:/root/data forum-app
```

### Run in Background
```bash
docker run -d -p 8080:8080 -v $(pwd)/data:/root/data --name forum forum-app
```

## Accessing the Application

Once the container is running, access the forum at:
- **URL**: http://localhost:8080

## Managing the Container

### View running containers
```bash
docker ps
```

### Stop the container
```bash
docker stop forum
```

### Start the container again
```bash
docker start forum
```

### View logs
```bash
docker logs forum
```

### Remove the container
```bash
docker rm forum
```
