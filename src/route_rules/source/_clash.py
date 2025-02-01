from pathlib import Path
from string import Template

import route_rules as rr
from route_rules import Rule, Source
from route_rules.typing import StrPath


class ClashClassicalText(Source):
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
            self.url.substitute({"key": key}), self.dpath / f"{key}.list"
        )
        return ClashClassicalText.from_file(filepath)

    async def _keys_nocache(self) -> list[str]:
        raise NotImplementedError

    @staticmethod
    def from_file(fpath: StrPath) -> Rule:
        text: str = Path(fpath).read_text()
        rule: Rule = Rule()
        for line in rr.utils.strip_comments(text):
            words: list[str] = rr.utils.split_strip(line)
            match words[0]:
                case "DOMAIN":
                    rule.domain.add(words[1])
                case "DOMAIN-SUFFIX":
                    rule.domain_suffix.add(words[1])
                case "DOMAIN-KEYWORD":
                    rule.domain_keyword.add(words[1])
                case "DOMAIN-REGEX":
                    rule.domain_regex.add(words[1])
                case "IP-CIDR" | "IP-CIDR6":
                    rule.ip_cidr.add(words[1])
                case "IP-ASN":
                    pass  # TODO
                case "PROCESS-NAME":
                    pass  # TODO
                case _:
                    msg: str = f"Unknown rule: {line}"
                    raise ValueError(msg)
        return rule
