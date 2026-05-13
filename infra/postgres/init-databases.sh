#!/bin/bash
# init-databases.sh
# Runs automatically on first PostgreSQL startup (place in /docker-entrypoint-initdb.d/)
# Creates all Runns service databases if they don't already exist.

set -e

echo "Creating Runns databases..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" << 'EOF'

CREATE DATABASE Runns_auth;
CREATE DATABASE Runns_orders;
CREATE DATABASE Runns_dispatch;
CREATE DATABASE Runns_payments;
CREATE DATABASE Runns_riders;

-- Grant all privileges to the postgres user on each database
GRANT ALL PRIVILEGES ON DATABASE Runns_auth     TO postgres;
GRANT ALL PRIVILEGES ON DATABASE Runns_orders   TO postgres;
GRANT ALL PRIVILEGES ON DATABASE Runns_dispatch TO postgres;
GRANT ALL PRIVILEGES ON DATABASE Runns_payments TO postgres;
GRANT ALL PRIVILEGES ON DATABASE Runns_riders   TO postgres;

EOF

echo "All databases created successfully:"
echo "  - Runns_auth     (auth-service)"
echo "  - Runns_orders   (order-service)"
echo "  - Runns_dispatch (dispatch-service)"
echo "  - Runns_payments (payment-service)"
echo "  - Runns_riders   (rider-service)"