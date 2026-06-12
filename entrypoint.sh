#!/bin/sh
set -e

# Load secrets recursively from /run/secrets if present
echo "=== DEBUG SECRETS ==="
if [ -d "/run/secrets" ]; then
  echo "/run/secrets is a directory"
  ls -la /run/secrets || true
  echo "Recursive structure of /run/secrets:"
  find /run/secrets || true
  
  for secret_file in $(find /run/secrets -type f 2>/dev/null); do
    if [ -f "$secret_file" ] && [ -r "$secret_file" ]; then
      echo "Loading secrets from $secret_file..."
      while IFS='=' read -r key value || [ -n "$key" ]; do
        # Skip comment lines and empty lines
        case "$key" in
          \#*|"") continue ;;
        esac
        # Remove any Windows-style carriage returns and strip surrounding quotes
        val_clean=$(echo "$value" | tr -d '\r')
        val_clean=$(echo "$val_clean" | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")
        export "$key"="$val_clean"
      done < "$secret_file"
    fi
  done
else
  echo "/run/secrets is NOT a directory"
fi
echo "======================"

echo "Starting flight.service.ma..."
exec ./main
