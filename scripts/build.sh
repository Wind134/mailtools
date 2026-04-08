#!/bin/bash
set -e

cd /home/ping/mailtools

echo "=== Building MailTools ==="

echo "Building web server..."
go build -o ~/go/bin/mailtools-web ./cmd/web

echo "Building CLI..."
go build -o ~/go/bin/ping-mail ./cmd/send

echo "Building createuser tool..."
go build -o ~/go/bin/createuser ./cmd/createuser

echo "=== Installing frontend dependencies ==="
cd frontend
npm install

echo ""
echo "=== Build complete ==="
echo ""
echo "Next steps:"
echo "1. Create database: mysql -u root -p < scripts/init.sql"
echo "2. Copy config: cp scripts/config.toml ~/.config/mailtools/config.toml"
echo "3. Edit ~/.config/mailtools/config.toml with your DB password"
echo "4. Copy systemd service: sudo cp scripts/mailtools-web.service /etc/systemd/system/"
echo "5. Reload systemd: sudo systemctl daemon-reload"
echo "6. Start service: sudo systemctl start mailtools-web"
echo "7. Create admin user: ~/go/bin/createuser <your_password>"
echo "8. Start frontend: cd frontend && npm run dev"
echo ""
echo "Or run manually:"
echo "  Backend: ~/go/bin/mailtools-web -config ~/.config/mailtools/config.toml"
echo "  Frontend: cd frontend && npm run dev"
