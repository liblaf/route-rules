import abc
import os
from pathlib import Path

import attrs

from route_rules.core import RuleSet


@attrs.define
class Exporter(abc.ABC):
    @abc.abstractmethod
    def export(
        self, folder: str | os.PathLike[str], slug: str, ruleset: RuleSet
    ) -> Path | None:
        raise NotImplementedError
