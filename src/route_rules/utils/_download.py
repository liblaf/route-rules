import datetime
import functools
import os
import time
from pathlib import Path

import httpx
import humanize
import ubelt as ub
from loguru import logger

from route_rules.typing import StrPath


@functools.cache
def _client() -> httpx.AsyncClient:
    return httpx.AsyncClient(follow_redirects=True)


async def _download(url: str, fpath: Path) -> None:
    client: httpx.AsyncClient = _client()
    start: float = time.perf_counter()
    async with client.stream("GET", url) as r:
        r.raise_for_status()
        with fpath.open("wb") as fp:
            length = 0
            async for chunk in r.aiter_bytes():
                bytes_written: int = fp.write(chunk)
                length += bytes_written
    end: float = time.perf_counter()
    size_pretty: str = humanize.naturalsize(length, binary=True)
    delta: datetime.timedelta = datetime.timedelta(seconds=end - start)
    minutes: int = delta.seconds // 60
    seconds: float = delta.seconds % 60 + delta.microseconds / 1000000
    delta_pretty: str = f"{minutes:02}:{seconds:08.5f}"
    speed_pretty: str = humanize.naturalsize(length / (end - start), binary=True)
    logger.info(
        "Downloaded to '{}'. {} in {} ({}/s).",
        fpath,
        size_pretty,
        delta_pretty,
        speed_pretty,
    )


async def download(
    url: str,
    _fpath: StrPath | None = None,
    *,
    redo: bool = False,
    verbose: bool | None = True,
    expires: str | int | datetime.datetime | datetime.timedelta | None = None,
) -> Path:
    if _fpath is None:
        _fpath = os.path.basename(url)  # noqa: PTH119
    fpath: Path = Path(_fpath)
    fname: str = fpath.name
    fpath.parent.mkdir(parents=True, exist_ok=True)
    stamp = ub.CacheStamp(
        fname + ".stamp",
        dpath=fpath.parent,
        product=fpath,
        verbose=verbose,
        expires=expires,
        ext=".json",
    )
    if redo or stamp.expired():
        await _download(url, fpath)
        stamp.renew()
    return fpath
