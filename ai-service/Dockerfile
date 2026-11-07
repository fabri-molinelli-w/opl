# OPSBot (FastAPI + LangChain + ChromaDB). Nunca se publica: solo lo llama el backend por la red interna.
FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 \
    CHROMA_PATH=/data/chroma FASTEMBED_CACHE_PATH=/data/fastembed
WORKDIR /app
COPY pyproject.toml ./
COPY app ./app
RUN pip install --no-cache-dir . \
 && useradd --system --uid 10001 opsbot \
 && mkdir -p /data && chown opsbot /data
USER opsbot
EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
