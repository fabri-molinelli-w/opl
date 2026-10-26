from functools import lru_cache
from typing import Literal

from pydantic import SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Toda la configuración de OPSBot sale de variables de entorno (nunca hardcodeada)."""

    model_config = SettingsConfigDict(env_file=None, extra="ignore")

    backend_url: str = "http://localhost:8080"
    ai_internal_token: SecretStr

    # auto: Gemini (generativelanguage.googleapis.com) usa la integración nativa; el resto, la API de OpenAI.
    llm_provider: Literal["auto", "openai", "google"] = "auto"
    llm_base_url: str
    llm_api_key: SecretStr
    llm_model: str
    llm_temperature: float = 0.1

    embeddings_model: str = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
    chroma_path: str = "chroma_data"
    rag_top_k: int = 4
    memory_turns: int = 10


@lru_cache
def get_settings() -> Settings:
    return Settings()  # type: ignore[call-arg]  # valores obligatorios vienen del entorno
