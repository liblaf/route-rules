import asyncio
import asyncio.subprocess as asp
import functools
import pathlib
import shlex
import subprocess as sp
import tempfile

import cachetools

from sbr import utils
from sbr.rule import Rule
from sbr.typing import StrPath


class GeoIP:
    file: pathlib.Path
    _export_cache: cachetools.Cache[str, Rule]
    _dtemp: tempfile.TemporaryDirectory[str]

    @classmethod
    async def from_url(cls, url: StrPath) -> "GeoIP":
        geoip: GeoIP = cls()
        geoip.file = await utils.download_file(url, geoip.dtemp)
        return geoip

    def __init__(self) -> None:
        self._export_cache = cachetools.LRUCache(128)
        self._dtemp = tempfile.TemporaryDirectory()

    def __repr__(self) -> str:
        return self.summary

    @functools.cached_property
    def args(self) -> list[str]:
        return ["sing-box", "geoip", "--file", str(self.file)]

    @functools.cached_property
    def countries(self) -> list[str]:
        proc: sp.CompletedProcess[str] = sp.run(
            [*self.args, "list"],
            stdin=sp.DEVNULL,
            stdout=sp.PIPE,
            text=True,
            check=True,
        )
        countries: list[str] = proc.stdout.splitlines()
        return countries

    @functools.cached_property
    def dtemp(self) -> pathlib.Path:
        return pathlib.Path(self._dtemp.name)

    async def export(self, country: str) -> Rule:
        if country not in self._export_cache:
            output: pathlib.Path = self.dtemp / f"geoip-{country}.json"
            args: list[str] = [*self.args, "export", country, "--output", str(output)]
            proc: asp.Process = await asyncio.create_subprocess_exec(
                *args, stdin=asp.DEVNULL
            )
            retcode: int = await proc.wait()
            if retcode != 0:
                raise sp.CalledProcessError(retcode, shlex.join(args))
            self._export_cache[country] = await Rule.from_json_url(output)
        return self._export_cache[country]

    @functools.cached_property
    def summary(self) -> str:
        return "\n".join(self.countries)
