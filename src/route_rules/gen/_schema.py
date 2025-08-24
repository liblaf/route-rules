import os
from pathlib import Path
from typing import Any, Self

import msgspec
import pydantic


class Target(pydantic.BaseModel):
    name: str
    providers: list[str]


class Config(pydantic.BaseModel):
    targets: dict[str, Target]

    @classmethod
    def load(cls, file: str | os.PathLike[str]) -> Self:
        file = Path(file)
        data: Any = msgspec.yaml.decode(file.read_bytes())
        return cls.model_validate(data)
