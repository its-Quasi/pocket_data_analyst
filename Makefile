DB_CONTAINER=mysql-employees
OLLAMA_MODEL=gpt-oss:20b-cloud

env-init:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env creado a partir de .env.example"; \
	else \
		echo ".env ya existe"; \
	fi

db-download:
	@if [ ! -d test_db ]; then \
		git clone https://github.com/datacharmer/test_db.git; \
	fi

db-up:
	docker compose up -d

db-wait:
	@until docker exec $(DB_CONTAINER) mysqladmin ping -uroot -proot --silent; do \
		echo "Esperando MySQL..."; \
		sleep 2; \
	done

db-seed:
	docker cp test_db/. $(DB_CONTAINER):/tmp/test_db
	docker exec $(DB_CONTAINER) bash -c \
	"cd /tmp/test_db && mysql -uroot -proot employees < employees.sql"

db-clean:
	rm -rf test_db

ollama-check:
	@command -v ollama >/dev/null 2>&1 || \
		(echo "❌ Ollama no está instalado. Instálalo primero: https://ollama.com/download"; exit 1)

ollama-serve:
	@pgrep -x ollama >/dev/null || (ollama serve >/dev/null 2>&1 &)
	@sleep 3

ollama-pull: ollama-check ollama-serve
	@echo "Descargando modelo $(OLLAMA_MODEL)..."
	ollama pull $(OLLAMA_MODEL)

setup: db-download db-up db-wait db-seed db-clean ollama-pull

run:
	go run ./cmd/dbagent