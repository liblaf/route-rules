import functools
import pathlib
import urllib.parse
from collections.abc import Generator

import httpx
import pydantic.alias_generators

from sbr.typing import StrPath


def as_list(x: str | list[str]) -> list[str]:
    if isinstance(x, str):
        return [x]
    return x


async def download_file(url: StrPath, directory: StrPath) -> pathlib.Path:
    match url := path_or_url(url):
        case pathlib.Path():
            return url
        case str():
            parse: urllib.parse.ParseResult = urllib.parse.urlparse(url)
            path = pathlib.Path(parse.path)
            filename: pathlib.Path = pathlib.Path(directory) / path.name
            client: httpx.AsyncClient = _client()
            resp: httpx.Response = await client.get(url)
            resp = resp.raise_for_status()
            filename.write_bytes(resp.content)
            return filename


def path_or_url(x: StrPath) -> pathlib.Path | str:
    if isinstance(x, str):
        parse: urllib.parse.ParseResult = urllib.parse.urlparse(x)
        if parse.scheme in ("http", "https"):
            return x
    return pathlib.Path(x)


def strip_comments(text: str) -> Generator[str, None, None]:
    for line in text.splitlines():
        stripped: str = line.partition("#")[0].strip()
        if stripped:
            yield stripped


async def text_from_url(url: StrPath) -> str:
    match url := path_or_url(url):
        case pathlib.Path():
            return url.read_text()
        case str():
            client: httpx.AsyncClient = _client()
            resp: httpx.Response = await client.get(url)
            resp = resp.raise_for_status()
            return resp.text


def to_kebab(s: str) -> str:
    s = pydantic.alias_generators.to_snake(s)
    return s.replace("_", "-")


@functools.cache
def _client() -> httpx.AsyncClient:
    return httpx.AsyncClient(follow_redirects=True)
