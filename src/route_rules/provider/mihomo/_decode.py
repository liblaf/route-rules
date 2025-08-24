import re
from collections.abc import Iterable

import attrs
import msgspec
from loguru import logger

from route_rules.core import RuleSet

from ._enum import Behavior, Format


@attrs.define
class DecodeError(RuntimeError):
    behavior: Behavior = attrs.field()
    format: Format = attrs.field()


def decode(data: str | bytes, behavior: Behavior, format: Format) -> RuleSet:  # noqa: A002
    payload: list[str]
    match format:
        case Format.YAML:
            payload = _decode_yaml(data)
        case Format.TEXT:
            payload = _decode_text(data)
        case _:
            raise DecodeError(behavior=behavior, format=format)
    match behavior:
        case Behavior.DOMAIN:
            return _decode_domain(payload)
        case Behavior.CLASSICAL:
            return _decode_classical(payload)
        case _:
            raise DecodeError(behavior=behavior, format=format)


def _decode_domain(payload: Iterable[str]) -> RuleSet:
    # ref: <https://wiki.metacubex.one/en/handbook/syntax/#domain-wildcards>
    ruleset = RuleSet()
    for line in payload:
        if line.startswith("*."):
            logger.warning("Unsupported: Domain Wildcard `*`.", once=True)
        elif line.startswith("+."):
            ruleset.domain_suffix.add(line[2:])
        elif line.startswith("."):
            logger.warning("Unsupported: Domain Wildcard `.`.", once=True)
        else:
            ruleset.domain.add(line)
    return ruleset


def _decode_classical(payload: Iterable[str]) -> RuleSet:
    ruleset = RuleSet()
    for line in payload:
        typ: str
        value: str
        typ, value, *_ = line.split(",", maxsplit=2)
        ruleset.add(typ, value)
    return ruleset


def _decode_yaml(data: str | bytes) -> list[str]:
    return msgspec.yaml.decode(data)["payload"]


def _decode_text(text: str | bytes) -> list[str]:
    if isinstance(text, bytes):
        text = text.decode()
    text = re.sub(r"#.*", "", text, flags=re.MULTILINE)
    lines: list[str] = text.splitlines()
    lines = [line.strip() for line in lines]
    lines = [line for line in lines if line]
    return lines
