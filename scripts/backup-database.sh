#!/bin/bash
set -eo pipefail

# Required environment variables
: "${DB_HOST?DB_HOST not set}"
: "${DB_USER?DB_USER not set}"
: "${DB_PASSWORD?DB_PASSWORD not set}"
: "${S3_BUCKET?S3_BUCKET not set}"
: "${AWS_ACCESS_KEY_ID?AWS_ACCESS_KEY_ID not set}"
: "${AWS_SECRET_ACCESS_KEY?AWS_SECRET_ACCESS_KEY not set}"
: "${AWS_DEFAULT_REGION?AWS_DEFAULT_REGION not set}"

BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/quantaid_$DATE.sql.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Dump the database, compress it, and upload to S3
pg_dump -h "$DB_HOST" -U "$DB_USER" -d quantaid | gzip > "$BACKUP_FILE"
aws s3 cp "$BACKUP_FILE" "s3://$S3_BUCKET/backups/"

# Clean up local backups older than 30 days
find "$BACKUP_DIR" -name "quantaid_*.sql.gz" -mtime +30 -delete

echo "Backup successful: $BACKUP_FILE"
