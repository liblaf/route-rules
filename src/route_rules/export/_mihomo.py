import os
import subprocess
import tempfile
from collections.abc import Generator, Iterable
from pathlib import Path
from typing import override

import attrs
import msgspec

from route_rules.core import RuleSet
from route_rules.provider.mihomo import Behavior, Format

from ._abc import Exporter


@attrs.define
class ExporterMihomo(Exporter):
    behavior: Behavior = attrs.field()
    format: Format = attrs.field()
    export_path_template: str = attrs.field(
        default="mihomo/{slug}.{behavior}{format.ext}", kw_only=True
    )

    @override
    def export(
        self,
        file: str | os.PathLike[str],
        ruleset: RuleSet,
    ) -> int:
        data: bytes = encode(ruleset, behavior=self.behavior, format=self.format)
        if not data:
            return 0
        file = Path(file)
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_bytes(data)
        return len(data)

    @override
    def export_path(self, slug: str) -> Path:
        return Path(
            self.export_path_template.format(
                slug=slug, behavior=self.behavior, format=self.format
            )
        )


@attrs.define
class EncodeError(RuntimeError):
    behavior: Behavior = attrs.field()
    format: Format = attrs.field()


def encode(ruleset: RuleSet, behavior: Behavior, format: Format) -> bytes:  # noqa: A002
    payload: Iterable[str]
    match behavior:
        case Behavior.DOMAIN:
            payload = _encode_domain(ruleset)
        case Behavior.IPCIDR:
            payload = _encode_ipcidr(ruleset)
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


def _encode_ipcidr(ruleset: RuleSet) -> set[str]:
    return ruleset.ip_cidr


def _encode_classical(ruleset: RuleSet) -> Generator[str]:
    for typ, values in ruleset.data.items():
        if typ in {"DOMAIN", "DOMAIN-SUFFIX", "IP-CIDR"}:
            continue
        for value in values:
            yield f"{typ},{value}"


def _encode_yaml(payload: Iterable[str]) -> bytes:
    payload = list(payload)
    if not payload:
        return b""
    return msgspec.yaml.encode({"payload": list(payload)})


def _encode_text(payload: Iterable[str]) -> str:
    payload = list(payload)
    if not payload:
        return ""
    return "\n".join(payload)


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
