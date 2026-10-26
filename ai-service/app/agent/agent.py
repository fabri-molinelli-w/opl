"""Armado y ejecución del agente LangChain de OPSBot."""

from typing import Any

from langchain.agents import create_agent
from langchain_core.language_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langchain_core.tools import BaseTool
from langchain_openai import ChatOpenAI

from app.agent.prompt import SYSTEM_PROMPT
from app.config import Settings

MAX_AGENT_STEPS = 12


def build_llm(settings: Settings) -> BaseChatModel:
    """LLM compatible con la API de OpenAI (Gemini, Groq, OpenCode Zen, Ollama...), configurado por env vars."""
    return ChatOpenAI(
        base_url=settings.llm_base_url,
        api_key=settings.llm_api_key,
        model=settings.llm_model,
        temperature=settings.llm_temperature,
        timeout=60,
        max_retries=1,
    )


async def run_agent(llm: BaseChatModel, tools: list[BaseTool], history: list[BaseMessage], question: str) -> str:
    agent = create_agent(llm, tools=tools, system_prompt=SYSTEM_PROMPT)
    # Any: borde con LangGraph, cuyos overloads de ainvoke no aceptan un dict literal tipado.
    state: Any = {"messages": [*history, HumanMessage(content=question)]}
    result = await agent.ainvoke(state, config={"recursion_limit": MAX_AGENT_STEPS})
    for message in reversed(result["messages"]):
        if isinstance(message, AIMessage) and not message.tool_calls:
            answer = message.text.strip()
            if answer:
                return answer
    return "No pude generar una respuesta. Probá reformular la pregunta."
