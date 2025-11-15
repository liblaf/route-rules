import datetime
from pathlib import Path
from typing import Self

import pydantic

from route_rules.provider import Behavior, Format


class ArtifactMeta(pydantic.BaseModel):
    behavior: Behavior
    format: Format
    path: Path
    size: int


class ProviderMeta(pydantic.BaseModel):
    name: str
    download_url: str
    preview_url: str


class RecipeStatistics(pydantic.BaseModel):
    inputs: dict[str, int] = pydantic.Field(default_factory=dict)
    outputs: dict[str, int] = pydantic.Field(default_factory=dict)


class RecipeMeta(pydantic.BaseModel):
    name: str
    slug: str
    artifacts: list[ArtifactMeta] = pydantic.Field(default_factory=list)
    providers: list[ProviderMeta] = pydantic.Field(default_factory=list)
    statistics: RecipeStatistics = pydantic.Field(default_factory=RecipeStatistics)


class Meta(pydantic.BaseModel):
    build_time: datetime.datetime
    recipes: list[RecipeMeta] = pydantic.Field(default_factory=list)

    @classmethod
    def json_decode(cls, data: str | bytes | bytearray) -> Self:
        return cls.model_validate_json(data)

    def json_encode(self) -> str:
        return self.model_dump_json()
