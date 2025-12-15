#!/usr/bin/env bash

# Преобразование имени в snake_case (совместимо с macOS и Linux)
to_snake_case() {
    local input="$1"
    # Добавляем _ перед заглавными буквами, конвертируем в lowercase, убираем начальный _
    echo "$input" | sed -E 's/([A-Z])/_\1/g' | tr '[:upper:]' '[:lower:]' | sed 's/^_//'
}

# Преобразование в CamelCase
to_camel_case() {
    echo "$1" | sed -E 's/_([a-z])/\U\1/g' | sed -E 's/^([a-z])/\U\1/'
}

# Преобразование в lowercase
to_lower() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}

# Преобразование первой буквы в lowercase
to_lower_first() {
    local input="$1"
    local first="${input:0:1}"
    local rest="${input:1}"
    echo "$(echo "$first" | tr '[:upper:]' '[:lower:]')$rest"
}

# Преобразование в множественное число
to_plural() {
    local word="$1"
    
    # Правила для множественного числа
    if [[ $word =~ (s|x|z|ch|sh)$ ]]; then
        # words ending in s, x, z, ch, sh -> add 'es'
        echo "${word}es"
    elif [[ $word =~ [^aeiou]y$ ]]; then
        # words ending in consonant + y -> replace y with 'ies'
        echo "${word%y}ies"
    elif [[ $word =~ f$ ]]; then
        # words ending in f -> replace with 'ves'
        echo "${word%f}ves"
    elif [[ $word =~ fe$ ]]; then
        # words ending in fe -> replace with 'ves'
        echo "${word%fe}ves"
    elif [[ $word =~ [^aeiou]o$ ]]; then
        # words ending in consonant + o -> add 'es'
        echo "${word}es"
    else
        # default:  just add 's'
        echo "${word}s"
    fi
}

# Получить timestamp
get_timestamp() {
    date +%s
}

# Получить имя Go модуля из go.mod
get_module_name() {
    if [ -f "go.mod" ]; then
        grep "^module" go.mod | awk '{print $2}'
    else
        echo "github.com/user/project"
    fi
}

# Создать директорию если не существует
ensure_dir() {
    if [ ! -d "$1" ]; then
        mkdir -p "$1"
        echo "Created directory: $1"
    fi
}

# Создать файл из шаблона (совместимо с macOS и Linux)
# scripts/lib/helpers.sh - обновляем только функцию create_from_template

# Создать файл из шаблона (совместимо с macOS и Linux)
create_from_template() {
    local template=$1
    local output=$2
    local entity_name=$3
    local entity_lower=$4
    local entity_snake=$5
    local entity_plural=$6
    local module_name=$7
    
    if [ ! -f "$template" ]; then
        error "Template not found: $template"
        return 1
    fi
    
    # Экранируем спецсимволы в module_name
    local module_escaped=$(echo "$module_name" | sed 's/[\/&]/\\&/g')
    local temp_file="${output}.tmp"
    
    # Используем __VAR__ синтаксис
    cat "$template" | \
        sed "s/__ENTITY__/$entity_name/g" | \
        sed "s/__entity__/$entity_lower/g" | \
        sed "s/__entity_snake__/$entity_snake/g" | \
        sed "s/__entity_plural__/$entity_plural/g" | \
        sed "s|__MODULE__|$module_escaped|g" \
        > "$temp_file"
    
    mv "$temp_file" "$output"
    echo "Created:  $output"
}

# Добавить строку в файл после паттерна (совместимо с macOS)
add_line_after_pattern() {
    local file=$1
    local pattern=$2
    local line=$3
    
    # Проверяем существует ли файл
    if [ ! -f "$file" ]; then
        error "File not found: $file"
        return 1
    fi
    
    # Проверяем существует ли уже такая строка
    if grep -Fq "$line" "$file" 2>/dev/null; then
        return 0
    fi
    
    # Создаем временный файл
    local temp_file="${file}.tmp"
    
    # Используем более простой подход - ищем точное совпадение "var ("
    if [ "$pattern" = "var (" ]; then
        # Специальный случай для var (
        awk -v line="$line" '
            /^var \(/ {
                print
                if (! found) {
                    print line
                    found=1
                }
                next
            }
            {print}
        ' "$file" > "$temp_file"
    else
        # Общий случай
        awk -v pattern="$pattern" -v line="$line" '
            {print}
            index($0, pattern) > 0 && ! found {
                print line
                found=1
            }
        ' "$file" > "$temp_file"
    fi
    
    mv "$temp_file" "$file"
}

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN} $1${NC}"
}

error() {
    echo -e "${RED} $1${NC}"
}

warning() {
    echo -e "${YELLOW} $1${NC}"
}

info() {
    echo -e "${BLUE} $1${NC}"
}

debug() {
    if [ "${DEBUG}" = "true" ]; then
        echo -e "${MAGENTA} $1${NC}"
    fi
}
