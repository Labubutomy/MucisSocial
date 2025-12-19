FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN pip install --no-cache-dir poetry==1.8.4

COPY pyproject.toml /app/
RUN poetry config virtualenvs.create false \
  && poetry install --no-root --no-interaction --no-ansi

COPY . /app

EXPOSE 8080

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]


