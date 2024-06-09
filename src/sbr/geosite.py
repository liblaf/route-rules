import functools
import json
import pathlib
import subprocess
import tempfile
from concurrent.futures import ThreadPoolExecutor

import requests

from sbr.rule_set import RuleSet


class Geosite:
    file: pathlib.Path
    _dtemp: tempfile.TemporaryDirectory

    def __init__(
        self,
        file: pathlib.Path | None = None,
        url: str = "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.db",
    ) -> None:
        self._dtemp = tempfile.TemporaryDirectory()
        if file is None:
            self.file = self.dtemp / "geosite.db"
            with self.file.open("wb") as fp:
                resp: requests.Response = requests.get(url)
                resp.raise_for_status()
                fp.write(resp.content)
        else:
            self.file = file

    def export(self, *categories: str) -> RuleSet:
        with ThreadPoolExecutor() as executor:
            return sum(executor.map(self._export, categories), RuleSet())

    @functools.cache
    def list(self) -> list[str]:
        proc: subprocess.CompletedProcess[str] = subprocess.run(
            ["sing-box", "geosite", "--file", self.file, "list"],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            text=True,
            check=True,
        )
        categories: list[str] = []
        for line in proc.stdout.splitlines():
            categories.append(line.split()[0])
        return categories

    @functools.cached_property
    def dtemp(self) -> pathlib.Path:
        return pathlib.Path(self._dtemp.name)

    @functools.lru_cache()
    def _export(self, category: str) -> RuleSet:
        output: pathlib.Path = self.dtemp / f"geosite-{category}.json"
        subprocess.run(
            [
                "sing-box",
                "geosite",
                "--file",
                self.file,
                "export",
                category,
                "--output",
                output,
            ],
            stdin=subprocess.DEVNULL,
            check=True,
        )
        return RuleSet(**json.loads(output.read_text()))
