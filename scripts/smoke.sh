#!/bin/bash
# Smoke-тесты для http-rest-api
# Запуск: ./scripts/smoke.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_FILE=".cookies"
PASS=0
FAIL=0

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Хелпер: проверка статуса
check_status() {
    local name="$1"
    local expected="$2"
    local actual="$3"
    local extra="${4:-}"

    if [ "$actual" = "$expected" ]; then
        echo -e "${GREEN}✓${NC} $name → $actual"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}✗${NC} $name → got $actual, want $expected"
        [ -n "$extra" ] && echo "   body: $extra"
        FAIL=$((FAIL + 1))
    fi
}

# Хелпер: HTTP запрос с захватом кода и тела
# Все дополнительные аргументы (например, -b, -c, -H) передаются через "$@"
request() {
    local method="$1"
    local path="$2"
    local body="${3:-}"
    shift 3 2>/dev/null || shift $#

    local args=(-s -o /tmp/smoke_body.txt -w "%{http_code}" -X "$method" "$BASE_URL$path")

    if [ -n "$body" ]; then
        args+=(-H "Content-Type: application/json" -d "$body")
    fi

    local code
    code=$(curl "${args[@]}" "$@")
    local resp_body
    resp_body=$(cat /tmp/smoke_body.txt 2>/dev/null || echo "")
    echo "$code|$resp_body"
}

# Хелпер: извлечь код
get_code() {
    echo "$1" | cut -d'|' -f1
}

# Хелпер: извлечь тело
get_body() {
    echo "$1" | cut -d'|' -f2-
}

echo ""
echo "═══════════════════════════════════════"
echo "  Smoke-тесты http-rest-api"
echo "  Base URL: $BASE_URL"
echo "═══════════════════════════════════════"
echo ""

# ─────────────────────────────────────
# Подготовка
# ─────────────────────────────────────

echo "── Подготовка ──"
rm -f "$COOKIE_FILE"

UNIQUE_EMAIL="smoke-$(date +%s)@example.org"
echo "Тестовый email: $UNIQUE_EMAIL"
echo ""

# ─────────────────────────────────────
# 1. Health-check
# ─────────────────────────────────────

echo "── Health ──"

RESP=$(request GET /health)
check_status "GET /health" "200" "$(get_code "$RESP")" "$(get_body "$RESP")"

# ─────────────────────────────────────
# 2. Регистрация
# ─────────────────────────────────────

echo ""
echo "── Users: Create ──"

RESP=$(request POST /users "{\"email\":\"$UNIQUE_EMAIL\",\"password\":\"secret123\"}")
check_status "POST /users (ok)" "201" "$(get_code "$RESP")" "$(get_body "$RESP")"

