import asyncio
import logging

from liblaf import grapes

import route_rules as rr


async def main() -> None:
    grapes.logging.init()
    logging.getLogger("httpx").setLevel(logging.WARNING)
    builder: rr.Builder = rr.Builder.load("gen/config.yaml")
    await builder.build()


if __name__ == "__main__":
    asyncio.run(main())
