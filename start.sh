#!/bin/bash

echo "🔄 Updating dependencies..."
go mod tidy

echo "🚀 Starting Telegram bot..."
go run .
