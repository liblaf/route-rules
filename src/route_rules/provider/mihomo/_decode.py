import logging
import re
from collections.abc import Iterable

import autoregistry
import msgspec

from route_rules.core import RuleSet

from ._enum import Behavior, Format

logger: logging.Logger = logging.getLogger(__name__)


def decode(data: str | bytes, behavior: Behavior, format: Format) -> RuleSet:  # noqa: A002
    payload: list[str] = _format_decoders[format](data)
    ruleset: RuleSet = _behavior_decoders[behavior](payload)
    return ruleset


_format_decoders = autoregistry.Registry(prefix="_decode_")


@_format_decoders
def _decode_yaml(data: str | bytes) -> list[str]:
    return msgspec.yaml.decode(data)["payload"]


@_format_decoders
def _decode_text(text: str | bytes) -> list[str]:
    if isinstance(text, bytes):
        text = text.decode()
    text = re.sub(r"#.*", "", text, flags=re.MULTILINE)
    lines: list[str] = text.splitlines()
    lines = [line.strip() for line in lines]
    lines = [line for line in lines if line]
    return lines


_behavior_decoders = autoregistry.Registry(prefix="_decode_")


@_behavior_decoders
def _decode_domain(payload: Iterable[str]) -> RuleSet:
    # ref: <https://wiki.metacubex.one/en/handbook/syntax/#domain-wildcards>
    ruleset = RuleSet()
    for line in payload:
        if line.startswith("*."):
            logger.warning("Unsupported: Domain Wildcard `*`.")
        elif line.startswith("+."):
            ruleset.domain_suffix.add(line[2:])
        elif line.startswith("."):
            logger.warning("Unsupported: Domain Wildcard `.`.")
        else:
            ruleset.domain.add(line)
    return ruleset


@_behavior_decoders
def _decode_classical(payload: Iterable[str]) -> RuleSet:
    ruleset = RuleSet()
    for line in payload:
        typ: str
        value: str
        typ, value, *_ = line.split(",", maxsplit=2)
        ruleset.add(typ, value)
    return ruleset
