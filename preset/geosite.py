import argparse
import asyncio

from sbr import Rule

parser = argparse.ArgumentParser()
parser.add_argument("input", type=str)
parser.add_argument("output", type=str)


async def main() -> None:
    args: argparse.Namespace = parser.parse_args()
    rule: Rule = await Rule.from_json_url(args.input)
    rule.ip_cidr.clear()
    rule.save(args.output)


if __name__ == "__main__":
    asyncio.run(main())
