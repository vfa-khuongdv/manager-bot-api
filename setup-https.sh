#!/bin/bash

echo "🚀 Setting up HTTPS for Manager Bot API..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
fi

# Generate SSL certificates if they don't exist
if [ ! -f certs/localhost.crt ] || [ ! -f certs/localhost.key ]; then
    echo "🔐 Generating SSL certificates..."
    ./generate-ssl.sh
else
    echo "✅ SSL certificates already exist"
fi

# Build and start services
echo "🐳 Building and starting Docker services..."
docker-compose down
docker-compose up --build -d

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📋 Services:"
echo "  🌐 HTTPS API: https://localhost"
echo "  🔗 HTTP API: http://localhost (redirects to HTTPS)"
echo "  �️  Direct Go API: http://localhost:3001"
echo "  �🗄️  MySQL: localhost:3306"
echo "  🚀 Redis: localhost:6379"
echo "  📊 phpMyAdmin: http://localhost:8080"
echo ""
echo "⚠️  Note: You may see a security warning in your browser"
echo "   because we're using self-signed certificates."
echo "   Click 'Advanced' and 'Proceed to localhost' to continue."
echo ""
echo "📝 To stop services: docker-compose down"
echo "📝 To view logs: docker-compose logs -f"