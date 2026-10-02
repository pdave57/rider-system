#!/bin/bash
# init-databases.sh
# Runs automatically on first PostgreSQL startup (place in /docker-entrypoint-initdb.d/)
# Creates all Runns service databases if they don't already exist.

set -e

echo "Creating Runns databases..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" << 'EOF'

CREATE DATABASE runns_auth;
CREATE DATABASE runns_orders;
CREATE DATABASE runns_dispatch;
CREATE DATABASE runns_payments;
CREATE DATABASE runns_riders;
CREATE DATABASE runns_shopforme;

-- Grant all privileges to the admin user on each database
GRANT ALL PRIVILEGES ON DATABASE runns_auth TO admin;
GRANT ALL PRIVILEGES ON DATABASE runns_orders TO admin;
GRANT ALL PRIVILEGES ON DATABASE runns_dispatch TO admin;
GRANT ALL PRIVILEGES ON DATABASE runns_payments TO admin;
GRANT ALL PRIVILEGES ON DATABASE runns_riders TO admin;
GRANT ALL PRIVILEGES ON DATABASE runns_shopforme TO admin;

EOF

echo "All databases created successfully:"
echo "  - runns_auth     (auth-service)"
echo "  - runns_orders   (order-service)"
echo "  - runns_dispatch (dispatch-service)"
echo "  - runns_payments (payment-service)"
echo "  - runns_riders   (rider-service)"
echo "  - runns_shopforme (shopforme-service)"