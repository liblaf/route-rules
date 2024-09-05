import datetime
import functools
import os
import time
from pathlib import Path

import httpx
import humanize
import ubelt as ub
from icecream import ic
from rich.progress import (
    BarColumn,
    DownloadColumn,
    Progress,
    TaskID,
    TaskProgressColumn,
    TextColumn,
    TimeRemainingColumn,
    TransferSpeedColumn,
)

from sbr.typing import StrPath


@functools.cache
def _client() -> httpx.AsyncClient:
    return httpx.AsyncClient(follow_redirects=True)


@functools.cache
def _progress() -> Progress:
    return Progress(
        TextColumn("[progress.description]{task.description}"),
        "[",
        BarColumn(),
        "]",
        TaskProgressColumn(),
        "(",
        DownloadColumn(binary_units=True),
        ")",
        TimeRemainingColumn(),
        TransferSpeedColumn(),
        transient=True,
    )


async def _download(url: str, fpath: Path) -> None:
    ic(url)
    client: httpx.AsyncClient = _client()
    prog: Progress = _progress()
    task_id: TaskID = prog.add_task(fpath.name)
    start: float = time.perf_counter()
    async with client.stream("GET", url) as r:
        length: int = int(r.headers["Content-Length"])
        prog.reset(task_id, total=length)
        prog.start()
        with fpath.open("wb") as fp:
            length = 0
            async for chunk in r.aiter_bytes():
                bytes_written: int = fp.write(chunk)
                prog.advance(task_id, bytes_written)
                length += bytes_written
        prog.update(task_id, total=length, completed=length)
    end: float = time.perf_counter()
    prog.remove_task(task_id)
    size: str = humanize.naturalsize(length, binary=True)
    delta: datetime.timedelta = datetime.timedelta(seconds=end - start)
    minutes: int = delta.seconds // 60
    seconds: float = delta.seconds % 60 + delta.microseconds / 1000000
    delta_str: str = f"{minutes:02}:{seconds:08.5f}"
    speed: str = humanize.naturalsize(length / (end - start), binary=True)
    prog.console.log(f"Downloaded to '{fpath}'. {size} in {delta_str} ({speed}/s).")
    if len(prog.tasks) == 0:
        prog.stop()


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
