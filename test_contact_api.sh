#!/bin/bash

BASE_URL="http://localhost:8081/api/v1"
USER_ID="01941234-5678-7000-8000-000000000001"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║          Contact API - Повне тестування                    ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Test 1: Health Check
echo "1️⃣  Health Check"
curl -s http://localhost:8081/health | jq -r '"\(.status) - \(.timestamp)"'
echo ""

# Test 2: Створення EMAIL контакту
echo "2️⃣  Створення EMAIL контакту"
EMAIL_RESPONSE=$(curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "label": "Work Email",
    "email": "john.doe@company.com"
  }')
echo "$EMAIL_RESPONSE" | jq .
EMAIL_ID=$(echo "$EMAIL_RESPONSE" | jq -r '.data.id // empty')
echo "   📧 Email Contact ID: $EMAIL_ID"
echo ""

# Test 3: Створення PHONE контакту
echo "3️⃣  Створення PHONE контакту"
PHONE_RESPONSE=$(curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "phone",
    "label": "Mobile",
    "phone": {
      "number": "+380501234567",
      "country_code": "UA"
    }
  }')
echo "$PHONE_RESPONSE" | jq .
PHONE_ID=$(echo "$PHONE_RESPONSE" | jq -r '.data.id // empty')
echo "   📱 Phone Contact ID: $PHONE_ID"
echo ""

# Test 4: Створення ADDRESS контакту
echo "4️⃣  Створення ADDRESS контакту"
ADDRESS_RESPONSE=$(curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "address",
    "label": "Home",
    "address": {
      "street": "123 Main St",
      "city": "Kyiv",
      "country": "UA",
      "postal_code": "01001"
    }
  }')
echo "$ADDRESS_RESPONSE" | jq .
ADDRESS_ID=$(echo "$ADDRESS_RESPONSE" | jq -r '.data.id // empty')
echo "   🏠 Address Contact ID: $ADDRESS_ID"
echo ""

# Test 5: Список всіх контактів
echo "5️⃣  Список всіх контактів користувача"
curl -s "$BASE_URL/identity/contacts?user_id=$USER_ID" | jq .
echo ""

# Test 6: Отримання контакту по ID
if [ -n "$EMAIL_ID" ]; then
  echo "6️⃣  Отримання email контакту по ID"
  curl -s "$BASE_URL/identity/contacts/$EMAIL_ID" | jq .
  echo ""
fi

# Test 7: Оновлення контакту
if [ -n "$PHONE_ID" ]; then
  echo "7️⃣  Оновлення phone контакту (зміна label)"
  curl -s -X PUT "$BASE_URL/identity/contacts/$PHONE_ID" \
    -H "Content-Type: application/json" \
    -d '{
      "label": "Personal Mobile",
      "is_public": true
    }' | jq .
  echo ""
fi

# Test 8: Верифікація контакту
if [ -n "$EMAIL_ID" ]; then
  echo "8️⃣  Верифікація email контакту"
  curl -s -X PUT "$BASE_URL/identity/contacts/$EMAIL_ID/verify" | jq .
  echo ""
fi

# Test 9: Встановлення primary контакту
if [ -n "$PHONE_ID" ]; then
  echo "9️⃣  Встановлення phone як primary"
  curl -s -X PUT "$BASE_URL/identity/contacts/$PHONE_ID/primary" | jq .
  echo ""
fi

# Test 10: Перевірка в базі даних
echo "🔟 Перевірка даних в БД"
docker exec promenade_postgres psql -U system -d promenade_dev -c "
SELECT 
  id, 
  contact_type as type, 
  label, 
  is_primary, 
  is_verified, 
  is_public,
  CASE 
    WHEN email IS NOT NULL THEN email::text
    WHEN phone IS NOT NULL THEN phone::text
    ELSE 'address'
  END as value
FROM identity_contacts 
WHERE user_id = '$USER_ID' 
ORDER BY created_at;
"
echo ""

# Test 11: Видалення контакту
if [ -n "$ADDRESS_ID" ]; then
  echo "1️⃣1️⃣  Видалення address контакту"
  curl -s -X DELETE "$BASE_URL/identity/contacts/$ADDRESS_ID" | jq .
  echo ""
  
  echo "    Перевірка після видалення:"
  curl -s "$BASE_URL/identity/contacts?user_id=$USER_ID" | jq '.data | length' | xargs -I {} echo "    Залишилось контактів: {}"
  echo ""
fi

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                 ✅ Тестування завершено!                   ║"
echo "╚════════════════════════════════════════════════════════════╝"