# Извлекаем ID созданного пользователя
USER_ID=$(get_body "$RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "   Создан user ID=$USER_ID"

RESP=$(request POST /users "{not json")
check_status "POST /users (bad json)" "400" "$(get_code "$RESP")"

RESP=$(request POST /users '{"email":"","password":""}')
check_status "POST /users (empty fields)" "400" "$(get_code "$RESP")"

RESP=$(request POST /users "{\"email\":\"$UNIQUE_EMAIL\",\"password\":\"secret123\"}")
check_status "POST /users (duplicate)" "500" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 3. Get user
# ─────────────────────────────────────

echo ""
echo "── Users: Get ──"

RESP=$(request GET /users/"$USER_ID")
check_status "GET /users/{id} (ok)" "200" "$(get_code "$RESP")"

RESP=$(request GET /users/999999)
check_status "GET /users/{id} (not found)" "404" "$(get_code "$RESP")"

RESP=$(request GET /users/abc)
check_status "GET /users/{id} (invalid id)" "400" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 4. Update user
# ─────────────────────────────────────

echo ""
echo "── Users: Update ──"

UPDATED_EMAIL="updated-$(date +%s)@example.org"
RESP=$(request PUT /users/"$USER_ID" "{\"email\":\"$UPDATED_EMAIL\",\"password\":\"newsecret123\"}")
check_status "PUT /users/{id} (ok)" "200" "$(get_code "$RESP")" "$(get_body "$RESP")"

RESP=$(request PUT /users/abc '{"email":"x@y.z","password":"secret123"}')
check_status "PUT /users/{id} (invalid id)" "400" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 5. Sessions
# ─────────────────────────────────────

echo ""
echo "── Sessions ──"

# Логин с правильными данными → 200 + Set-Cookie
RESP=$(request POST /sessions "{\"email\":\"$UPDATED_EMAIL\",\"password\":\"newsecret123\"}" -c "$COOKIE_FILE")
check_status "POST /sessions (ok)" "200" "$(get_code "$RESP")" "$(get_body "$RESP")"

if [ -f "$COOKIE_FILE" ]; then
    echo -e "${GREEN}✓${NC} .cookies файл создан"
    PASS=$((PASS + 1))
else
    echo -e "${RED}✗${NC} .cookies файл НЕ создан"
    FAIL=$((FAIL + 1))
fi

# Логин с неверным паролем
RESP=$(request POST /sessions "{\"email\":\"$UPDATED_EMAIL\",\"password\":\"WRONG\"}")
check_status "POST /sessions (wrong password)" "401" "$(get_code "$RESP")"

# Логин с несуществующим email
RESP=$(request POST /sessions '{"email":"nobody@example.org","password":"secret123"}')
check_status "POST /sessions (unknown email)" "401" "$(get_code "$RESP")"

# Bad JSON
RESP=$(request POST /sessions "{not json")
check_status "POST /sessions (bad json)" "400" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 6. Private
# ─────────────────────────────────────

echo ""
echo "── Private ──"

RESP=$(request GET /private/whoami "" -b "$COOKIE_FILE")
check_status "GET /private/whoami (with cookie)" "200" "$(get_code "$RESP")" "$(get_body "$RESP")"

RESP=$(request GET /private/whoami)
check_status "GET /private/whoami (no cookie)" "401" "$(get_code "$RESP")"

RESP=$(request GET /private/whoami "" -H "Cookie: http-rest-api=fake-invalid-cookie")
check_status "GET /private/whoami (fake cookie)" "401" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 7. X-Request-Id
# ─────────────────────────────────────

echo ""
echo "── Middleware ──"

REQ_ID=$(curl -si "$BASE_URL/health" | grep -i "X-Request-Id" | tr -d '\r' | cut -d' ' -f2)
if [ -n "$REQ_ID" ]; then
    echo -e "${GREEN}✓${NC} X-Request-Id присутствует: $REQ_ID"
    PASS=$((PASS + 1))
else
    echo -e "${RED}✗${NC} X-Request-Id отсутствует"
    FAIL=$((FAIL + 1))
fi

# ─────────────────────────────────────
# 8. Delete user
# ─────────────────────────────────────

echo ""
echo "── Users: Delete ──"

RESP=$(request DELETE /users/"$USER_ID")
check_status "DELETE /users/{id} (ok)" "204" "$(get_code "$RESP")"

RESP=$(request DELETE /users/999999)
check_status "DELETE /users/{id} (unknown)" "204" "$(get_code "$RESP")"

# ─────────────────────────────────────
# 9. gRPC (опционально)
# ─────────────────────────────────────
echo ""
echo "── gRPC ──"
if command -v grpcurl &> /dev/null; then
    GRPC_RESP=$(grpcurl -plaintext localhost:50051 list 2>/dev/null)
    if [[ "$GRPC_RESP" == *"user.v1.UserService"* ]]; then
        echo -e "${GREEN}✓${NC} gRPC reflection is working"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}✗${NC} gRPC reflection failed"
        FAIL=$((FAIL + 1))
    fi
else
    echo -e "${YELLOW}⚠${NC} grpcurl not found, skipping gRPC tests"
fi

# ─────────────────────────────────────
# Итог
# ─────────────────────────────────────

echo ""
echo "═══════════════════════════════════════"
if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}  Все проверки прошли: $PASS passed${NC}"
else
    echo -e "${YELLOW}  PASS: $PASS   FAIL: $FAIL${NC}"
fi
echo "═══════════════════════════════════════"
echo ""

# Cleanup
rm -f /tmp/smoke_body.txt

exit $FAIL