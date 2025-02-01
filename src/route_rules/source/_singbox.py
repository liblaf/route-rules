from pathlib import Path
from string import Template

import route_rules as rr
from route_rules import Rule, Source
from route_rules.typing import StrPath


class SingBoxRuleSet(Source):
    name: str
    dpath: Path
    url: Template

    def __init__(self, name: str, url: str | Template, dpath: StrPath) -> None:
        super().__init__()
        self.name = name
        if isinstance(url, str):
            self.url = Template(url)
        else:
            self.url = url
        self.dpath = Path(dpath)

    async def _get_nocache(self, key: str) -> Rule:
        filepath: Path = await rr.utils.download(
            self.url.substitute({"key": key}), self.dpath / f"{key}.json"
        )
        return Rule.from_file(filepath)

    async def _keys_nocache(self) -> list[str]:
        raise NotImplementedError
