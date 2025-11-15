import asyncio
import logging

from liblaf import grapes

import route_rules as rr


async def main() -> None:
    grapes.logging.init()
    builder: rr.Builder = rr.Builder.load("gen/config.yaml")
    await builder.build()
    for name, logger in logging.root.manager.loggerDict.items():
        if isinstance(logger, logging.PlaceHolder):
            continue
        print(name, logger.handlers)
    print(logging.root.filters)


if __name__ == "__main__":
    asyncio.run(main())
