from pathlib import Path
from typing import Literal

from pydantic import BaseModel

import route_rules as rr
from route_rules.typing import StrPath


class RuleSet(BaseModel):
    version: Literal[1, 2]
    rules: list[rr.Rule]

    @classmethod
    def from_file(cls, path: StrPath) -> "RuleSet":
        fpath: Path = Path(path)
        return RuleSet.model_validate_json(fpath.read_text())

    def save(self, path: StrPath) -> None:
        json: str = self.model_dump_json(exclude_defaults=True)
        fpath: Path = Path(path)
        fpath.parent.mkdir(parents=True, exist_ok=True)
        fpath.write_text(json)
