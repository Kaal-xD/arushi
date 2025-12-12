#!/bin/bash

echo "🔧 Installing Go dependencies..."

# Init go module if not exists
if [ ! -f "go.mod" ]; then
    echo "📦 go.mod not found — creating..."
    go mod init tg-bot
fi

# Install all required modules
echo "📦 Installing telebot..."
go get gopkg.in/telebot.v3

echo "📦 Installing gopsutil..."
go get github.com/shirou/gopsutil/v3

# Clean + update modules
echo "🧹 Running go mod tidy..."
go mod tidy

echo "✅ All dependencies installed successfully!"
