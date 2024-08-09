import argparse
import asyncio

from icecream import ic
from sbr import Rule

parser = argparse.ArgumentParser()
parser.add_argument("input")


async def main() -> None:
    args: argparse.Namespace = parser.parse_args()
    rule: Rule = await Rule.from_json_url(args.input)
    ic(rule)


if __name__ == "__main__":
    asyncio.run(main())
