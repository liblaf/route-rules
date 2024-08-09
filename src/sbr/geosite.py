import asyncio
import asyncio.subprocess as asp
import functools
import pathlib
import re
import shlex
import subprocess as sp
import tempfile

import cachetools

from sbr import utils
from sbr.rule import Rule
from sbr.typing import StrPath


class GeoSite:
    file: pathlib.Path
    _dtemp: tempfile.TemporaryDirectory[str]
    _export_cache: cachetools.Cache[str, Rule]

    @classmethod
    async def from_url(cls, url: StrPath) -> "GeoSite":
        geosite: GeoSite = cls()
        geosite.file = await utils.download_file(url, geosite.dtemp)
        return geosite

    def __init__(self) -> None:
        self._export_cache = cachetools.LRUCache(128)
        self._dtemp = tempfile.TemporaryDirectory()

    def __repr__(self) -> str:
        return self.summary

    @functools.cached_property
    def args(self) -> list[str]:
        return ["sing-box", "geosite", "--file", str(self.file)]

    @functools.cached_property
    def categories(self) -> list[str]:
        return list(self._categories.keys())

    @functools.cached_property
    def dtemp(self) -> pathlib.Path:
        return pathlib.Path(self._dtemp.name)

    async def export(self, category: str) -> Rule:
        if category not in self._export_cache:
            output: pathlib.Path = self.dtemp / f"geosite-{category}.json"
            args: list[str] = [*self.args, "export", category, "--output", str(output)]
            proc: asp.Process = await asyncio.create_subprocess_exec(
                *args, stdin=asp.DEVNULL
            )
            retcode: int = await proc.wait()
            if retcode != 0:
                raise sp.CalledProcessError(retcode, shlex.join(args))
            self._export_cache[category] = await Rule.from_json_url(output)
        return self._export_cache[category]

    @functools.cached_property
    def summary(self) -> str:
        res: list[str] = []
        for category, size in self._categories.items():
            res.append(f"{category} ({size})")
        return "\n".join(res)

    @functools.cached_property
    def _categories(self) -> dict[str, int]:
        proc: sp.CompletedProcess[str] = sp.run(
            [*self.args, "list"],
            stdin=sp.DEVNULL,
            stdout=sp.PIPE,
            text=True,
            check=True,
        )
        categories: dict[str, int] = {}
        for line in proc.stdout.splitlines():
            match: re.Match[str] | None = re.fullmatch(
                r"(?P<name>.+) \((?P<size>\d+)\)", line
            )
            assert match
            categories[match["name"]] = int(match["size"])
        return categories
