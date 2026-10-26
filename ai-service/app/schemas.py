from typing import Literal

from pydantic import BaseModel, ConfigDict, Field
from pydantic.alias_generators import to_camel

SourceType = Literal["document", "service", "server", "incident"]


class CamelModel(BaseModel):
    model_config = ConfigDict(alias_generator=to_camel, populate_by_name=True)


class ChatRequest(CamelModel):
    session_id: str = Field(min_length=1, max_length=100)
    message: str = Field(min_length=1, max_length=4000)


class Source(CamelModel):
    type: SourceType
    id: str
    title: str


class ChatResponse(CamelModel):
    answer: str
    sources: list[Source]
