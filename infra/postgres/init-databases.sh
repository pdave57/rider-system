#!/bin/bash
# init-databases.sh
# Runs automatically on first PostgreSQL startup (place in /docker-entrypoint-initdb.d/)
# Creates all Rydex service databases if they don't already exist.

set -e

echo "Creating Rydex databases..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" << 'EOF'

CREATE DATABASE rydex_auth;
CREATE DATABASE rydex_orders;
CREATE DATABASE rydex_dispatch;
CREATE DATABASE rydex_payments;
CREATE DATABASE rydex_riders;

-- Grant all privileges to the postgres user on each database
GRANT ALL PRIVILEGES ON DATABASE rydex_auth     TO postgres;
GRANT ALL PRIVILEGES ON DATABASE rydex_orders   TO postgres;
GRANT ALL PRIVILEGES ON DATABASE rydex_dispatch TO postgres;
GRANT ALL PRIVILEGES ON DATABASE rydex_payments TO postgres;
GRANT ALL PRIVILEGES ON DATABASE rydex_riders   TO postgres;

EOF

echo "All databases created successfully:"
echo "  - rydex_auth     (auth-service)"
echo "  - rydex_orders   (order-service)"
echo "  - rydex_dispatch (dispatch-service)"
echo "  - rydex_payments (payment-service)"
echo "  - rydex_riders   (rider-service)"