import abc
import os

import attrs
from anyio import Path

from route_rules.core import RuleSet


@attrs.define
class Exporter(abc.ABC):
    @abc.abstractmethod
    async def export(
        self, folder: str | os.PathLike[str], slug: str, ruleset: RuleSet
    ) -> Path | None:
        raise NotImplementedError
