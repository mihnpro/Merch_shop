#!/usr/bin/env sh
# ─── User Service E2E Tests ──────────────────────────────────────────────────
# Usage:
#   GW=http://localhost:8080 sh services/user/test/user_e2e.sh
#
# Optional env vars:
#   ADMIN_LOGIN / ADMIN_PASS — credentials of an existing admin user.
#
# Requires: curl, jq

set -eu

GW="${GW:-http://localhost:8080}"
A="$GW/api/v1"
RAND=$(date +%s$$ | md5sum 2>/dev/null | head -c 8 || echo "$$")
LOGIN="e2e_$RAND"
PASS="password123"
EMAIL="$LOGIN@example.com"
PASS2="newpass456"
TOTAL=0
PASSED=0
FAILED=0

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'; NC='\033[0m'

# ─── Helpers ─────────────────────────────────────────────────────────────────

# POST/GET/PUT with body, no auth → prints "STATUS\nBODY"
call() {
  local method="$1" url="$2" data="${3:-}"
  if [ -n "$data" ]; then
    curl -s -w "\n%{http_code}" -X "$method" "$url" \
      -H 'Content-Type: application/json' -d "$data"
  else
    curl -s -w "\n%{http_code}" -X "$method" "$url"
  fi
}

# Same but with Bearer token
call_auth() {
  local method="$1" url="$2" token="$3" data="${4:-}"
  if [ -n "$data" ]; then
    curl -s -w "\n%{http_code}" -X "$method" "$url" \
      -H "Authorization: Bearer $token" \
      -H 'Content-Type: application/json' -d "$data"
  else
    curl -s -w "\n%{http_code}" -X "$method" "$url" \
      -H "Authorization: Bearer $token"
  fi
}

# Extract status code (last line)
sc() { echo "$1" | tail -1; }
# Extract body (all lines except last)
bd() { echo "$1" | sed '$d'; }

assert_code() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" expected="$2" actual="$3"
  if [ "$expected" = "$actual" ]; then
    PASSED=$((PASSED + 1))
    printf "${GREEN}✓${NC} %s → %s\n" "$desc" "$actual"
  else
    FAILED=$((FAILED + 1))
    printf "${RED}✗${NC} %s → %s (expected %s)\n" "$desc" "$actual" "$expected"
  fi
}

assert_json() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" field="$2" expected="$3" json="$4"
  actual=$(echo "$json" | jq -r "$field" 2>/dev/null || echo "PARSE_ERROR")
  if [ "$expected" = "$actual" ]; then
    PASSED=$((PASSED + 1))
    printf "${GREEN}✓${NC} %s → %s\n" "$desc" "$actual"
  else
    FAILED=$((FAILED + 1))
    printf "${RED}✗${NC} %s → %s (expected %s)\n" "$desc" "$actual" "$expected"
  fi
}

assert_json_nn() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" field="$2" json="$3"
  actual=$(echo "$json" | jq -r "$field" 2>/dev/null || echo "null")
  if [ "$actual" != "null" ] && [ "$actual" != "" ] && [ "$actual" != "PARSE_ERROR" ]; then
    PASSED=$((PASSED + 1))
    printf "${GREEN}✓${NC} %s → %s\n" "$desc" "$actual"
  else
    FAILED=$((FAILED + 1))
    printf "${RED}✗${NC} %s → %s (expected non-null)\n" "$desc" "$actual"
  fi
}

section() { printf "\n${YELLOW}── %s ──${NC}\n" "$1"; }

# ─── State ───────────────────────────────────────────────────────────────────
USER_ID=""
USER_TOKEN=""
ADMIN_ID=""
ADMIN_TOKEN=""

# ═════════════════════════════════════════════════════════════════════════════
# 1. AUTH — PUBLIC ENDPOINTS
# ═════════════════════════════════════════════════════════════════════════════
section "Auth — Register"

OUT=$(call POST "$A/auth/register" "{
  \"login\":\"$LOGIN\",\"password\":\"$PASS\",
  \"first_name\":\"Test\",\"last_name\":\"User\",\"email\":\"$EMAIL\",
  \"phone_number\":\"+1234567890\",\"patronymic\":\"Patronymic\"
}")
assert_code "Register happy path" "201" "$(sc "$OUT")"
BODY=$(bd "$OUT")
USER_ID=$(echo "$BODY" | jq -r .id)
assert_json_nn "Register returns id" '.id' "$BODY"
assert_json "Register returns login" '.login' "$LOGIN" "$BODY"
assert_json "Register returns role=user" '.role' 'user' "$BODY"
assert_json "Register returns status=active" '.status' 'active' "$BODY"

