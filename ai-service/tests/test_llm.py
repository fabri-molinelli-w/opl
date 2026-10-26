from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_openai import ChatOpenAI

from app.agent.agent import build_llm
from app.config import get_settings


def test_gemini_uses_native_integration() -> None:
    # Gemini 3 exige devolver la thought_signature al usar tools; la integración nativa la conserva.
    settings = get_settings().model_copy(
        update={"llm_base_url": "https://generativelanguage.googleapis.com/v1beta/openai/", "llm_model": "gemini-x"}
    )
    assert isinstance(build_llm(settings), ChatGoogleGenerativeAI)


def test_other_providers_use_openai_api() -> None:
    settings = get_settings().model_copy(update={"llm_base_url": "https://api.groq.com/openai/v1"})
    assert isinstance(build_llm(settings), ChatOpenAI)


def test_provider_can_be_forced() -> None:
    settings = get_settings().model_copy(
        update={"llm_base_url": "https://generativelanguage.googleapis.com/v1beta/openai/", "llm_provider": "openai"}
    )
    assert isinstance(build_llm(settings), ChatOpenAI)
