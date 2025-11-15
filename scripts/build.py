import asyncio

from liblaf import grapes

import route_rules as rr


async def main() -> None:
    grapes.logging.init(filter={"httpx": "WARNING"})
    builder: rr.Builder = rr.Builder.load("gen/config.yaml")
    await builder.build()


if __name__ == "__main__":
    asyncio.run(main())
