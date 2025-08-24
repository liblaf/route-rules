import enum
import os
import re
from collections.abc import Iterable
from pathlib import Path
from typing import Any, override

import attrs
import autoregistry
import httpx
import msgspec
from loguru import logger

from route_rules import utils
from route_rules.core import RuleSet

from ._abc import Provider, ProviderFactory

_behavior_parsers = autoregistry.Registry()
_format_parsers = autoregistry.Registry()


class Behavior(enum.StrEnum):
    DOMAIN = enum.auto()
    IPCIDR = enum.auto()
    CLASSICAL = enum.auto()

    def parse(self, payload: Iterable[str]) -> RuleSet:
        return _behavior_parsers[self](payload)


class Format(enum.StrEnum):
    YAML = enum.auto()
    TEXT = enum.auto()
    MRS = enum.auto()

    def parse(self, text: str) -> list[str]:
        return _format_parsers[self](text)


@_behavior_parsers(name=Behavior.DOMAIN)
def _parse_domain(payload: Iterable[str]) -> RuleSet:
    # ref: <https://wiki.metacubex.one/en/handbook/syntax/#domain-wildcards>
    rule_set = RuleSet()
    for line in payload:
        if line.startswith("*."):
            logger.warning("Unsupported Domain Wildcard *", once=True)
        elif line.startswith("+."):
            rule_set.domain_suffix.add(line[2:])
        elif line.startswith("."):
            logger.warning("Unsupported Domain Wildcard .", once=True)
        else:
            rule_set.domain.add(line)
    return rule_set


@_behavior_parsers(name=Behavior.CLASSICAL)
def _parse_classical(payload: Iterable[str]) -> RuleSet:
    rule_set = RuleSet()
    for line in payload:
        typ: str
        value: str
        typ, value = line.split(",", 1)
        rule_set.add(typ, value)
    return rule_set


@_format_parsers(name=Format.YAML)
def _parse_yaml(text: str) -> list[str]:
    data: Any = msgspec.yaml.decode(text)
    return data["payload"]


@_format_parsers(name=Format.TEXT)
def _parse_text(text: str) -> list[str]:
    text = re.sub(r"#.*$", "", text, flags=re.MULTILINE)
    return [stripped for line in text.splitlines() if (stripped := line.strip())]


@attrs.define
class ProviderMihomo(Provider):
    url: str = attrs.field()
    behavior: Behavior = attrs.field(kw_only=True)
    format: Format = attrs.field(default=Format.YAML, kw_only=True)

    @classmethod
    def save(
        cls,
        file: str | os.PathLike[str],
        rule_set: RuleSet,
        *,
        behavior: Behavior,
        format: Format = Format.YAML,  # noqa: A002
    ) -> None:
        if behavior != Behavior.DOMAIN:
            raise NotImplementedError
        if format != Format.YAML:
            raise NotImplementedError
        file = Path(file)
        payload: list[str] = []
        payload += rule_set.domain
        payload += [f"+.{suffix}" for suffix in rule_set.domain_suffix]
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_bytes(msgspec.yaml.encode({"payload": payload}))

    @override
    async def load(self) -> RuleSet:
        response: httpx.Response = await utils.download(self.url)
        lines: list[str] = self.format.parse(response.text)
        return self.behavior.parse(lines)


@attrs.define
class ProviderMihomoFactory(ProviderFactory):
    name: str = attrs.field()
    url: str = attrs.field()
    behavior: Behavior = attrs.field(kw_only=True)
    format: Format = attrs.field(default=Format.YAML, kw_only=True)

    @override
    def create(self, name: str, /) -> ProviderMihomo:
        return ProviderMihomo(
            url=self.url.format(name=name),
            behavior=self.behavior,
            format=self.format,
        )
