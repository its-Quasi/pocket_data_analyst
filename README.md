# Pocket Data Analyst

Agente autónomo para análisis de bases de datos con interfaz TUI basado en **Code-as-Action (CodeAct)**. Convierte preguntas en lenguaje natural en código Go ejecutable y utiliza mecanismos de autoreparación para resolver errores durante la ejecución.

Permite consultar bases de datos, generar gráficos y obtener insights de forma interactiva desde la terminal, reduciendo la distancia entre una pregunta de negocio y el análisis de datos.

## Requisitos

| Dependencia         | Verificación             |
| ------------------- | ------------------------ |
| Go 1.26+            | `go version`             |
| Docker + Compose v2 | `docker compose version` |
| Ollama              | `ollama --version`       |

---

## Acerca de Ollama

Este proyecto utiliza **Ollama** para exponer un modelo de lenguaje compatible con la API de OpenAI. Antes de ejecutar el proyecto es necesario instalar Ollama, iniciar sesión con una cuenta gratuita y levantar el servicio local.

```bash
# ===== Ollama =====

# Inicia el servicio local
ollama serve

# Requiere una cuenta gratuita en ollama.com
ollama signin

# Este paso es opcional.
# El comando "make setup" descargará automáticamente este modelo
# (asegúrese de haber iniciado sesión previamente).
ollama pull gpt-oss:20b-cloud
```

---

## Setup y ejecución

Una vez completada la configuración de Ollama y realizado el `signin`, puede preparar el entorno y ejecutar el proyecto mediante el `Makefile`.

```bash
# Genera el archivo .env con la configuración por defecto.
# Puede modificar estos valores si, por ejemplo, el puerto de MySQL
# entra en conflicto con otro servicio. Esta configuración será utilizada
# por el asistente (wizard) al crear una nueva sesión.
make env-init

# Levanta MySQL 8.4, carga el dataset employees y descarga el modelo de Ollama
make setup

# Ejecuta la aplicación
go run ./cmd/dbagent
```

---

## Code-as-Action (Cómo se utiliza en el proyecto)

El agente transforma consultas en lenguaje natural en código Go ejecutable para resolver tareas de análisis de datos. Genera consultas SQL, procesa los resultados y crea visualizaciones HTML utilizando `go-echarts`.

Flujo de ejecución:

* **Genera** código Go a partir de la intención del usuario.
* **Ejecuta** el código de forma aislada en `sandbox_area/temporal.go`.
* **Analiza y repara** automáticamente los errores encontrados (hasta 5 intentos).

Cuando detecta errores relacionados con `go-echarts`, utiliza la documentación incluida en el proyecto como fuente de contexto para mejorar la generación y autoreparación del código.

<p align="center">
  <img src="images/codeact_flow.png" width="700" alt="flujo de accion">
</p>

---

## Uso

Una vez iniciada la aplicación, las acciones principales son:

* `n` → Crear una nueva sesión mediante el asistente de conexión (wizard). Ingrese la configuración correspondiente a la base de datos de prueba utilizando los valores definidos en `.env`.
* `Enter` → Enviar una consulta.
* `↑` / `↓` → Desplazarse por el historial de la conversación.
* `Esc` → Volver a la lista de sesiones.
* Los gráficos generados se abrirán automáticamente en el navegador predeterminado.

---

## Ejemplos de consultas

La base de datos luce así:

<p align="center">
  <img src="images/db_diagram.png" width="700" alt="diagrama base de datos">
</p>

<<<<<<< Updated upstream
Consultas de ejemplo sobre el dataset employees (en caso de no revisar los registros de la DB):
- Obten los 10 cargos con mayor salario promedio entre 1986 y 1993. Para cada cargo, muestra la evolucion de su salario promedio año a año durante ese periodo, permitiendo analizar cómo cambiaron sus ingresos a lo largo del tiempo.
- Comparacion de salario promedio del cargo manager por género en el año 1992 representado en grafico de barras, usa las etiquetas 'Hombre' y 'Mujer'
- Distribución porcentual de empleados por departamento representalo en un gráfico de pie
=======
Consultas de ejemplo sobre el dataset `employees` (en caso de no revisar los registros de la DB):
>>>>>>> Stashed changes

* Obten los 10 cargos con mayor salario promedio entre 1986 y 1993. Para cada cargo, muestra la evolucion de su salario promedio año a año durante ese periodo, permitiendo analizar cómo cambiaron sus ingresos a lo largo del tiempo. Representalo en un grafico de lineas
* Comparacion de salario promedio del cargo manager por género en el año 1992 representado en grafico de barras, usa las etiquetas 'Hombre' y 'Mujer'
* Distribución porcentual de empleados por departamento representalo en un gráfico de pie

---

## Comentarios adicionales

1. La carpeta `internal/lib/go-echarts` contiene la documentación oficial de la librería y se incluye como fuente de contexto para el agente. Su propósito es facilitar la generación y corrección del código cuando se requieren funcionalidades específicas de visualización.

2. Los gráficos generados se almacenan en `sandbox_area/charts`. Si ocurre un error durante la ejecución, estos archivos pueden inspeccionarse directamente para facilitar el diagnóstico del problema.
