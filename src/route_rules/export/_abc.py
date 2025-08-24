import abc
import os
from pathlib import Path

import attrs

from route_rules.core import RuleSet


@attrs.define
class Exporter(abc.ABC):
    export_path_template: str = attrs.field(kw_only=True)

    @abc.abstractmethod
    def export(self, file: str | os.PathLike[str], ruleset: RuleSet) -> int:
        raise NotImplementedError

    def export_path(self, slug: str) -> Path:
        return Path(self.export_path_template.format(slug=slug))
