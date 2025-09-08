from typing import Self

import attrs

from route_rules._recipe import Recipe

from ._config import RecipeConfig


@attrs.define
class RecipeWrapper(Recipe):
    @classmethod
    def from_config(cls, config: RecipeConfig) -> Self:
        return cls(name=config.name, providers=config.providers)
