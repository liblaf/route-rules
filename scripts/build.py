import asyncio

from liblaf import grapes
from loguru import logger

import route_rules as rr


async def main() -> None:
    grapes.logging.init()
    logger.disable("httpx")
    builder: rr.Builder = rr.Builder.load("gen/config.yaml")
    await builder.build()


if __name__ == "__main__":
    asyncio.run(main())
