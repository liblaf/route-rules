import hishel
import httpx
from loguru import logger

storage = hishel.AsyncFileStorage(ttl=86400)  # seconds
client = hishel.AsyncCacheClient(follow_redirects=True, storage=storage)


async def download(url: str) -> httpx.Response:
    response: httpx.Response = await client.get(url)
    response = response.raise_for_status()
    if response.extensions["from_cache"]:
        logger.success("Cache Hit: {}", url)
    else:
        logger.info("Cache Miss: {}", url)
    return response
