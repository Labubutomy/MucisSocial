FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir poetry==1.8.3

COPY pyproject.toml ./
RUN poetry config virtualenvs.create false \
    && poetry install --no-dev --no-interaction --no-ansi

COPY app ./app

EXPOSE 8000

CMD ["poetry", "run", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]

