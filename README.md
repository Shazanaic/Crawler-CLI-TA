# Crawler-CLI-TA

CLI-приложение на Go для асинхронного обхода веб-страниц и формирования дерева найденных ресурсов.

Приложение принимает список стартовых URL, рекурсивно обходит HTML-страницы в пределах соответствующих доменов и сохраняет результат в JSON формате.

## Возможности

- несколько стартовых URL;
- настройка максимальной глубины обхода;
- обход только в пределах домена соответствующего стартового URL;
- извлечение ссылок из HTML-страниц;
- извлечение содержимого тега `<title>`;
- до 10 одновременно работающих workers;
- использование goroutine и channel;
- timeout отдельного HTTP-запроса;
- общий timeout выполнения crawler'а;
- graceful shutdown по `Ctrl+C`;
- пропуск HTTP redirects;
- пропуск не-HTML ресурсов;
- пропуск недоступных страниц;
- защита от повторного посещения URL;
- предотвращение циклического обхода;
- логирование ошибок и HTTP-статусов;
- сохранение результата в JSON;
- безопасная работа с общими данными;
- unit-тесты основных компонентов;
- интеграционный тест всей цепочки crawler'а.

## Требования

Для запуска необходим Go версии 1.20 или выше.

## Сборка

Клонировать репозиторий и перейти в его директорию:

```bash
git clone <repository-url>
cd crawler-cli-ta
```

Установить зависимости:

```bash
go mod download
```

Собрать приложение:

```bash
go build -o crawler-cli-ta ./cmd/crawler
```

После сборки будет создан исполняемый файл `crawler-cli-ta`.

## Запуск

Пример запуска:

```bash
./crawler-cli \
  --urls="https://google.com,https://example.com" \
  --depth=3 \
  --workers=10 \
  --timeout=2m \
  --request-timeout=10s \
  --output=result.json \
  --log=crawler.log
```

На Windows:

```powershell
.\crawler-cli.exe `
  --urls="https://google.com,https://example.com" `
  --depth=3 `
  --workers=10 `
  --timeout=2m `
  --request-timeout=10s `
  --output=result.json `
  --log=crawler.log
```

Также приложение можно запустить без предварительной сборки:

```bash
go run ./cmd/crawler \
  --urls="https://example.com" \
  --depth=2 \
  --workers=5 \
  --timeout=30s \
  --request-timeout=5s \
  --output=result.json \
  --log=crawler.log
```

## Параметры запуска

| Параметр | Описание | Значение по умолчанию |
|---|---|---|
| `--urls` | Список стартовых URL через запятую | — |
| `--depth` | Максимальная глубина рекурсивного обхода | `1` |
| `--workers` | Количество одновременно работающих workers | `10` |
| `--timeout` | Общий timeout выполнения | `10s` |
| `--request-timeout` | Timeout отдельного HTTP-запроса | `5s` |
| `--output` | Путь к JSON-файлу с результатом | `output.txt` |
| `--log` | Путь к файлу логов | `logs.log` |

Количество workers ограничено значением `10`.

## Результат

Результат сохраняется в JSON-файл.

Пример:

```json
[
  {
    "resource": "https://google.com",
    "title": "Google",
    "links": [
      {
        "resource": "https://google.com/about",
        "title": "About Google",
        "links": []
      }
    ]
  }
]
```

Каждая страница представлена узлом дерева:

- `resource` — URL страницы;
- `title` — содержимое `<title>`;
- `links` — найденные и успешно обработанные дочерние страницы.

## Логирование

Логи записываются в отдельный файл, указанный параметром `--log`.

Логируются:

- HTTP-статусы;
- ошибки HTTP-запросов;
- ошибки обработки страниц;
- ошибки парсинга;
- другие ошибки выполнения crawler'а.

Пример:

```text
time=2026-09-27T15:07:36.247+03:00 level=INFO msg="http request" url=https://google.com status_code=301 status="301 Moved Permanently"
time=2026-09-27T15:07:36.322+03:00 level=ERROR msg="failed to fetch page" url=https://google.com error="redirection 301 Moved Permanently"
time=2026-09-27T15:09:01.223+03:00 level=INFO msg="http request" url=https://www.google.com/ status_code=200 status="200 OK"
```

Ошибка одного ресурса не останавливает весь обход.

## Архитектура

Проект разделён на несколько компонентов:

```text
crawler-cli-ta/
├── cmd/
│   └── crawler/
│       └── main.go
│
├── base/
│   ├── config/
│   │   └── config.go
│   │
│   ├── models/
│   │   ├── page.go
│   │   ├── task.go
│   │   └── result.go
│   │
│   ├── crawler/
│   │   ├── crawler.go
│   │   ├── scheduler.go
│   │   ├── worker.go
│   │   └── *_test.go
│   │
│   ├── fetcher/
│   │   ├── fetcher.go
│   │   ├── http_fetcher.go
│   │   └── *_test.go
│   │
│   ├── parser/
│   │   ├── parser.go
│   │   ├── html_parser.go
│   │   └── *_test.go
│   │
│   ├── output/
│   │   ├── writer.go
│   │   ├── json_writer.go
│   │   └── *_test.go
│   │
│   └── logging/
│       └── logger.go
│
├── go.mod
├── go.sum
└── README.md
```

### Компоненты

**Config**

Отвечает за работу параметров командной строки.

**Fetcher**

Выполняет HTTP-запросы и скачивает контент с сайта, контролирует timeout, redirects, HTTP-статус и тип содержимого.

**Parser**

Использует goquery. Разбирает HTML страницу и извлекает:

- содержимое `<title>`;
- ссылки `<a href="...">`.

**Worker**

Считывает из канала и обрабатывает отдельную задачу:

```text
URL
 ↓
