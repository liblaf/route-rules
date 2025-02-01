import asyncio
import asyncio.subprocess as asp
import re
import subprocess as sp
from pathlib import Path

import route_rules as rr
from route_rules import Rule, Source
from route_rules.typing import StrPath


class GeoSite(Source):
    name: str
    dpath: Path
    url: str

    def __init__(self, name: str, url: str, dpath: StrPath) -> None:
        super().__init__()
        self.name = name
        self.dpath = Path(dpath)
        self.url = url

    @property
    def fpath(self) -> Path:
        return self.dpath / "geosite.db"

    async def _get_nocache(self, key: str) -> Rule:
        await rr.utils.download(self.url, self.fpath)
        output: Path = self.dpath / f"{key}.json"
        args: list[StrPath] = [
            "sing-box",
            "geosite",
            "export",
            key,
            "--output",
            output,
            "--file",
            self.fpath,
        ]
        proc: asp.Process = await asyncio.create_subprocess_exec(
            *args, stdin=asp.DEVNULL
        )
        ret: int = await proc.wait()
        if ret != 0:
            raise sp.CalledProcessError(ret, args)
        return Rule.from_file(output)

    async def _keys_nocache(self) -> list[str]:
        await rr.utils.download(self.url, self.fpath)
        args: list[StrPath] = ["sing-box", "geosite", "list", "--file", self.fpath]
        proc: asp.Process = await asyncio.create_subprocess_exec(
            *args, stdin=asp.DEVNULL, stdout=asp.PIPE
        )
        stdout: bytes
        stdout, _ = await proc.communicate()
        ret: int = await proc.wait()
        if ret != 0:
            raise sp.CalledProcessError(ret, args)
        categories: dict[str, int] = {}
        for line in stdout.decode().splitlines():
            if m := re.match(r"(?P<name>.*) \((?P<count>\d+)\)", line):
                categories[m["name"]] = int(m["count"])
        return list(categories.keys())
