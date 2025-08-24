import os
from pathlib import Path
from typing import Any, Self

import msgspec
import pydantic


class RecipeConfig(pydantic.BaseModel):
    name: str
    providers: list[str]


class Config(pydantic.BaseModel):
    recipes: list[RecipeConfig]

    @classmethod
    def load(cls, file: str | os.PathLike[str]) -> Self:
        file = Path(file)
        data: Any = msgspec.yaml.decode(file.read_bytes())
        return cls.model_validate(data)