OUT=$(call POST "$A/auth/register" "{
  \"login\":\"$LOGIN\",\"password\":\"$PASS\",
  \"first_name\":\"X\",\"last_name\":\"Y\",\"email\":\"dup@example.com\"
}")
assert_code "Register duplicate login → 409" "409" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/register" '{
  "login":"bademail","password":"password123",
  "first_name":"A","last_name":"B","email":"not-an-email"
}')
assert_code "Register invalid email → 400" "400" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/register" '{
  "login":"weakpw","password":"123",
  "first_name":"A","last_name":"B","email":"w@x.com"
}')
assert_code "Register weak password → 400" "400" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/register" "{
  \"login\":\"noname\",\"password\":\"password123\",
  \"first_name\":\"\",\"last_name\":\"B\",\"email\":\"n@x.com\"
}")
assert_code "Register empty first name → 400" "400" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/register" '{')
assert_code "Register broken JSON → 400" "400" "$(sc "$OUT")"

section "Auth — Login"

OUT=$(call POST "$A/auth/login" "{\"login\":\"$LOGIN\",\"password\":\"$PASS\"}")
assert_code "Login happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json "Login returns user" '.user.login' "$LOGIN" "$BODY"

# Extract access_token from Set-Cookie
COOKIE_TMP=$(mktemp)
curl -s -X POST "$A/auth/login" -H 'Content-Type: application/json' \
  -d "{\"login\":\"$LOGIN\",\"password\":\"$PASS\"}" \
  -c "$COOKIE_TMP" -o /dev/null
USER_TOKEN=$(grep access_token "$COOKIE_TMP" | awk '{print $NF}')
rm -f "$COOKIE_TMP"

OUT=$(call POST "$A/auth/login" "{\"login\":\"$LOGIN\",\"password\":\"wrongpass\"}")
assert_code "Login wrong password → 401" "401" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/login" '{"login":"ghost_user_99999","password":"password123"}')
assert_code "Login unknown user → 401" "401" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/login" '{"login":"  ","password":""}')
assert_code "Login empty credentials → 401" "401" "$(sc "$OUT")"

section "Auth — Refresh & Logout"

OUT=$(call POST "$A/auth/refresh")
assert_code "Refresh without cookie → 401" "401" "$(sc "$OUT")"

OUT=$(call POST "$A/auth/logout")
assert_code "Logout happy path" "204" "$(sc "$OUT")"

# ═════════════════════════════════════════════════════════════════════════════
# 2. ME — AUTHENTICATED ENDPOINTS (Bearer token)
# ═════════════════════════════════════════════════════════════════════════════
section "Me — Profile"

OUT=$(call_auth GET "$A/me" "$USER_TOKEN")
assert_code "Me happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json_nn "Me returns user_id" '.user_id' "$BODY"
assert_json "Me returns role=user" '.role' 'user' "$BODY"

OUT=$(call_auth GET "$A/me" "")
assert_code "Me without token → 401" "401" "$(sc "$OUT")"

OUT=$(call_auth GET "$A/me/profile" "$USER_TOKEN")
assert_code "Get profile happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json "Profile returns login" '.login' "$LOGIN" "$BODY"

section "Me — Update Profile"

OUT=$(call_auth PUT "$A/me/profile" "$USER_TOKEN" "{
  \"first_name\":\"Updated\",\"last_name\":\"Name\",
  \"email\":\"$EMAIL\",\"phone_number\":\"+1987654321\"
}")
assert_code "Update profile happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json "Profile updated first_name" '.first_name' 'Updated' "$BODY"

OUT=$(call_auth PUT "$A/me/profile" "$USER_TOKEN" '{"first_name":"","last_name":"X","email":"a@b.com"}')
assert_code "Update profile empty name → 400" "400" "$(sc "$OUT")"

OUT=$(call_auth PUT "$A/me/profile" "$USER_TOKEN" '{"first_name":"A","last_name":"B","email":"bad"}')
assert_code "Update profile invalid email → 400" "400" "$(sc "$OUT")"

section "Me — Change Password"

OUT=$(call_auth POST "$A/me/password" "$USER_TOKEN" "{\"old_password\":\"$PASS\",\"new_password\":\"$PASS2\"}")
assert_code "Change password happy path" "204" "$(sc "$OUT")"

OUT=$(call_auth POST "$A/me/password" "$USER_TOKEN" '{"old_password":"wrongold","new_password":"newpass789"}')
assert_code "Change password wrong old → 401" "401" "$(sc "$OUT")"

OUT=$(call_auth POST "$A/me/password" "$USER_TOKEN" "{\"old_password\":\"$PASS2\",\"new_password\":\"123\"}")
assert_code "Change password weak new → 400" "400" "$(sc "$OUT")"

