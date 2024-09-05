import inspect
import logging
import sys

import rich.traceback
from loguru import logger


class InterceptHandler(logging.Handler):
    def emit(self, record: logging.LogRecord) -> None:
        # Get corresponding Loguru level if it exists.
        level: str | int
        try:
            level = logger.level(record.levelname).name
        except ValueError:
            level = record.levelno

        # Find caller from where originated the logged message.
        frame, depth = inspect.currentframe(), 0
        while frame and (depth == 0 or frame.f_code.co_filename == logging.__file__):
            frame = frame.f_back
            depth += 1

        logger.opt(depth=depth, exception=record.exc_info).log(
            level, record.getMessage()
        )


def init(level: str | int = logging.NOTSET) -> None:
    logger.remove()
    logger.add(
        sys.stderr,
        level=level,
        filter={"httpcore": logging.INFO, "httpx": logging.INFO},
    )
    logging.basicConfig(handlers=[InterceptHandler()], level=0, force=True)
    rich.traceback.install(show_locals=True)
