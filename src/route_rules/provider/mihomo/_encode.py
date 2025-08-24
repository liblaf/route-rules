import subprocess
import tempfile
from collections.abc import Generator, Iterable
from pathlib import Path

import attrs
import msgspec

from route_rules.core import RuleSet

from ._enum import Behavior, Format


@attrs.define
class EncodeError(RuntimeError):
    behavior: Behavior = attrs.field()
    format: Format = attrs.field()


def encode(ruleset: RuleSet, behavior: Behavior, format: Format) -> bytes:  # noqa: A002
    payload: Iterable[str]
    match behavior:
        case Behavior.DOMAIN:
            payload = _encode_domain(ruleset)
        case Behavior.CLASSICAL:
            payload = _encode_classical(ruleset)
        case _:
            raise EncodeError(behavior=behavior, format=format)
    match format:
        case Format.YAML:
            return _encode_yaml(payload)
        case Format.TEXT:
            return _encode_text(payload).encode()
        case Format.MRS:
            return _encode_mrs(payload, behavior=behavior)
        case _:
            raise EncodeError(behavior=behavior, format=format)


def _encode_domain(ruleset: RuleSet) -> Generator[str]:
    yield from ruleset.domain
    for domain in ruleset.domain_suffix:
        yield f"+.{domain}"


def _encode_classical(ruleset: RuleSet) -> Generator[str]:
    for typ, value in ruleset.data.items():
        yield f"{typ},{value}"


def _encode_yaml(payload: Iterable[str]) -> bytes:
    return msgspec.yaml.encode({"payload": list(payload)})


def _encode_text(payload: Iterable[str]) -> str:
    return "\n".join(payload)


def _encode_mrs(payload: Iterable[str], behavior: Behavior) -> bytes:
    with tempfile.TemporaryDirectory() as tmpdir_str:
        tmpdir = Path(tmpdir_str)
        yaml_file: Path = tmpdir / "rule-set.yaml"
        mrs_file: Path = tmpdir / "rule-set.mrs"
        yaml_file.write_bytes(_encode_yaml(payload))
        subprocess.run(
            ["mihomo", "convert-ruleset", behavior, "yaml", yaml_file, mrs_file],
            check=True,
        )
        return mrs_file.read_bytes()
