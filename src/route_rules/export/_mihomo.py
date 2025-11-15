import os
import subprocess
import tempfile
from collections.abc import Generator, Iterable
from pathlib import Path
from typing import override

import attrs
import autoregistry
import msgspec

from route_rules.core import RuleSet
from route_rules.provider.mihomo import Behavior, Format

from ._abc import Exporter


@attrs.define
class ExporterMihomo(Exporter):
    behavior: Behavior
    format: Format

    @override
    def export(
        self, folder: str | os.PathLike[str], slug: str, ruleset: RuleSet
    ) -> Path | None:
        data: bytes = encode(ruleset, behavior=self.behavior, format=self.format)
        if not data:
            return None
        folder = Path(folder)
        file: Path = folder / "mihomo" / f"{slug}.{self.behavior}{self.format.ext}"
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_bytes(data)
        return file


def encode(ruleset: RuleSet, behavior: Behavior, format: Format) -> bytes:  # noqa: A002
    payload: Iterable[str] = _behavior_encoders[behavior](ruleset)
    match format:
        case Format.MRS:
            return _format_encoders[format](payload, behavior)
        case _:
            return _format_encoders[format](payload)


_behavior_encoders = autoregistry.Registry(prefix="_encode_")


@_behavior_encoders
def _encode_domain(ruleset: RuleSet) -> Generator[str]:
    yield from ruleset.domain
    for domain in ruleset.domain_suffix:
        yield f"+.{domain}"


@_behavior_encoders
def _encode_ipcidr(ruleset: RuleSet) -> set[str]:
    return ruleset.ip_cidr


_CLASSICAL_TYPE_EXCLUDE: set[str] = {"DOMAIN", "DOMAIN-SUFFIX", "IP-CIDR"}


@_behavior_encoders
def _encode_classical(ruleset: RuleSet) -> Generator[str]:
    for typ, values in ruleset.items():
        if typ in _CLASSICAL_TYPE_EXCLUDE:
            continue
        for value in values:
            yield f"{typ},{value}"


_format_encoders = autoregistry.Registry(prefix="_encode_")


@_format_encoders
def _encode_yaml(payload: Iterable[str]) -> bytes:
    payload = list(payload)
    if not payload:
        return b""
    return msgspec.yaml.encode({"payload": list(payload)})


@_format_encoders
def _encode_text(payload: Iterable[str]) -> bytes:
    payload = list(payload)
    if not payload:
        return b""
    return "\n".join(payload).encode()


@_format_encoders
def _encode_mrs(payload: Iterable[str], behavior: Behavior) -> bytes:
    payload = list(payload)
    if not payload:
        return b""
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
