import asyncio
import os
import subprocess
import tempfile
from typing import override

import attrs
import autoregistry
import msgspec
from anyio import Path

from route_rules.core import RuleSet
from route_rules.provider.mihomo import Behavior, Format

from ._abc import Exporter


@attrs.define
class ExporterMihomo(Exporter):
    behavior: Behavior
    format: Format

    @override
    async def export(
        self, folder: str | os.PathLike[str], slug: str, ruleset: RuleSet
    ) -> Path | None:
        data: bytes = await encode(ruleset, behavior=self.behavior, format=self.format)
        if not data:
            return None
        folder = Path(folder)
        file: Path = folder / "mihomo" / f"{slug}.{self.behavior}{self.format.ext}"
        await file.parent.mkdir(parents=True, exist_ok=True)
        await file.write_bytes(data)
        return file


async def encode(ruleset: RuleSet, behavior: Behavior, format: Format) -> bytes:  # noqa: A002
    payload: list[str] = await _behavior_encoders[behavior](ruleset)  # ty:ignore[invalid-argument-type]
    match format:
        case Format.MRS:
            return await _format_encoders[format](payload, behavior)  # ty:ignore[invalid-argument-type]
        case _:
            return await _format_encoders[format](payload)  # ty:ignore[invalid-argument-type]


_behavior_encoders = autoregistry.Registry(prefix="_encode_")


@_behavior_encoders
async def _encode_domain(ruleset: RuleSet) -> list[str]:
    result: list[str] = []
    result.extend(ruleset.domain)
    result.extend(f"+.{domain}" for domain in ruleset.domain_suffix)
    return result


@_behavior_encoders
async def _encode_ipcidr(ruleset: RuleSet) -> set[str]:
    return ruleset.ip_cidr


_CLASSICAL_TYPE_EXCLUDE: set[str] = {"DOMAIN", "DOMAIN-SUFFIX", "IP-CIDR"}


@_behavior_encoders
async def _encode_classical(ruleset: RuleSet) -> list[str]:
    result: list[str] = []
    for typ, values in ruleset.items():
        if typ in _CLASSICAL_TYPE_EXCLUDE:
            continue
        result.extend(f"{typ},{value}" for value in values)
    return result


_format_encoders = autoregistry.Registry(prefix="_encode_")


@_format_encoders
async def _encode_yaml(payload: list[str]) -> bytes:
    if not payload:
        return b""
    return msgspec.yaml.encode({"payload": payload})


@_format_encoders
async def _encode_text(payload: list[str]) -> bytes:
    if not payload:
        return b""
    return "\n".join(payload).encode()


@_format_encoders
async def _encode_mrs(payload: list[str], behavior: Behavior) -> bytes:
    if not payload:
        return b""
    with tempfile.TemporaryDirectory() as tmpdir_str:
        tmpdir = Path(tmpdir_str)
        yaml_file: Path = tmpdir / "rule-set.yaml"
        mrs_file: Path = tmpdir / "rule-set.mrs"
        await yaml_file.write_bytes(await _encode_yaml(payload))
        cmd: list[str | os.PathLike[str]] = [
            "mihomo",
            "convert-ruleset",
            behavior,
            "yaml",
            yaml_file,
            mrs_file,
        ]
        process: asyncio.subprocess.Process = await asyncio.create_subprocess_exec(*cmd)
        returncode: int = await process.wait()
        if returncode:
            raise subprocess.CalledProcessError(returncode, cmd)
        return await mrs_file.read_bytes()