# Re-login with new password
COOKIE_TMP=$(mktemp)
curl -s -X POST "$A/auth/login" -H 'Content-Type: application/json' \
  -d "{\"login\":\"$LOGIN\",\"password\":\"$PASS2\"}" \
  -c "$COOKIE_TMP" -o /dev/null
USER_TOKEN=$(grep access_token "$COOKIE_TMP" | awk '{print $NF}')
rm -f "$COOKIE_TMP"

section "Me — Balance & Transactions"

OUT=$(call_auth GET "$A/me/balance" "$USER_TOKEN")
assert_code "Get balance happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json "Balance has points field" '.points' '0' "$BODY"

OUT=$(call_auth GET "$A/me/transactions" "$USER_TOKEN")
assert_code "Get transactions happy path" "200" "$(sc "$OUT")"
BODY=$(bd "$OUT")
assert_json "Transactions is array" '.transactions | type' 'array' "$BODY"

# ═════════════════════════════════════════════════════════════════════════════
# 3. ADMIN — Admin endpoints (require ADMIN_LOGIN + ADMIN_PASS)
# ═════════════════════════════════════════════════════════════════════════════
section "Admin Setup"

if [ -n "${ADMIN_LOGIN:-}" ] && [ -n "${ADMIN_PASS:-}" ]; then
  COOKIE_TMP=$(mktemp)
  curl -s -X POST "$A/auth/login" -H 'Content-Type: application/json' \
    -d "{\"login\":\"$ADMIN_LOGIN\",\"password\":\"$ADMIN_PASS\"}" \
    -c "$COOKIE_TMP" -o /dev/null
  ADMIN_TOKEN=$(grep access_token "$COOKIE_TMP" | awk '{print $NF}')
  rm -f "$COOKIE_TMP"

  OUT=$(call_auth GET "$A/me" "$ADMIN_TOKEN")
  ADMIN_ID=$(bd "$OUT" | jq -r .user_id)
  printf "  Admin ID: %s\n" "$ADMIN_ID"
else
  printf "${YELLOW}⚠ Set ADMIN_LOGIN + ADMIN_PASS to run admin tests${NC}\n"
  ADMIN_TOKEN=""
fi

section "Admin — Users"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  OUT=$(call_auth GET "$A/admin/users" "$ADMIN_TOKEN")
  assert_code "Admin list users → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "List users returns array" '.users | type' 'array' "$BODY"

  OUT=$(call_auth GET "$A/admin/users?search=$LOGIN" "$ADMIN_TOKEN")
  assert_code "Admin search users → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "Search finds our user" '.users | length | tostring' '1' "$BODY"

  OUT=$(call_auth GET "$A/admin/users?role=user" "$ADMIN_TOKEN")
  assert_code "Admin filter by role → 200" "200" "$(sc "$OUT")"

  OUT=$(call_auth GET "$A/admin/users/$USER_ID" "$ADMIN_TOKEN")
  assert_code "Admin get user by ID → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "Get user returns correct id" '.id' "$USER_ID" "$BODY"

  OUT=$(call_auth GET "$A/admin/users/00000000-0000-0000-0000-000000000000" "$ADMIN_TOKEN")
  assert_code "Admin get user not found → 404" "404" "$(sc "$OUT")"
fi

section "Admin — Grant Points"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  GRANT_OP=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || uuidgen 2>/dev/null || echo "00000000-0000-0000-0000-$(date +%s%N | head -c 13)")
  OUT=$(call_auth POST "$A/admin/users/$USER_ID/grant-points" "$ADMIN_TOKEN" "{
    \"amount\":500,\"operation_id\":\"$GRANT_OP\",\"reason\":\"test bonus\"
  }")
  assert_code "Grant points → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "Grant returns points" '.points' '500' "$BODY"

  OUT=$(call_auth POST "$A/admin/users/00000000-0000-0000-0000-000000000000/grant-points" "$ADMIN_TOKEN" "{
    \"amount\":100,\"operation_id\":\"op_nope\",\"reason\":\"nope\"
  }")
  assert_code "Grant to nonexistent → 404" "404" "$(sc "$OUT")"

  OUT=$(call_auth GET "$A/me/balance" "$USER_TOKEN")
  BODY=$(bd "$OUT")
  assert_json "User balance updated" '.points' '500' "$BODY"
fi

