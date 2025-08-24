import datetime
from collections.abc import Buffer
from pathlib import Path
from typing import Any, Self

import msgspec

from route_rules.provider import Behavior, Format


class ArtifactMeta(msgspec.Struct):
    behavior: Behavior
    format: Format
    path: Path
    size: int


class ProviderMeta(msgspec.Struct):
    name: str
    download_url: str
    preview_url: str


class RecipeStatistics(msgspec.Struct):
    inputs: dict[str, int] = msgspec.field(default_factory=dict)
    outputs: dict[str, int] = msgspec.field(default_factory=dict)


class RecipeMeta(msgspec.Struct):
    name: str
    slug: str
    artifacts: list[ArtifactMeta] = msgspec.field(default_factory=list)
    providers: list[ProviderMeta] = msgspec.field(default_factory=list)
    statistics: RecipeStatistics = msgspec.field(default_factory=RecipeStatistics)


class Meta(msgspec.Struct):
    build_time: datetime.datetime
    recipes: list[RecipeMeta] = msgspec.field(default_factory=list)

    @classmethod
    def json_decode(cls, data: Buffer | str) -> Self:
        return msgspec.json.decode(data, type=cls, dec_hook=dec_hook)

    def json_encode(self) -> bytes:
        return msgspec.json.encode(self, enc_hook=enc_hook)


def dec_hook(typ: type, obj: Any) -> Any:
    return typ(obj)


def enc_hook(obj: Any) -> Any:
    match obj:
        case Path():
            return obj.as_posix()
        case _:
            return obj
