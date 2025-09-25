# Goga (Go Gallery)

A modern, lightweight local web application built in Go for managing and editing photos. Upload, store, optimize, and reformat images directly in your browser with an intuitive interface.

## Features

- 📸 **Photo Upload & Management** - Drag & drop or browse to upload images
- 🎨 **Image Editing** - Basic editing tools (resize, rotate, crop)
- 🔄 **Format Conversion** - Convert between JPEG, PNG, WebP formats
- ⚡ **Image Optimization** - Automatic compression and optimization
- 📱 **Responsive Dashboard** - Clean, modern web interface
- 🗂️ **Gallery Organization** - Browse and organize your photo collection
- 🚀 **Fast & Lightweight** - Built with Go for optimal performance

## Tech Stack

- **Backend**: Go 1.21+
- **Web Framework**: Gin
- **Image Processing**: Imaging library
- **Frontend**: HTML5, Tailwind CSS, JavaScript (Vanilla)
- **Database**: SQLite (local storage)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git

#### Installing Go

**macOS:**
```bash
# Using Homebrew
brew install go

# Or download from https://golang.org/dl/
```

**Linux:**
```bash
# Download and install
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**Windows:**
- Download installer from https://golang.org/dl/
- Run the installer and follow instructions

**Verify installation:**
```bash
go version
```

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd goga
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run cmd/goga/main.go
```

4. Open your browser and navigate to `http://localhost:8080`

### Development

```bash
# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Run tests
go test ./...

# Build for production
go build -o bin/goga cmd/goga/main.go
```

## Configuration

The application can be configured via environment variables:

- `PORT` - Server port (default: 8080)
- `UPLOAD_DIR` - Upload directory (default: ./uploads)
- `DB_PATH` - Database file path (default: ./goga.db)