section "Admin — Reset Password"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  OUT=$(call_auth POST "$A/admin/users/$USER_ID/reset-password" "$ADMIN_TOKEN")
  assert_code "Reset password → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  NEW_PASS=$(echo "$BODY" | jq -r .new_password)
  assert_json "Returns 12-char password" '.new_password | length | tostring' '12' "$BODY"

  # Login with new password
  OUT=$(call POST "$A/auth/login" "{\"login\":\"$LOGIN\",\"password\":\"$NEW_PASS\"}")
  assert_code "Login with reset password → 200" "200" "$(sc "$OUT")"

  # Re-login with the new password for remaining tests
  COOKIE_TMP=$(mktemp)
  curl -s -X POST "$A/auth/login" -H 'Content-Type: application/json' \
    -d "{\"login\":\"$LOGIN\",\"password\":\"$NEW_PASS\"}" \
    -c "$COOKIE_TMP" -o /dev/null
  USER_TOKEN=$(grep access_token "$COOKIE_TMP" | awk '{print $NF}')
  rm -f "$COOKIE_TMP"
fi

section "Admin — Block / Unblock"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  OUT=$(call_auth PUT "$A/admin/users/$USER_ID/status" "$ADMIN_TOKEN" '{"blocked":true}')
  assert_code "Block user → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "User is blocked" '.status' 'blocked' "$BODY"

  OUT=$(call_auth GET "$A/me" "$USER_TOKEN")
  assert_code "Blocked user /me → 403" "403" "$(sc "$OUT")"

  OUT=$(call_auth PUT "$A/admin/users/$ADMIN_ID/status" "$ADMIN_TOKEN" '{"blocked":true}')
  assert_code "Admin block self → 403" "403" "$(sc "$OUT")"

  OUT=$(call_auth PUT "$A/admin/users/$USER_ID/status" "$ADMIN_TOKEN" '{"blocked":false}')
  assert_code "Unblock user → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "User is active" '.status' 'active' "$BODY"
fi

section "Admin — Change Role"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  OUT=$(call_auth PUT "$A/admin/users/$ADMIN_ID/role" "$ADMIN_TOKEN" '{"role":"user"}')
  assert_code "Admin change own role → 403" "403" "$(sc "$OUT")"

  RUSER="role_$RAND"
  OUT=$(call POST "$A/auth/register" "{
    \"login\":\"$RUSER\",\"password\":\"$PASS\",
    \"first_name\":\"R\",\"last_name\":\"U\",\"email\":\"$RUSER@example.com\"
  }")
  RUSER_ID=$(bd "$OUT" | jq -r .id)

  OUT=$(call_auth PUT "$A/admin/users/$RUSER_ID/role" "$ADMIN_TOKEN" '{"role":"admin"}')
  assert_code "Change role to admin → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "Role is admin" '.role' 'admin' "$BODY"

  OUT=$(call_auth PUT "$A/admin/users/$RUSER_ID/role" "$ADMIN_TOKEN" '{"role":"user"}')
  BODY=$(bd "$OUT")
  assert_json "Role back to user" '.role' 'user' "$BODY"

  OUT=$(call_auth PUT "$A/admin/users/$RUSER_ID/role" "$ADMIN_TOKEN" '{"role":"superadmin"}')
  assert_code "Invalid role → 400" "400" "$(sc "$OUT")"
fi

section "Admin — Stats & Transactions"

if [ -z "${ADMIN_TOKEN:-}" ]; then
  printf "${YELLOW}⚠ Skipped${NC}\n"
else
  OUT=$(call_auth GET "$A/admin/users/stats?period=day" "$ADMIN_TOKEN")
  assert_code "Stats day → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json_nn "Stats new_users" '.new_users' "$BODY"

  OUT=$(call_auth GET "$A/admin/users/stats?period=week" "$ADMIN_TOKEN")
  assert_code "Stats week → 200" "200" "$(sc "$OUT")"

  OUT=$(call_auth GET "$A/admin/users/stats?period=month" "$ADMIN_TOKEN")
  assert_code "Stats month → 200" "200" "$(sc "$OUT")"

  OUT=$(call_auth GET "$A/admin/users/$USER_ID/transactions" "$ADMIN_TOKEN")
  assert_code "Admin user transactions → 200" "200" "$(sc "$OUT")"
  BODY=$(bd "$OUT")
  assert_json "Transactions is array" '.transactions | type' 'array' "$BODY"
fi

# ═════════════════════════════════════════════════════════════════════════════
# 4. SUMMARY
# ═════════════════════════════════════════════════════════════════════════════
printf "\n${YELLOW}═══════════════════════════════════════════════════════${NC}\n"
printf "Total: %d  " "$TOTAL"
printf "${GREEN}Passed: %d${NC}  " "$PASSED"
if [ "$FAILED" -gt 0 ]; then
  printf "${RED}Failed: %d${NC}\n" "$FAILED"
else
  printf "Failed: 0\n"
fi
printf "${YELLOW}═══════════════════════════════════════════════════════${NC}\n"

if [ "$FAILED" -gt 0 ]; then
  exit 1
fi