Fetcher
 ↓
HTML
 ↓
Parser
 ↓
Result
```
Записывает результат в канал results.

**Scheduler**

Управляет очередью задач, worker пулом, обработкой посещённых URL и построением дерева страниц.

**Crawler**

Объединяет Scheduler, Worker, Fetcher и Parser в единый интерфейс.

**Output**

Сохраняет сформированное дерево страниц в JSON.

**Logging**

Настраивает запись логов в отдельный файл.

## Глубина обхода

Стартовая страница имеет глубину `0`.

Например, при:

```bash
--depth=2
```

будут обработаны:

```text
depth 0
└── start

depth 1
└── start/about

depth 2
└── start/about/team
```

Ссылки со страниц глубины `2` уже не добавляются в очередь дальнейшего обхода.

## Ограничение домена

Для каждого стартового URL запоминается его домен.

Например:

```text
https://example.com
```

может вести на:

```text
https://example.com/about
```

но ссылка:

```text
https://google.com
```

будет проигнорирована.

Относительные ссылки разрешаются относительно текущего URL.

## Повторные URL и циклы

Crawler хранит множество уже посещённых URL, поэтому циклы невозможны.

Нормализация URL не выполняется. При проверке повторов URL сравниваются в том виде, в котором они были получены.

## Graceful shutdown

При нажатии `Ctrl+C` приложение отменяет общий `context.Context`.

Workers прекращают получение новых задач, а текущие HTTP-запросы получают отменённый context.

## Таймауты

Используются два уровня timeout.

### Timeout отдельного запроса

Параметр:

```bash
--request-timeout=5s
```

ограничивает время одного HTTP-запроса.

### Общий timeout

Параметр:

```bash
--timeout=30s
```

ограничивает общее время работы crawler'а.

## Обработка ошибок

Ошибка отдельного ресурса не прекращает весь обход.

Например:

```text
/start
├── /about       → 200 OK
├── /missing     → 404
├── /redirect    → 301
└── /image.jpg   → non-HTML
```

Все ошибки и HTTP-статусы записываются в логи.

## Ограничения

НЕ поддерживаются:

- авторизация;
- `robots.txt`;
- повторные запросы после ошибки;
- JavaScript-страницы;
- infinite scroll;
- нормализация URL.

## Тестирование

Запустить все тесты:

```bash
go test ./...
```

Запустить тесты с подробным выводом:

```bash
go test ./... -v
```

В проекте присутствуют:

- unit-тесты Fetcher;
- unit-тесты Parser;
- unit-тесты Scheduler;
- unit-тесты JSON Writer;
- тесты построения вложенного дерева;
- интеграционный тест работы всех компонентов crawler.

Интеграционный тест использует локальный `httptest.Server` и проверяет взаимодействие основных компонентов:

```text
HTTP Server
    ↓
HTTPFetcher
    ↓
HTMLParser
    ↓
Worker
    ↓
Scheduler
    ↓
Crawler
    ↓
Result tree
```

## Стек технологий и инструменты

### Использованный стек

- **Go** — основной язык разработки.
- **`net/http`** — выполнение HTTP-запросов.
- **`goquery`** — парсинг HTML и извлечение `<title>` и ссылок.
- **`encoding/json`** — сериализация результатов в JSON.
- **`log/slog`** — логирование ошибок и HTTP-статусов.
- **`flag`** — обработка параметров командной строки.
- **`httptest`** — создание HTTP-сервера для интеграционного тестирования.
- **Go testing** — написание и запуск unit- и интеграционных тестов.

### AI Usage

В разработке использовалась LLM **ChatGPT — GPT-5.6 Luna (OpenAI)** при:

- проектировании архитектуры приложения;
- разборе и исправлении ошибок;
- разработке отдельных компонентов;
- написании и проверке тестов;
- подготовке документации;
- анализе соответствия проекта требованиям задания.
