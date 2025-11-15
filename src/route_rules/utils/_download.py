import logging

import httpx
from hishel.httpx import AsyncCacheClient

client = AsyncCacheClient(follow_redirects=True)
logger: logging.Logger = logging.getLogger(__name__)


async def download(url: str) -> httpx.Response:
    response: httpx.Response = await client.get(url)
    response = response.raise_for_status()
    if response.extensions["hishel_from_cache"]:
        logger.info("Cache Hit: %s", url)
    else:
        logger.info("Cache Miss: %s", url)
    return response
